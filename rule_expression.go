package actionlint

import (
	"fmt"
	"strconv"
	"strings"
)

// RuleExpression is a rule checker to check expression syntax in string values of workflow syntax.
// It checks syntax and semantics of the expressions including type checks and functions/contexts
// definitions. For more details see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type RuleExpression struct {
	RuleBase
	matrixTy *ObjectType
	stepsTy  *ObjectType
	needsTy  *ObjectType
	workflow *Workflow
}

// NewRuleExpression creates new RuleExpression instance.
func NewRuleExpression() *RuleExpression {
	return &RuleExpression{
		RuleBase: RuleBase{name: "expression"},
		matrixTy: nil,
		stepsTy:  nil,
		needsTy:  nil,
		workflow: nil,
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
				rule.checkBool(i.Required)
			}
		case *RepositoryDispatchEvent:
			rule.checkStrings(e.Types)
		}
	}

	rule.checkEnv(n.Env)

	rule.checkDefaults(n.Defaults)
	rule.checkConcurrency(n.Concurrency)

	rule.workflow = n
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children
func (rule *RuleExpression) VisitWorkflowPost(n *Workflow) {
	rule.workflow = nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleExpression) VisitJobPre(n *Job) {
	// Type of needs must be resolved before resolving type of matrix because `needs` context can
	// be used in matrix configuration.
	rule.needsTy = rule.calcNeedsType(n)

	// Set matrix type at start of VisitJobPre() because matrix values are available in
	// jobs.<job_id> section. For example:
	//   jobs:
	//     foo:
	//       strategy:
	//         matrix:
	//           os: [ubuntu-latest, macos-latest, windows-latest]
	//       runs-on: ${{ matrix.os }}
	if n.Strategy != nil && n.Strategy.Matrix != nil {
		rule.matrixTy = rule.guessTypeOfMatrix(n.Strategy.Matrix)
	}

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

	rule.checkEnv(n.Env)

	rule.checkDefaults(n.Defaults)
	rule.checkIfCondition(n.If)

	if n.Strategy != nil {
		if n.Strategy.Matrix != nil {
			for _, r := range n.Strategy.Matrix.Rows {
				for _, v := range r.Values {
					rule.checkRawYAMLValue(v)
				}
			}
			rule.checkMatrixCombinations(n.Strategy.Matrix.Include, "include")
			rule.checkMatrixCombinations(n.Strategy.Matrix.Exclude, "exclude")
		}
		rule.checkBool(n.Strategy.FailFast)
		rule.checkInt(n.Strategy.MaxParallel)
	}

	rule.checkBool(n.ContinueOnError)
	rule.checkFloat(n.TimeoutMinutes)
	rule.checkContainer(n.Container)

	for _, s := range n.Services {
		rule.checkContainer(s.Container)
	}

	rule.stepsTy = NewStrictObjectType()
}

// VisitJobPost is callback when visiting Job node after visiting its children
func (rule *RuleExpression) VisitJobPost(n *Job) {
	// outputs section is evaluated after all steps are run
	for _, output := range n.Outputs {
		rule.checkString(output.Value)
	}

	rule.matrixTy = nil
	rule.stepsTy = nil
	rule.needsTy = nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleExpression) VisitStep(n *Step) {
	rule.checkIfCondition(n.If)
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

	rule.checkEnv(n.Env)
	rule.checkBool(n.ContinueOnError)
	rule.checkFloat(n.TimeoutMinutes)

	if n.ID != nil {
		// Step ID is case insensitive
		id := strings.ToLower(n.ID.Value)
		rule.stepsTy.Props[id] = &ObjectType{
			Props: map[string]ExprType{
				"outputs":    NewObjectType(),
				"conclusion": StringType{},
				"outcome":    StringType{},
			},
			StrictProps: true,
		}
	}
}

func (rule *RuleExpression) checkOneExpression(s *String, what string) ExprType {
	if s == nil {
		return nil
	}
	tys := rule.checkString(s)
	if len(tys) != 1 {
		// This case should be unreachable since only one ${{ }} is included is checked by parser
		rule.errorf(s.Pos, "one ${{ }} expression should be included in %q value but got %d expressions", what, len(tys))
		return nil
	}
	return tys[0]
}

