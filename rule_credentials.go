package actionlint

import (
	"fmt"
	"strings"
)

// RuleCredentials is a rule to check credentials in workflows
type RuleCredentials struct {
	RuleBase
}

// NewRuleCredentials creates new RuleCredentials instance
func NewRuleCredentials() *RuleCredentials {
	return &RuleCredentials{}
}

func (rule *RuleCredentials) VisitJobPre(n *Job) {
	if n.Container != nil {
		rule.checkContainer("\"container\" section", n.Container)
	}
	for _, s := range n.Services {
		rule.checkContainer(fmt.Sprintf("%q service", s.Name.Value), s.Container)
	}
}

func (rule *RuleCredentials) checkContainer(where string, n *Container) {
	if n.Credentials == nil || n.Credentials.Password == nil {
		return
	}

	p := n.Credentials.Password
	s := strings.TrimSpace(p.Value)
	if !strings.HasPrefix(s, "${{") || !strings.HasSuffix(s, "}}") {
		rule.errorf(p.Pos, "\"password\" section in %s should be specified via secrets. do not put password value directly", where)
	}
}
