package actionlint

import (
	"fmt"
)

// RuleCredentials is a rule to check credentials in workflows
type RuleCredentials struct {
	RuleBase
}

// NewRuleCredentials creates new RuleCredentials instance
func NewRuleCredentials() *RuleCredentials {
	return &RuleCredentials{
		RuleBase: RuleBase{
			name: "credentials",
			desc: "Checks for credentials in \"services:\" configuration",
		},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleCredentials) VisitJobPre(n *Job) error {
	if n.Container != nil {
		rule.checkContainer("\"container\" section", n.Container)
	}
	if n.Services != nil {
		for _, s := range n.Services.Value {
			rule.checkContainer(fmt.Sprintf("%q service", s.Name.Value), s.Container)
		}
	}
	return nil
}

func (rule *RuleCredentials) checkContainer(where string, n *Container) {
	if n.Credentials == nil || n.Credentials.Password == nil {
		return
	}

	p := n.Credentials.Password
	if !p.IsExpressionAssigned() {
		rule.Errorf(p.Pos, "\"password\" section in %s should be specified via secrets. do not put password value directly", where)
	}
}
