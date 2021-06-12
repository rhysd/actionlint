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

// https://docs.github.com/en/actions/hosting-your-own-runners/using-self-hosted-runners-in-a-workflow#using-default-labels-to-route-jobs
var allSelfHostedRunnerPresetLabels = []string{
	"self-hosted",
	"linux",
	"macos",
	"windows",
	"x64",
	"arm",
	"arm64",
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

	// TODO: Check labels conflict
	// runs-on: [linux, windows]

	for _, label := range n.RunsOn.Labels {
		rule.checkLabel(label, m)
	}
}

// https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners
func (rule *RuleRunnerLabel) checkLabel(label *String, m *Matrix) {
	l := label.Value
	if strings.Contains(l, "${{") && strings.Contains(l, "}}") {
		for _, l := range rule.tryToGetLabelsInMatrix(l, m) {
			rule.verifyRunnerLabel(l)
		}
		return
	}

	rule.verifyRunnerLabel(label)
}

func (rule *RuleRunnerLabel) verifyRunnerLabel(label *String) string {
	l := strings.TrimSpace(label.Value)
	for _, p := range allGitHubHostedRunnerLabels {
		// use EqualFold for ignoring case e.g. both macos-latest and macOS-latest should be accepted
		if strings.EqualFold(l, p) {
			return p // ok
		}
	}
	for _, p := range allSelfHostedRunnerPresetLabels {
		if strings.EqualFold(l, p) {
			return p // ok
		}
	}
	for _, k := range rule.knownLabels {
		if strings.EqualFold(l, k) {
			return k // ok
		}
	}

	qs := make([]string, 0, len(allGitHubHostedRunnerLabels)+len(allSelfHostedRunnerPresetLabels)+len(rule.knownLabels))
	for _, ls := range [][]string{
		allGitHubHostedRunnerLabels,
		allSelfHostedRunnerPresetLabels,
		rule.knownLabels,
	} {
		for _, l := range ls {
			qs = append(qs, strconv.Quote(l))
		}
	}
	rule.errorf(
		label.Pos,
		"label %q is unknown. available labels are %s. if it is a custom label for self-hosted runner, set list of labels in actionlint.yaml config file",
		label.Value,
		strings.Join(qs, ", "),
	)

	return ""
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
				if s, ok := v.(*RawYAMLString); ok && !strings.Contains(s.Value, "${{") {
					// When the value does not have expression syntax ${{ }}
					labels = append(labels, &String{s.Value, s.Pos()})
				}
			}
		}
	}

	if m.Include != nil {
		for _, combi := range m.Include.Combinations {
			if combi.Assigns != nil {
				if assign, ok := combi.Assigns[prop]; ok {
					if s, ok := assign.Value.(*RawYAMLString); ok && !strings.Contains(s.Value, "${{") {
						// When the value does not have expression syntax ${{ }}
						labels = append(labels, &String{s.Value, s.Pos()})
					}
				}
			}
		}
	}

	return labels
}
