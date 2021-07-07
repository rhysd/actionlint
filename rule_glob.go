package actionlint

// RuleGlob is a rule to check glob syntax.
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
type RuleGlob struct {
	RuleBase
}

// NewRuleGlob creates new RuleGlob instance.
func NewRuleGlob() *RuleGlob {
	return &RuleGlob{
		RuleBase: RuleBase{name: "glob"},
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleGlob) VisitWorkflowPre(n *Workflow) error {
	for _, e := range n.On {
		if w, ok := e.(*WebhookEvent); ok {
			rule.checkGitRefGlobs(w.Branches)
			rule.checkGitRefGlobs(w.BranchesIgnore)
			rule.checkGitRefGlobs(w.Tags)
			rule.checkGitRefGlobs(w.TagsIgnore)
			rule.checkFilePathGlobs(w.Paths)
			rule.checkFilePathGlobs(w.PathsIgnore)
		}
	}
	return nil
}

func (rule *RuleGlob) checkGitRefGlobs(names []*String) {
	for _, n := range names {
		errs := ValidateRefGlob(n.Value)
		for i := range errs {
			rule.globSyntaxError(&errs[i], *n.Pos)
		}
	}
}

func (rule *RuleGlob) checkFilePathGlobs(paths []*String) {
	for _, p := range paths {
		errs := ValidateRefGlob(p.Value)
		for i := range errs {
			rule.globSyntaxError(&errs[i], *p.Pos)
		}
	}
}

func (rule *RuleGlob) globSyntaxError(err *InvalidGlobPattern, pos Pos) {
	if err.Column != 0 {
		pos.Col += err.Column - 1
	}
	rule.errorf(&pos, "%s. note: filter pattern syntax is explained at https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet", err.Message)
}
