package actionlint

import (
	"regexp"
	"strings"
)

var jobIDPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// RuleID is a rule to check step IDs in workflow.
type RuleID struct {
	RuleBase
	seen map[string]*Pos
}

// NewRuleID creates a new RuleID instance.
func NewRuleID() *RuleID {
	return &RuleID{
		RuleBase: RuleBase{
			name: "id",
			desc: "Checks for duplication and naming convention of job/step IDs",
		},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleID) VisitJobPre(n *Job) error {
	rule.seen = map[string]*Pos{}

	rule.validateConvention(n.ID, "job")
	for _, j := range n.Needs {
		rule.validateConvention(j, "job")
	}

	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleID) VisitJobPost(n *Job) error {
	rule.seen = nil
	return nil
}

// VisitActionPre is callback when visiting Job node before visiting its children.
func (rule *RuleID) VisitActionPre(n *Action) error {
	rule.seen = map[string]*Pos{}

	return nil
}

// VisitActionPost is callback when visiting Job node after visiting its children.
func (rule *RuleID) VisitActionPost(n *Action) error {
	rule.seen = nil
	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleID) VisitStep(n *Step) error {
	if n.ID == nil {
		return nil
	}

	rule.validateConvention(n.ID, "step")

	id := strings.ToLower(n.ID.Value)
	if prev, ok := rule.seen[id]; ok {
		rule.Errorf(n.ID.Pos, "step ID %q duplicates. previously defined at %s. step ID must be unique within a job. note that step ID is case insensitive", n.ID.Value, prev.String())
		return nil
	}
	rule.seen[id] = n.ID.Pos
	return nil
}

func (rule *RuleID) validateConvention(id *String, what string) {
	if id == nil || id.Value == "" || id.ContainsExpression() || jobIDPattern.MatchString(id.Value) {
		return
	}
	rule.Errorf(id.Pos, "invalid %s ID %q. %s ID must start with a letter or _ and contain only alphanumeric characters, -, or _", what, id.Value, what)
}
