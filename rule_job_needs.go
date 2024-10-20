package actionlint

import (
	"strconv"
	"strings"
)

type nodeStatus int

const (
	nodeStatusNew nodeStatus = iota
	nodeStatusActive
	nodeStatusFinished
)

type jobNode struct {
	id       string
	needs    []string
	resolved []*jobNode
	status   nodeStatus
	pos      *Pos
}

type edge struct {
	from *jobNode
	to   *jobNode
}

// RuleJobNeeds is a rule to check 'needs' field in each job configuration. For more details, see
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idneeds
type RuleJobNeeds struct {
	RuleBase
	nodes map[string]*jobNode
}

// NewRuleJobNeeds creates new RuleJobNeeds instance.
func NewRuleJobNeeds() *RuleJobNeeds {
	return &RuleJobNeeds{
		RuleBase: RuleBase{
			name: "job-needs",
			desc: "Checks for job IDs in \"needs:\". Undefined IDs and cyclic dependencies are checked",
		},
		nodes: map[string]*jobNode{},
	}
}

func contains[T comparable](heystack []T, needle T) bool {
	for _, s := range heystack {
		if s == needle {
			return true
		}
	}
	return false
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleJobNeeds) VisitJobPre(n *Job) error {
	needs := make([]string, 0, len(n.Needs))
	for _, j := range n.Needs {
		id := strings.ToLower(j.Value)
		if contains(needs, id) {
			rule.Errorf(j.Pos, "job ID %q duplicates in \"needs\" section. note that job ID is case insensitive", j.Value)
			continue
		}
		if id != "" {
			// Job ID is key of mapping. Key mapping is stored in lowercase since it is case
			// insensitive. So values in 'needs' array must be compared in lowercase.
			needs = append(needs, id)
		}
	}

	id := strings.ToLower(n.ID.Value)
	if id == "" {
		return nil
	}
	if prev, ok := rule.nodes[id]; ok {
		rule.Errorf(n.Pos, "job ID %q duplicates. previously defined at %s. note that job ID is case insensitive", n.ID.Value, prev.pos.String())
	}

	rule.nodes[id] = &jobNode{
		id:     id,
		needs:  needs,
		status: nodeStatusNew,
		pos:    n.ID.Pos,
	}

	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (rule *RuleJobNeeds) VisitWorkflowPost(n *Workflow) error {
	// Resolve nodes
	valid := true
	for id, node := range rule.nodes {
		node.resolved = make([]*jobNode, 0, len(node.needs))
		for _, dep := range node.needs {
			n, ok := rule.nodes[dep]
			if !ok {
				rule.Errorf(node.pos, "job %q needs job %q which does not exist in this workflow", id, dep)
				valid = false
				continue
			}
			node.resolved = append(node.resolved, n)
		}
	}
	if !valid {
		return nil
	}

	// Note: Only the first cycle can be detected even if there are multiple cycles in "needs:" configurations.
	if edge := detectFirstCycle(rule.nodes); edge != nil {
		edges := map[*jobNode]*jobNode{}
		edges[edge.from] = edge.to
		collectCycle(edge.to, edges)

		// Start cycle from the smallest position to make the error message deterministic
		start := edge.from
		for n := range edges {
			if n.pos.IsBefore(start.pos) {
				start = n
			}
		}

		var msg strings.Builder
		msg.WriteString("cyclic dependencies in \"needs\" job configurations are detected. detected cycle is ")

		msg.WriteString(strconv.Quote(start.id))
		from, to := start, edges[start]
		for {
			msg.WriteString(" -> ")
			msg.WriteString(strconv.Quote(to.id))
			from, to = to, edges[to]
			if from == start {
				break
			}
		}

		rule.Error(start.pos, msg.String())
	}

	return nil
}

func collectCycle(src *jobNode, edges map[*jobNode]*jobNode) bool {
	for _, dest := range src.resolved {
		if dest.status != nodeStatusActive {
			continue
		}
		edges[src] = dest
		if _, ok := edges[dest]; ok {
			return true
		}
		if collectCycle(dest, edges) {
			return true
		}
		delete(edges, src)
	}
	return false
}

// Detect cyclic dependencies
// https://inzkyk.xyz/algorithms/depth_first_search/detecting_cycles/

func detectFirstCycle(nodes map[string]*jobNode) *edge {
	for _, v := range nodes {
		if v.status == nodeStatusNew {
			if e := detectCyclicNode(v); e != nil {
				return e
			}
		}
	}
	return nil
}

func detectCyclicNode(v *jobNode) *edge {
	v.status = nodeStatusActive
	for _, w := range v.resolved {
		switch w.status {
		case nodeStatusActive:
			return &edge{v, w}
		case nodeStatusNew:
			if e := detectCyclicNode(w); e != nil {
				return e
			}
		}
	}
	v.status = nodeStatusFinished
	return nil
}
