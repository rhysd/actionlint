package actionlint

import (
	"strings"
)

// RuleIfCond is a rule to check if: conditions.
type RuleIfCond struct {
	RuleBase
}

// NewRuleIfCond creates new RuleIfCond instance.
func NewRuleIfCond() *RuleIfCond {
	return &RuleIfCond{
		RuleBase: RuleBase{
			name: "if-cond",
			desc: "Checks for if: conditions which are always true/false",
		},
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleIfCond) VisitStep(n *Step) error {
	rule.checkIfCond(n.If)
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleIfCond) VisitJobPre(n *Job) error {
	rule.checkIfCond(n.If)
	if n.Snapshot != nil {
		rule.checkIfCond(n.Snapshot.If)
	}
	return nil
}

func (rule *RuleIfCond) checkIfCond(n *String) {
	if n == nil {
		return
	}
	if n.ContainsExpression() {
		rule.checkIfCondExpression(n)
	} else {
		rule.checkTrueFalseConstants(n)
	}
}

func (rule *RuleIfCond) checkIfCondExpression(n *String) {
	// Check number of ${{ }} for conditions like `${{ false }} || ${{ true }}` which are always evaluated to true
	if strings.HasPrefix(n.Value, "${{") && strings.HasSuffix(n.Value, "}}") && strings.Count(n.Value, "${{") == 1 {
		return
	}
	rule.Errorf(
		n.Pos,
		"if: condition %q is always evaluated to true because extra characters are around ${{ }}",
		n.Value,
	)
}

func (rule *RuleIfCond) checkTrueFalseConstants(n *String) {
	v := strings.TrimSpace(n.Value)
	if v == "true" || v == "false" {
		rule.Errorf(
			n.Pos,
			"condition %q is always evaluated to %s. remove the if: section",
			n.Value,
			v,
		)
	}
}
