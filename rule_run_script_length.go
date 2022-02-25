package actionlint

import "strings"

// RuleRunScriptLength is a rule to check the 'run' field does not contain too long scripts.
type RuleRunScriptLength struct {
	RuleBase
	maxLines int
}

// NewRuleRunScriptLength creates new RuleRunScriptLength instance.
func NewRuleRunScriptLength(maxLines int) *RuleRunScriptLength {
	return &RuleRunScriptLength{
		RuleBase: RuleBase{name: "run-script-length"},
		maxLines: maxLines,
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleRunScriptLength) VisitStep(n *Step) error {
	if run, ok := n.Exec.(*ExecRun); ok {
		rule.checkRunScriptLength(run.Run)
	}
	return nil
}

func (rule *RuleRunScriptLength) checkRunScriptLength(node *String) {
	if node == nil {
		return
	}

	numberOfLines := strings.Count(node.Value, "\n")

	if numberOfLines <= rule.maxLines {
		return
	}

	rule.errorf(
		node.Pos,
		"run script has too many lines (%d), prefer outsourcing into a standalone script",
		numberOfLines,
	)
}
