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
	s, e := strings.Index(n.Value, "${{"), strings.Index(n.Value, "}}")
	if s >= 0 && e >= 0 {
		rule.checkPlaceholder(n, s, e)
	} else {
		rule.checkExpression(n.Pos, n.Value)
	}
}

func (rule *RuleIfCond) checkPlaceholder(n *String, start, end int) {
	// Check number of ${{ }} for conditions like `${{ false }} || ${{ true }}` which are always evaluated to true
	if start > 0 || end+len("}}") < len(n.Value) || strings.Count(n.Value, "${{") > 1 {
		rule.Errorf(
			n.Pos,
			"if: condition %q is always evaluated to true because extra characters are around ${{ }}",
			n.Value,
		)
		return
	}
	rule.checkExpression(n.Pos, n.Value[start+len("${{"):end])
}

func (rule *RuleIfCond) checkExpression(pos *Pos, input string) {
	i := strings.TrimSpace(input)
	l := NewExprLexer(i + "}}")
	if e, err := NewExprParser().Parse(l); err == nil {
		if NewExprSemanticsChecker(false, nil).IsConstant(e) {
			rule.Errorf(pos, "constant expression %q in condition. remove the if: section", i)
		}
	}
}
