package actionlint

import "strings"

// RuleExpression is a rule checker to check expression syntax in string values of workflow syntax.
// It checks syntax and semantics of the expressions including type checks and functions/contexts
// definitions. For more details see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type RuleExpression struct {
	RuleBase
}

// NewRuleExpression creates new RuleExpression instance.
func NewRuleExpression() *RuleExpression {
	return &RuleExpression{
		RuleBase: RuleBase{name: "expression"},
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleExpression) VisitWorkflowPre(n *Workflow) {
	rule.checkString(n.Name)

	for _, e := range n.On {
		switch e := e.(type) {
		case *WebhookEvent:
			rule.checkStrings(e.Types)
			rule.checkStrings(e.Branches)
			rule.checkStrings(e.BranchesIgnore)
			rule.checkStrings(e.Tags)
			rule.checkStrings(e.TagsIgnore)
			rule.checkStrings(e.Paths)
			rule.checkStrings(e.PathsIgnore)
			rule.checkStrings(e.Workflows)
		case *ScheduledEvent:
			rule.checkStrings(e.Cron)
		case *WorkflowDispatchEvent:
			for _, i := range e.Inputs {
				rule.checkString(i.Description)
				rule.checkString(i.Default)
			}
		case *RepositoryDispatchEvent:
			rule.checkStrings(e.Types)
		}
	}

	for _, e := range n.Env {
		rule.checkString(e.Value)
	}

	rule.checkDefaults(n.Defaults)
	rule.checkConcurrency(n.Concurrency)
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleExpression) VisitJobPre(n *Job) {
	rule.checkString(n.Name)
	rule.checkStrings(n.Needs)

	switch runner := n.RunsOn.(type) {
	case *GitHubHostedRunner:
		rule.checkString(runner.Label)
	case *SelfHostedRunner:
		rule.checkStrings(runner.Labels)
	}

	if n.Environment != nil {
		rule.checkString(n.Environment.Name)
		rule.checkString(n.Environment.URL)
	}

	rule.checkConcurrency(n.Concurrency)

	for _, output := range n.Outputs {
		rule.checkString(output.Value)
	}

	for _, e := range n.Env {
		rule.checkString(e.Value)
	}

	rule.checkDefaults(n.Defaults)
	rule.checkBoolString("if", n.If)

	if n.Strategy != nil && n.Strategy.Matrix != nil {
		for _, r := range n.Strategy.Matrix.Rows {
			rule.checkStrings(r.Values)
		}
		for _, cs := range n.Strategy.Matrix.Include {
			for _, c := range cs {
				rule.checkString(c.Value)
			}
		}
		for _, cs := range n.Strategy.Matrix.Exclude {
			for _, c := range cs {
				rule.checkString(c.Value)
			}
		}
	}

	rule.checkContainer(n.Container)

	for _, s := range n.Services {
		rule.checkContainer(s.Container)
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleExpression) VisitStep(n *Step) {
	rule.checkBoolString("if", n.If)
	rule.checkString(n.Name)

	switch e := n.Exec.(type) {
	case *ExecRun:
		rule.checkString(e.Run)
		rule.checkString(e.Shell)
		rule.checkString(e.WorkingDirectory)
	case *ExecAction:
		rule.checkString(e.Uses)
		for _, i := range e.Inputs {
			rule.checkString(i.Value)
		}
		rule.checkString(e.Entrypoint)
		rule.checkString(e.Args)
	}

	for _, e := range n.Env {
		rule.checkString(e.Value)
	}
}

func (rule *RuleExpression) checkContainer(c *Container) {
	if c == nil {
		return
	}
	rule.checkString(c.Image)
	if c.Credentials != nil {
		rule.checkString(c.Credentials.Username)
		rule.checkString(c.Credentials.Password)
	}
	for _, e := range c.Env {
		rule.checkString(e.Value)
	}
	rule.checkStrings(c.Ports)
	rule.checkStrings(c.Volumes)
	rule.checkString(c.Options)
}

func (rule *RuleExpression) checkConcurrency(c *Concurrency) {
	if c == nil {
		return
	}
	rule.checkString(c.Group)
}

func (rule *RuleExpression) checkDefaults(d *Defaults) {
	if d == nil || d.Run == nil {
		return
	}
	rule.checkString(d.Run.Shell)
	rule.checkString(d.Run.WorkingDirectory)
}

func (rule *RuleExpression) checkStrings(ss []*String) {
	for _, s := range ss {
		rule.checkString(s)
	}
}

func (rule *RuleExpression) checkBoolString(sec string, str *String) {
	if str == nil {
		return
	}
	tys := rule.checkString(str)
	if len(tys) == 0 {
		if str.Value != "true" && str.Value != "false" {
			rule.errorf(str.Pos, "expected bool string \"true\" or \"false\" but got %q", str.Value)
		}
		return
	}

	if len(tys) > 1 {
		return // The string contains two or more placeholders. Give up
	}
	s := strings.TrimSpace(str.Value)
	if !strings.HasPrefix(s, "${{") || !strings.HasSuffix(s, "}}") {
		return // When return value is not entire string of the section, give up
	}

	switch ty := tys[0].(type) {
	case BoolType:
		// OK
	default:
		rule.errorf(str.Pos, "value at %q section should be type \"bool\" but got type %q", sec, ty.String())
	}
}

func (rule *RuleExpression) checkString(str *String) []ExprType {
	if str == nil {
		return nil
	}

	line, col := str.Pos.Line, str.Pos.Col
	offset := 0
	tys := []ExprType{}
	s := str.Value
	for {
		idx := strings.Index(s, "${{")
		if idx == -1 {
			break
		}

		start := idx + 3 // 3 means removing "${{"
		s = s[start:]
		offset += start

		ty, offsetAfter := rule.checkSemantics(s, line, col+offset)
		s = s[offsetAfter:]
		offset += offsetAfter
		tys = append(tys, ty)
	}

	return tys
}

func (rule *RuleExpression) exprError(err *ExprError, lineBase, colBase int) {
	// Line and column in ExprError are 1-based
	line := err.Line - 1 + lineBase
	col := err.Column - 1 + colBase
	pos := Pos{Line: line, Col: col}
	rule.error(&pos, err.Message)
}

func (rule *RuleExpression) checkSemantics(src string, line, col int) (ExprType, int) {
	l := NewExprLexer()
	tok, offset, err := l.Lex(src)
	if err != nil {
		rule.exprError(err, line, col)
		return nil, offset
	}

	p := NewExprParser()
	expr, err := p.Parse(tok)
	if err != nil {
		rule.exprError(err, line, col)
		return nil, offset
	}

	// TODO: The first return value should be used to check correct value is specified
	c := NewExprSemanticsChecker()
	ty, errs := c.Check(expr)
	for _, err := range errs {
		rule.exprError(err, line, col)
	}

	return ty, offset
}
