package actionlint

import (
	"fmt"
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

// RuleJobNeeds is a rule to check 'needs' field in each job conifiguration. For more details, see
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idneeds
type RuleJobNeeds struct {
	RuleBase
	nodes map[string]*jobNode
}

// NewRuleJobNeeds creates new RuleJobNeeds instance.
func NewRuleJobNeeds() *RuleJobNeeds {
	return &RuleJobNeeds{
		RuleBase: RuleBase{name: "job-needs"},
		nodes:    map[string]*jobNode{},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleJobNeeds) VisitJobPre(n *Job) {
	needs := make([]string, 0, len(n.Needs))
	for _, j := range n.Needs {
		needs = append(needs, j.Value)
	}

	rule.nodes[n.ID.Value] = &jobNode{
		id:     n.ID.Value,
		needs:  needs,
		status: nodeStatusNew,
		pos:    n.ID.Pos,
	}
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (rule *RuleJobNeeds) VisitWorkflowPost(n *Workflow) {
	// Resolve nodes
	valid := true
	for id, node := range rule.nodes {
		node.resolved = make([]*jobNode, 0, len(node.needs))
		for _, dep := range node.needs {
			n, ok := rule.nodes[dep]
			if !ok {
				rule.errorf(n.pos, "job %q needs job %q which does not exist in this workflow", id, dep)
				valid = false
				continue
			}
			node.resolved = append(node.resolved, n)
		}
	}
	if !valid {
		return
	}

	if edge := detectCyclic(rule.nodes); edge != nil {
		edges := map[string]string{}
		edges[edge.from.id] = edge.to.id
		collectCyclic(edge.to, edges)

		desc := make([]string, 0, len(edges))
		for from, to := range edges {
			desc = append(desc, fmt.Sprintf("%q -> %q", from, to))
		}

		rule.errorf(
			edge.from.pos,
			"cyclic dependencies in \"needs\" configurations of jobs are detected. detected cycle is %s",
			strings.Join(desc, ", "),
		)
	}
}

func collectCyclic(src *jobNode, edges map[string]string) bool {
	for _, dest := range src.resolved {
		if dest.status != nodeStatusActive {
			continue
		}
		edges[src.id] = dest.id
		if _, ok := edges[dest.id]; ok {
			return true
		}
		if collectCyclic(dest, edges) {
			return true
		}
		delete(edges, src.id)
	}
	return false
}

// Detect cyclic dependencies
// https://inzkyk.xyz/algorithms/depth_first_search/detecting_cycles/

func detectCyclic(nodes map[string]*jobNode) *edge {
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
