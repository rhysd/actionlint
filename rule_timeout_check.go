package actionlint

type RuleTimeoutCheck struct {
	RuleBase
}

func NewRuleTimeoutCheck() *RuleTimeoutCheck {
	return &RuleTimeoutCheck{
		RuleBase: RuleBase{
			name: "timeout-check",
			desc: "Checks that timeout-minutes is set per job, with optional max limit",
		},
	}
}

func (rule *RuleTimeoutCheck) VisitJobPre(n *Job) error {
	if rule.config == nil || !rule.Config().TimeoutMinutes.Required {
		// No need to check anything
		return nil
	}

	if n.Steps == nil {
		// This must be using a reusable workflow which does not support timeout-minutes
		return nil
	}

	if n.TimeoutMinutes == nil {
		rule.Error(
			n.Pos,
			"You must have a timeout-minutes set to avoid overspend.",
		)
	}

	if n.TimeoutMinutes != nil &&
		rule.config.TimeoutMinutes.MaxMinutes != 0 &&
		n.TimeoutMinutes.Value > rule.config.TimeoutMinutes.MaxMinutes {
		rule.Errorf(
			n.Pos,
			"Your timeout-minutes is greater than %d minutes.",
			int(rule.config.TimeoutMinutes.MaxMinutes),
		)
	}
	return nil
}