func (rule *RuleExpression) checkObjectTy(ty ExprType, pos *Pos, what string) ExprType {
	if ty == nil {
		return nil
	}
	switch ty.(type) {
	case *ObjectType, AnyType:
		return ty
	default:
		rule.errorf(pos, "type of expression at %q must be object but found type %s", what, ty.String())
		return nil
	}
}

func (rule *RuleExpression) checkArrayTy(ty ExprType, pos *Pos, what string) ExprType {
	if ty == nil {
		return nil
	}
	switch ty.(type) {
	case *ArrayType, *ArrayDerefType, AnyType:
		return ty
	default:
		rule.errorf(pos, "type of expression at %q must be array but found type %s", what, ty.String())
		return nil
	}
}

func (rule *RuleExpression) checkNumberTy(ty ExprType, pos *Pos, what string) ExprType {
	if ty == nil {
		return nil
	}
	switch ty.(type) {
	case NumberType, AnyType:
		return ty
	default:
		rule.errorf(pos, "type of expression at %q must be array but found type %s", what, ty.String())
		return nil
	}
}

func (rule *RuleExpression) checkObjectExpression(s *String, what string) ExprType {
	ty := rule.checkOneExpression(s, what)
	if ty == nil {
		return nil
	}
	return rule.checkObjectTy(ty, s.Pos, what)
}

func (rule *RuleExpression) checkArrayExpression(s *String, what string) ExprType {
	ty := rule.checkOneExpression(s, what)
	if ty == nil {
		return nil
	}
	return rule.checkArrayTy(ty, s.Pos, what)
}

func (rule *RuleExpression) checkNumberExpression(s *String, what string) ExprType {
	ty := rule.checkOneExpression(s, what)
	if ty == nil {
		return nil
	}
	return rule.checkNumberTy(ty, s.Pos, what)
}

func (rule *RuleExpression) checkMatrixCombinations(cs *MatrixCombinations, what string) {
	if cs == nil {
		return
	}

	if cs.Expression != nil {
		ty := rule.checkArrayExpression(cs.Expression, what)
		if elem, ok := ElemTypeOf(ty); ok {
			rule.checkObjectTy(elem, cs.Expression.Pos, what)
		}
		return
	}

	what = fmt.Sprintf("matrix combination at element of %s section", what)
	for _, combi := range cs.Combinations {
		if combi.Expression != nil {
			rule.checkObjectExpression(combi.Expression, what)
			continue
		}
		for _, a := range combi.Assigns {
			rule.checkRawYAMLValue(a.Value)
		}
	}
}

func (rule *RuleExpression) checkEnv(env *Env) {
	if env == nil {
		return
	}

	if env.Vars != nil {
		for _, e := range env.Vars {
			rule.checkString(e.Value)
		}
		return
	}

	// When form of "env: ${{...}}"
	rule.checkObjectExpression(env.Expression, "env")
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
	rule.checkEnv(c.Env)
	rule.checkStrings(c.Ports)
	rule.checkStrings(c.Volumes)
	rule.checkString(c.Options)
}

