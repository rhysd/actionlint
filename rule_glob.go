package actionlint

// RuleGlob is a rule to check glob syntax.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
type RuleGlob struct {
	RuleBase
}

// NewRuleGlob creates new RuleGlob instance.
func NewRuleGlob() *RuleGlob {
	return &RuleGlob{
		RuleBase: RuleBase{
			name: "glob",
			desc: "Checks for glob syntax used in branch names, tags, and paths",
		},
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleGlob) VisitWorkflowPre(n *Workflow) error {
	for _, e := range n.On {
		switch e := e.(type) {
		case *WebhookEvent:
			rule.checkGitRefGlobs(e.Branches)
			rule.checkGitRefGlobs(e.BranchesIgnore)
			rule.checkGitRefGlobs(e.Tags)
			rule.checkGitRefGlobs(e.TagsIgnore)
			rule.checkFilePathGlobs(e.Paths)
			rule.checkFilePathGlobs(e.PathsIgnore)
		case *ImageVersionEvent:
			for _, v := range e.Versions {
				rule.checkRefGlob(v)
			}
		}
	}
	return nil
}

func (rule *RuleGlob) VisitJobPre(n *Job) error {
	if n.Snapshot != nil && n.Snapshot.Version != nil {
		rule.checkRefGlob(n.Snapshot.Version)
	}
	return nil
}

func (rule *RuleGlob) checkRefGlob(s *String) {
	// Empty value is already checked by parser. Avoid duplicate errors
	if s == nil || s.Value == "" {
		return
	}
	rule.globErrors(ValidateRefGlob(s.Value), s.Pos, s.Quoted)
}

func (rule *RuleGlob) checkGitRefGlobs(filter *WebhookEventFilter) {
	if filter == nil {
		return
	}
	for _, v := range filter.Values {
		rule.checkRefGlob(v)
	}
}

func (rule *RuleGlob) checkFilePathGlobs(filter *WebhookEventFilter) {
	if filter == nil {
		return
	}
	for _, v := range filter.Values {
		// Empty value is already checked by parser. Avoid duplicate errors
		if v.Value != "" {
			rule.globErrors(ValidatePathGlob(v.Value), v.Pos, v.Quoted)
		}
	}
}

func (rule *RuleGlob) globErrors(errs []InvalidGlobPattern, pos *Pos, quoted bool) {
	for i := range errs {
		err := &errs[i]
		p := *pos
		if quoted {
			p.Col++
		}
		if err.Column != 0 {
			p.Col += err.Column - 1
		}
		rule.Errorf(&p, "%s. note: filter pattern syntax is explained at https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet", err.Message)
	}
}
