package actionlint

import (
	"path/filepath"
	"strings"
	"sync"
)

type RuleWorkflowRunSharedState struct {
	refs  map[string]*workflowRef
	mutex sync.Mutex
}

type RuleWorkflowRun struct {
	RuleBase
	*RuleWorkflowRunSharedState
}

type workflowRef struct {
	seen  bool
	refAt []*workflowRefBy
}

type workflowRefBy struct {
	path string
	*Pos
}

func NewRuleWorkflowRunSharedState() *RuleWorkflowRunSharedState {
	return &RuleWorkflowRunSharedState{
		refs:  make(map[string]*workflowRef),
		mutex: sync.Mutex{},
	}
}

func NewRuleWorkflowRun(state *RuleWorkflowRunSharedState) *RuleWorkflowRun {
	return &RuleWorkflowRun{
		RuleBase: RuleBase{
			name: "workflow_run",
			desc: "Checks referenced workflows in `workflow_run` exists",
		},
		RuleWorkflowRunSharedState: state,
	}
}

func (state *RuleWorkflowRunSharedState) resolve(workflow string) *workflowRef {
	ref := state.refs[workflow]
	if ref == nil {
		ref = &workflowRef{refAt: make([]*workflowRefBy, 0)}
		state.refs[workflow] = ref
	}
	return ref
}

func (w *Workflow) computeName() string {
	if w.Name != nil {
		return w.Name.Value
	}
	return strings.TrimSuffix(filepath.Base(w.Path), filepath.Ext(w.Path))
}

func (rule *RuleWorkflowRun) VisitWorkflowPre(node *Workflow) error {
	rule.workflowSeen(node)

	for _, event := range node.On {
		switch e := event.(type) {
		case *WebhookEvent:
			for _, ref := range e.Workflows {
				rule.workflowReferenced(node.Path, ref)
			}
		}
	}
	return nil
}

func (rule *RuleWorkflowRun) workflowSeen(node *Workflow) {
	name := node.computeName()
	rule.Debug("workflowSeen: %q\n", name)

	rule.mutex.Lock()
	defer rule.mutex.Unlock()

	rule.resolve(name).seen = true
}

func (rule *RuleWorkflowRun) workflowReferenced(path string, refAt *String) {
	rule.Debug("workflowReferenced: %q\n", refAt.Value)

	rule.mutex.Lock()
	defer rule.mutex.Unlock()

	ref := rule.resolve(refAt.Value)
	ref.refAt = append(ref.refAt, &workflowRefBy{path, refAt.Pos})
}

func (state *RuleWorkflowRunSharedState) ComputeMissingReferences() map[string][]*Error {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	errs := make(map[string][]*Error)
	for workflow, refs := range state.refs {
		if !refs.seen {
			for _, refAt := range refs.refAt {
				errs[refAt.path] = append(
					errs[refAt.path],
					errorfAt(refAt.Pos, "workflow_run", "Workflow %q is not defined", workflow),
				)
			}
		}
	}
	return errs
}
