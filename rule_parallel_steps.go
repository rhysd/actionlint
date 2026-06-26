package actionlint

import "strings"

// RuleParallelSteps is a rule to check parallel steps: a 'wait' or 'cancel' step must refer to the ID
// of a preceding background step, and a 'parallel' group may only contain 'run' and 'uses' steps
// ('background', 'wait', 'wait-all', 'cancel', and nested 'parallel' steps are not allowed in it).
// https://github.blog/changelog/2026-06-25-actions-steps-can-now-be-run-in-parallel/
type RuleParallelSteps struct {
	RuleBase
	// background holds the IDs (lower-cased; step IDs are case insensitive) of the 'background: true'
	// steps seen so far in the current job. Steps are visited in order, including steps nested in
	// 'parallel:' groups, so a 'wait'/'cancel' reference found in this set both exists and precedes
	// the reference.
	background map[string]struct{}
	// inParallel holds the steps that are direct children of a 'parallel' group. Such steps are not
	// allowed there at all (only 'run'/'uses' are), so their 'wait'/'cancel' references are not
	// separately validated: the "not allowed inside a parallel group" error stands on its own
	// instead of cascading into a redundant "unknown reference" error from the same step.
	inParallel map[*Step]struct{}
}

// NewRuleParallelSteps creates a new RuleParallelSteps instance.
func NewRuleParallelSteps() *RuleParallelSteps {
	return &RuleParallelSteps{
		RuleBase: RuleBase{
			name: "parallel-steps",
			desc: "Checks \"wait\"/\"cancel\" references to background steps and steps forbidden inside a \"parallel\" group",
		},
	}
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleParallelSteps) VisitJobPre(n *Job) error {
	rule.background = map[string]struct{}{}
	rule.inParallel = map[*Step]struct{}{}
	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RuleParallelSteps) VisitJobPost(n *Job) error {
	rule.background = nil
	rule.inParallel = nil
	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleParallelSteps) VisitStep(n *Step) error {
	// A step that is itself inside a 'parallel' group is already reported by checkParallelChildren as
	// not allowed there, so skip its reference check to avoid a second, redundant error.
	_, insideParallel := rule.inParallel[n]
	switch e := n.Exec.(type) {
	case *ExecWait:
		if !insideParallel {
			for _, name := range e.Names {
				rule.checkRef(name)
			}
		}
	case *ExecCancel:
		if !insideParallel {
			rule.checkRef(e.Name)
		}
	case *ExecParallel:
		rule.checkParallelChildren(e.Steps)
	}

	if n.ID != nil && !n.ID.ContainsExpression() && isBackgroundStep(n) {
		rule.background[strings.ToLower(n.ID.Value)] = struct{}{}
	}

	return nil
}

// checkParallelChildren checks the steps inside a 'parallel' group. Only 'run' and 'uses' steps are
// allowed there: 'background', 'wait', 'wait-all', 'cancel', and nested 'parallel' steps are not,
// because steps in a 'parallel' group already run in the background with an implicit wait at the end.
func (rule *RuleParallelSteps) checkParallelChildren(steps []*Step) {
	for _, s := range steps {
		rule.inParallel[s] = struct{}{}
		if s.Background != nil {
			rule.Errorf(s.Background.Pos, "\"background\" is not allowed for a step inside a \"parallel\" group because the group's steps already run in the background")
		}
		switch e := s.Exec.(type) {
		case *ExecWait:
			kind := "wait"
			if e.All {
				kind = "wait-all"
			}
			rule.Errorf(s.Pos, "%q step is not allowed inside a \"parallel\" group", kind)
		case *ExecCancel:
			rule.Errorf(s.Pos, "\"cancel\" step is not allowed inside a \"parallel\" group")
		case *ExecParallel:
			rule.Errorf(s.Pos, "\"parallel\" step cannot be nested in another \"parallel\" step")
		}
	}
}

func (rule *RuleParallelSteps) checkRef(ref *String) {
	if ref == nil || ref.Value == "" || ref.ContainsExpression() {
		// Empty values come from an already-reported parse error; expression step IDs can't be
		// resolved statically. Skip both.
		return
	}
	if _, ok := rule.background[strings.ToLower(ref.Value)]; !ok {
		rule.Errorf(
			ref.Pos,
			"%q is not the ID of a preceding background step. \"wait\" and \"cancel\" steps can only refer to an earlier step that has \"background: true\"",
			ref.Value,
		)
	}
}

// isBackgroundStep reports whether the step runs (or may run) in the background. When 'background' is
// an expression, whether the step runs in the background can't be known statically, so it's treated
// as a possible background step to avoid false positives.
func isBackgroundStep(s *Step) bool {
	return s.Background != nil && (s.Background.Expression != nil || s.Background.Value)
}