func (rule *RuleExpression) checkConcurrency(c *Concurrency) {
	if c == nil {
		return
	}
	rule.checkString(c.Group)
	rule.checkBool(c.CancelInProgress)
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

func (rule *RuleExpression) checkIfCondition(str *String) {
	if str == nil {
		return
	}

	// Note:
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idif
	//
	// > When you use expressions in an if conditional, you may omit the expression syntax (${{ }})
	// > because GitHub automatically evaluates the if conditional as an expression, unless the
	// > expression contains any operators. If the expression contains any operators, the expression
	// > must be contained within ${{ }} to explicitly mark it for evaluation.
	//
	// This document is wrong. Actually I confirmed that any strings without surrounding in ${{ }}
	// are evaluated.
	//
	// - run: echo 'run'
	//   if: true
	// - run: echo 'not run'
	//   if: false
	// - run: echo 'run'
	//   if: contains('abc', 'ab')
	// - run: echo 'not run'
	//   if: contains('abc', 'xy')

	var condTy ExprType
	if strings.Contains(str.Value, "${{") && strings.Contains(str.Value, "}}") {
		if tys := rule.checkString(str); len(tys) == 1 {
			s := strings.TrimSpace(str.Value)
			if strings.HasPrefix(s, "${{") && strings.HasSuffix(s, "}}") {
				condTy = tys[0]
			}
		}
	} else {
		src := str.Value + "}}" // }} is necessary since lexer lexes it as end of tokens
		condTy, _ = rule.checkSemantics(src, str.Pos.Line, str.Pos.Col)
	}

	if condTy != nil && !(BoolType{}).Assignable(condTy) {
		rule.errorf(str.Pos, "\"if\" condition should be type \"bool\" but got type %q", condTy.String())
	}
}

func (rule *RuleExpression) checkString(str *String) []ExprType {
	if str == nil {
		return nil
	}
	return rule.checkExprsIn(str.Value, str.Pos)
}

func (rule *RuleExpression) checkBool(b *Bool) {
	if b == nil || b.Expression == nil {
		return
	}
	ty := rule.checkOneExpression(b.Expression, "bool value")
	switch ty.(type) {
	case BoolType, AnyType:
		// ok
	default:
		rule.errorf(b.Expression.Pos, "type of expression must be bool but found type %s", ty.String())
	}
}

func (rule *RuleExpression) checkInt(i *Int) {
	if i == nil {
		return
	}
	rule.checkNumberExpression(i.Expression, "integer value")
}

func (rule *RuleExpression) checkFloat(f *Float) {
	if f == nil {
		return
	}
	rule.checkNumberExpression(f.Expression, "float number value")
}

func (rule *RuleExpression) checkExprsIn(s string, pos *Pos) []ExprType {
	line, col := pos.Line, pos.Col
	offset := 0
	tys := []ExprType{}
	for {
		idx := strings.Index(s, "${{")
		if idx == -1 {
			break
		}

		start := idx + 3 // 3 means removing "${{"
		s = s[start:]
		offset += start

		ty, offsetAfter := rule.checkSemantics(s, line, col+offset)
		if ty == nil || offsetAfter == 0 {
			return nil
		}

		s = s[offsetAfter:]
		offset += offsetAfter
		tys = append(tys, ty)
	}

	return tys
}

func (rule *RuleExpression) checkRawYAMLValue(v RawYAMLValue) {
	switch v := v.(type) {
	case *RawYAMLObject:
		for _, p := range v.Props {
			rule.checkRawYAMLValue(p)
		}
	case *RawYAMLArray:
		for _, v := range v.Elems {
			rule.checkRawYAMLValue(v)
		}
	case *RawYAMLString:
		rule.checkExprsIn(v.Value, v.Pos())
	default:
		panic("unreachable")
	}
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

	c := NewExprSemanticsChecker()
	if rule.matrixTy != nil {
		c.UpdateMatrix(rule.matrixTy)
	}
	if rule.stepsTy != nil {
		c.UpdateSteps(rule.stepsTy)
	}
	if rule.needsTy != nil {
		c.UpdateNeeds(rule.needsTy)
	}
	ty, errs := c.Check(expr)
	for _, err := range errs {
		rule.exprError(err, line, col)
	}

	return ty, offset
}

func (rule *RuleExpression) calcNeedsType(job *Job) *ObjectType {
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#needs-context
	o := NewStrictObjectType()
	rule.populateDependantNeedsTypes(o, job, job)
	return o
}

func (rule *RuleExpression) populateDependantNeedsTypes(out *ObjectType, job *Job, root *Job) {
	for _, id := range job.Needs {
		if id.Value == root.ID.Value {
			continue // When cyclic dependency exists. This does not happen normally.
		}
		if _, ok := out.Props[id.Value]; ok {
			continue // Already added
		}

		j, ok := rule.workflow.Jobs[id.Value]
		if !ok {
			continue
		}

		outputs := NewStrictObjectType()
		for name := range j.Outputs {
			outputs.Props[name] = StringType{}
		}

		out.Props[id.Value] = &ObjectType{
			Props: map[string]ExprType{
				"outputs": outputs,
				"result":  StringType{},
			},
			StrictProps: true,
		}

		rule.populateDependantNeedsTypes(out, j, root) // Add necessary needs props recursively
	}
}

func (rule *RuleExpression) guessTypeOfMatrixExpression(expr *String) *ObjectType {
	matTy, ok := rule.checkObjectExpression(expr, "matrix").(*ObjectType)
	if !ok {
		return NewObjectType()
	}

	// Consider properties in include section elements since 'include' section adds matrix values
	incTy, ok := matTy.Props["include"]
	if ok {
		delete(matTy.Props, "include")
		if elem, ok := ElemTypeOf(incTy); ok {
			if o, ok := elem.(*ObjectType); ok {
				for n, p := range o.Props {
					t, ok := matTy.Props[n]
					if !ok {
						matTy.Props[n] = p
						continue
					}
					matTy.Props[n] = t.Fuse(p)
				}
			}
		}
	}

	delete(matTy.Props, "exclude")

	return matTy
}

func (rule *RuleExpression) guessTypeOfMatrix(m *Matrix) *ObjectType {
	if m.Expression != nil {
		return rule.guessTypeOfMatrixExpression(m.Expression)
	}

	o := NewStrictObjectType()

	for n, r := range m.Rows {
		o.Props[n] = rule.guessTypeOfMatrixRow(r)
	}

	// Note: Type check in 'include' section duplicates with checkMatrixCombinations() method

	if m.Include == nil {
		return o
	}

	if m.Include.Expression != nil {
		elem, ok := ElemTypeOf(rule.checkOneExpression(m.Include.Expression, "include"))
		if !ok {
			return NewObjectType()
		}
		ret, ok := o.Fuse(elem).(*ObjectType)
		if !ok {
			return NewObjectType()
		}
		return ret
	}

	for _, combi := range m.Include.Combinations {
		if combi.Expression != nil {
			ty := rule.checkOneExpression(m.Include.Expression, "matrix combination at element of include section")
			if fused, ok := o.Fuse(ty).(*ObjectType); ok {
				o = fused
			} else {
				o.StrictProps = false
			}
			continue
		}

		for n, assign := range combi.Assigns {
			ty := guessTypeOfRawYAMLValue(assign.Value)
			if t, ok := o.Props[n]; ok {
				// When the combination exists in 'matrix' section, fuse type into existing one
				ty = t.Fuse(ty)
			}
			o.Props[n] = ty
		}
	}

	// Note: m.Exclude is not considered when guessing type of matrix

	return o
}

func (rule *RuleExpression) guessTypeOfMatrixRow(r *MatrixRow) ExprType {
	if r.Expression != nil {
		ty, ok := ElemTypeOf(rule.checkArrayExpression(r.Expression, "matrix row"))
		if !ok {
			return AnyType{}
		}
		return ty
	}

	var ty ExprType
	for _, v := range r.Values {
		t := guessTypeOfRawYAMLValue(v)
		if ty == nil {
			ty = t
		} else {
			ty = ty.Fuse(t)
		}
	}

	if ty == nil {
		return AnyType{} // No element
	}

	return ty
}

func guessTypeOfRawYAMLValue(v RawYAMLValue) ExprType {
	switch v := v.(type) {
	case *RawYAMLObject:
		m := make(map[string]ExprType, len(v.Props))
		for k, p := range v.Props {
			m[k] = guessTypeOfRawYAMLValue(p)
		}
		return &ObjectType{Props: m, StrictProps: false}
	case *RawYAMLArray:
		if len(v.Elems) == 0 {
			return &ArrayType{AnyType{}}
		}
		elem := guessTypeOfRawYAMLValue(v.Elems[0])
		for _, v := range v.Elems[1:] {
			elem = elem.Fuse(guessTypeOfRawYAMLValue(v))
		}
		return &ArrayType{elem}
	case *RawYAMLString:
		return guessTypeFromString(v.Value)
	default:
		panic("unreachable")
	}
}

func guessTypeFromString(s string) ExprType {
	if s == "true" || s == "false" {
		return BoolType{}
	}
	if s == "null" {
		return NullType{}
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return NumberType{}
	}
	return StringType{}
}
