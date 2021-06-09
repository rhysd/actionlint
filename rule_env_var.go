package actionlint

import "strings"

// RuleEnvVar is a rule checker to check environment variables setup.
type RuleEnvVar struct {
	RuleBase
}

// NewRuleEnvVar creates new RuleEnvVar instance.
func NewRuleEnvVar() *RuleEnvVar {
	return &RuleEnvVar{
		RuleBase: RuleBase{name: "env-var"},
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleEnvVar) VisitStep(n *Step) {
	rule.checkEnv(n.Env)
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleEnvVar) VisitJobPre(n *Job) {
	rule.checkEnv(n.Env)
	if n.Container != nil {
		rule.checkEnv(n.Container.Env)
	}
	for _, s := range n.Services {
		rule.checkEnv(s.Container.Env)
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleEnvVar) VisitWorkflowPre(n *Workflow) {
	rule.checkEnv(n.Env)
}

func (rule *RuleEnvVar) checkEnv(env Env) {
	vars, ok := env.(EnvVars)
	if !ok {
		return
	}
	for _, v := range vars {
		if strings.ContainsAny(v.Name.Value, "&=- 	") {
			rule.errorf(
				v.Name.Pos,
				"environment variable name %q is invalid. '&', '=', '-' and spaces should not be contained",
				v.Name.Value,
			)
		}
	}
}
