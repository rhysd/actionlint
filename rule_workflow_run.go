package actionlint

import (
	"path/filepath"
	"strings"
	"sync"
)

type RuleWorkflowRun struct {
	RuleBase
	refs  map[string]*workflowRef
	mutex sync.Mutex
}

type workflowRef struct {
	seen  bool
	refAt []*workflowRefBy
}

type workflowRefBy struct {
	path string
	*Pos
}

func NewRuleWorkflowRun() *RuleWorkflowRun {
	return &RuleWorkflowRun{
		RuleBase: RuleBase{
			name: "workflow_run",
			desc: "Checks referenced workflows in `workflow_run` exists",
		},
		refs:  make(map[string]*workflowRef),
		mutex: sync.Mutex{},
	}
}

func (rule *RuleWorkflowRun) resolve(workflow string) *workflowRef {
	rule.mutex.Lock()
	defer rule.mutex.Unlock()

	ref := rule.refs[workflow]
	if ref == nil {
		ref = &workflowRef{refAt: make([]*workflowRefBy, 0)}
		rule.refs[workflow] = ref
	}
	return ref
}

func (w *Workflow) computeName(path string) string {
	if w.Name != nil {
		return w.Name.Value
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func (rule *RuleWorkflowRun) VisitWorkflowPre(path string, node *Workflow) error {
	rule.workflowSeen(path, node)

	for _, event := range node.On {
		switch e := event.(type) {
		case *WebhookEvent:
			for _, ref := range e.Workflows {
				rule.workflowReferenced(path, ref)
			}
		}
	}
	return nil
}

func (rule *RuleWorkflowRun) workflowSeen(path string, node *Workflow) {
	name := node.computeName(path)
	rule.Debug("workflowSeen: %q\n", name)
	rule.resolve(name).seen = true
}

func (rule *RuleWorkflowRun) workflowReferenced(path string, refAt *String) {
	rule.Debug("workflowReferenced: %q\n", refAt.Value)
	ref := rule.resolve(refAt.Value)
	ref.refAt = append(ref.refAt, &workflowRefBy{path, refAt.Pos})
}

func (rule *RuleWorkflowRun) ComputeMissingReferences() map[string][]*Error {
	rule.mutex.Lock()
	defer rule.mutex.Unlock()

	errs := make(map[string][]*Error)
	for workflow, refs := range rule.refs {
		if !refs.seen {
			for _, refAt := range refs.refAt {
				errs[refAt.path] = append(
					errs[refAt.path],
					errorfAt(refAt.Pos, rule.name, "Workflow %q is not defined", workflow),
				)
			}
		}
	}
	rule.Debug("ComputeMissingReferences: %q\n", errs)
	return errs
}
