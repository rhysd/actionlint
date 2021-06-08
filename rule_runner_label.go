package actionlint

import (
	"strconv"
	"strings"
)

var allGitHubHostedRunnerLabels = []string{
	"windows-latest",
	"windows-2019",
	"windows-2016",
	"ubuntu-latest",
	"ubuntu-20.04",
	"ubuntu-18.04",
	"ubuntu-16.04",
	"macos-latest",
	"macos-11",
	"macos-11.0",
	"macos-10.15",
}

// RuleRunnerLabel is a rule to check runner label like "ubuntu-latest". There are two types of
// runners, GitHub-hosted runner and Self-hosted runner. GitHub-hosted runner is described at
// https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners .
// And Self-hosted runner is described at
// https://docs.github.com/en/actions/hosting-your-own-runners/using-self-hosted-runners-in-a-workflow .
type RuleRunnerLabel struct {
	RuleBase
	knownLabels []string
}

// NewRuleRunnerLabel creates new RuleRunnerLabel instance.
func NewRuleRunnerLabel(labels []string) *RuleRunnerLabel {
	return &RuleRunnerLabel{
		RuleBase:    RuleBase{name: "runner-label"},
		knownLabels: labels,
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleRunnerLabel) VisitJobPre(n *Job) {
	if n.RunsOn == nil {
		return
	}

	var m *Matrix
	if n.Strategy != nil {
		m = n.Strategy.Matrix
	}

	switch r := n.RunsOn.(type) {
	case *GitHubHostedRunner:
		rule.checkGitHubHostedRunner(r, m)
	case *SelfHostedRunner:
		// Nothing to check. Since any custom labels can be specified, there is no way to verify
		// labels.
	}
}

// https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners
func (rule *RuleRunnerLabel) checkGitHubHostedRunner(r *GitHubHostedRunner, m *Matrix) {
	l := r.Label.Value
	if strings.Contains(l, "${{") && strings.Contains(l, "}}") {
		for _, l := range rule.tryToGetLabelsInMatrix(l, m) {
			rule.verifyGitHubHotedRunnerLabel(l)
		}
		return
	}

	rule.verifyGitHubHotedRunnerLabel(r.Label)
}

func (rule *RuleRunnerLabel) verifyGitHubHotedRunnerLabel(label *String) {
	l := strings.TrimSpace(label.Value)
	for _, p := range allGitHubHostedRunnerLabels {
		// use EqualFold for ignoring case e.g. both macos-latest and macOS-latest should be accepted
		if strings.EqualFold(l, p) {
			return // ok
		}
	}
	for _, k := range rule.knownLabels {
		if strings.EqualFold(l, k) {
			return // ok
		}
	}

	qs := make([]string, 0, len(allGitHubHostedRunnerLabels))
	for _, p := range allGitHubHostedRunnerLabels {
		qs = append(qs, strconv.Quote(p))
	}
	rule.errorf(
		label.Pos,
		"label %q is unknown. available labels are %s. if the label is for self-hosted runner, set list of labels in actionlint.yaml config file",
		label.Value,
		strings.Join(qs, ", "),
	)
}

func (rule *RuleRunnerLabel) tryToGetLabelsInMatrix(l string, m *Matrix) []*String {
	if m == nil {
		return nil
	}
	l = strings.TrimSpace(l)

	// Only when the form of "${{...}}", evaluate the expression
	if strings.Count(l, "${{") != 1 || strings.Count(l, "}}") != 1 || !strings.HasPrefix(l, "${{") || !strings.HasSuffix(l, "}}") {
		return nil
	}

	lex := NewExprLexer()
	tok, _, err := lex.Lex(l[3:]) // 3 means omit first "${{"
	if err != nil {
		return nil
	}

	p := NewExprParser()
	expr, err := p.Parse(tok)
	if err != nil {
		return nil
	}

	deref, ok := expr.(*ObjectDerefNode)
	if !ok {
		return nil
	}
	recv, ok := deref.Receiver.(*VariableNode)
	if !ok {
		return nil
	}
	if recv.Name != "matrix" {
		return nil
	}

	prop := deref.Property
	labels := []*String{}

	if m.Rows != nil {
		if row, ok := m.Rows[prop]; ok {
			for _, v := range row.Values {
				if s, ok := v.(*RawYAMLString); ok {
					labels = append(labels, &String{s.Value, s.Pos()})
				}
			}
		}
	}

	for _, inc := range m.Include {
		if c, ok := inc[prop]; ok {
			if s, ok := c.Value.(*RawYAMLString); ok {
				labels = append(labels, &String{s.Value, s.Pos()})
			}
		}
	}

	return labels
}
