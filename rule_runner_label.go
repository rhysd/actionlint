package actionlint

import (
	"sort"
	"strings"
)

type runnerLabelCompat uint

const (
	compatInvalid                      = 0
	compatUbuntu1604 runnerLabelCompat = 1 << iota
	compatUbuntu1804
	compatUbuntu2004
	compatMacOS1015
	compatMacOS110
	compatWindows2016
	compatWindows2019
)

type runnerLabelCompats map[string]runnerLabelCompat

var githubHostedRunnerCompats = runnerLabelCompats{
	"ubuntu-latest":  compatUbuntu2004,
	"ubuntu-20.04":   compatUbuntu2004,
	"ubuntu-18.04":   compatUbuntu1804,
	"ubuntu-16.04":   compatUbuntu1604,
	"macos-latest":   compatMacOS110,
	"macos-11":       compatMacOS110,
	"macos-11.0":     compatMacOS110,
	"macos-10.15":    compatMacOS1015,
	"windows-latest": compatWindows2019,
	"windows-2019":   compatWindows2019,
	"windows-2016":   compatWindows2016,
}

func (compat runnerLabelCompats) workerNames() []string {
	ns := make([]string, 0, len(compat))
	for n := range compat {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	return ns
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
	// Note: Using only one compatibility integer is enough to check compatibility. But we remember
	// all past compatibility values here for better error message. If accumulating all compatibility
	// values into one integer, we can no longer know what labels are conflicting.
	compats map[runnerLabelCompat]*String
}

// NewRuleRunnerLabel creates new RuleRunnerLabel instance.
func NewRuleRunnerLabel(labels []string) *RuleRunnerLabel {
	return &RuleRunnerLabel{
		RuleBase:    RuleBase{name: "runner-label"},
		knownLabels: labels,
		compats:     nil,
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleRunnerLabel) VisitJobPre(n *Job) error {
	if n.RunsOn == nil {
		return nil
	}

	var m *Matrix
	if n.Strategy != nil {
		m = n.Strategy.Matrix
	}

	if len(n.RunsOn.Labels) == 1 {
		rule.checkLabel(n.RunsOn.Labels[0], m)
		return nil
	}

	rule.compats = map[runnerLabelCompat]*String{}
	for _, label := range n.RunsOn.Labels {
		rule.checkLabelAndConflict(label, m)
	}

	rule.compats = nil // reset
	return nil
}

// https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners
func (rule *RuleRunnerLabel) checkLabelAndConflict(label *String, m *Matrix) {
	if l := label.Value; strings.Contains(l, "${{") {
		ls := rule.tryToGetLabelsInMatrix(l, m)
		cs := make([]runnerLabelCompat, 0, len(ls))
		for _, l := range ls {
			comp := rule.verifyRunnerLabel(l)
			cs = append(cs, comp)
		}
		rule.checkCombiCompat(cs, ls)
		return
	}

	comp := rule.verifyRunnerLabel(label)
	rule.checkCompat(comp, label)
}

func (rule *RuleRunnerLabel) checkLabel(label *String, m *Matrix) {
	if l := label.Value; strings.Contains(l, "${{") {
		ls := rule.tryToGetLabelsInMatrix(l, m)
		for _, l := range ls {
			rule.verifyRunnerLabel(l)
		}
		return
	}

	rule.verifyRunnerLabel(label)
}

func (rule *RuleRunnerLabel) verifyRunnerLabel(label *String) runnerLabelCompat {
	l := label.Value
	if c, ok := githubHostedRunnerCompats[strings.ToLower(l)]; ok {
		return c
	}

	for _, p := range allSelfHostedRunnerPresetLabels {
		if strings.EqualFold(l, p) {
			return compatInvalid
		}
	}

	for _, k := range rule.knownLabels {
		if strings.EqualFold(l, k) {
			return compatInvalid
		}
	}

	rule.errorf(
		label.Pos,
		"label %q is unknown. available labels are %s. if it is a custom label for self-hosted runner, set list of labels in actionlint.yaml config file",
		label.Value,
		quotesAll(
			githubHostedRunnerCompats.workerNames(),
			allSelfHostedRunnerPresetLabels,
			rule.knownLabels,
		),
	)

	return compatInvalid
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

	p := NewExprParser()
	expr, err := p.Parse(NewExprLexer(l[3:])) // 3 means omit first "${{"
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
					labels = append(labels, &String{s.Value, false, s.Pos()})
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
						labels = append(labels, &String{s.Value, false, s.Pos()})
					}
				}
			}
		}
	}

	return labels
}

func (rule *RuleRunnerLabel) checkConflict(comp runnerLabelCompat, label *String) bool {
	for c, l := range rule.compats {
		if c&comp == 0 {
			rule.errorf(label.Pos, "label %q conflicts with label %q defined at %s. note: to run your job on each workers, use matrix", label.Value, l.Value, l.Pos)
			return false
		}
	}
	return true
}

func (rule *RuleRunnerLabel) checkCompat(comp runnerLabelCompat, label *String) {
	if comp == compatInvalid || !rule.checkConflict(comp, label) {
		return
	}
	if _, ok := rule.compats[comp]; !ok {
		rule.compats[comp] = label
	}
}

func (rule *RuleRunnerLabel) checkCombiCompat(comps []runnerLabelCompat, labels []*String) {
	for i, c := range comps {
		if c != compatInvalid && !rule.checkConflict(c, labels[i]) {
			// Overwrite the compatibility value with compatInvalid at conflicted label not to
			// register the label to `rule.compats`.
			comps[i] = compatInvalid
		}
	}
	for i, c := range comps {
		if c != compatInvalid {
			if _, ok := rule.compats[c]; !ok {
				rule.compats[c] = labels[i]
			}
		}
	}
}
