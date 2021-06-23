package actionlint

import "strings"

// RuleStepID is a rule to check step IDs in workflow.
type RuleStepID struct {
	RuleBase
	seen map[string]*Pos
}

// NewRuleStepID creates a new RuleStepID instance.
func NewRuleStepID() *RuleStepID {
	return &RuleStepID{
		RuleBase: RuleBase{name: "step-id"},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleStepID) VisitJobPre(n *Job) error {
	rule.seen = map[string]*Pos{}
	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleStepID) VisitJobPost(n *Job) error {
	rule.seen = nil
	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleStepID) VisitStep(n *Step) error {
	if n.ID == nil {
		return nil
	}

	id := strings.ToLower(n.ID.Value)
	if prev, ok := rule.seen[id]; ok {
		rule.errorf(n.ID.Pos, "step ID %q duplicates. previously defined at %s. step ID must be unique within a job. note that step ID is case insensitive", prev.String(), n.ID.Value)
		return nil
	}
	rule.seen[id] = n.ID.Pos
	return nil
}
