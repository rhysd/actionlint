package actionlint

import (
	"fmt"
	"strconv"
	"strings"
)

type typedExpr struct {
	ty  ExprType
	pos Pos
}

// RuleExpression is a rule checker to check expression syntax in string values of workflow syntax.
// It checks syntax and semantics of the expressions including type checks and functions/contexts
// definitions. For more details see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type RuleExpression struct {
	RuleBase
	matrixTy     *ObjectType
	stepsTy      *ObjectType
	needsTy      *ObjectType
	workflow     *Workflow
	localActions *LocalActionsCache
}

// NewRuleExpression creates new RuleExpression instance.
func NewRuleExpression(cache *LocalActionsCache) *RuleExpression {
	return &RuleExpression{
		RuleBase:     RuleBase{name: "expression"},
		matrixTy:     nil,
		stepsTy:      nil,
		needsTy:      nil,
		workflow:     nil,
		localActions: cache,
	}
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RuleExpression) VisitWorkflowPre(n *Workflow) error {
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
	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children
func (rule *RuleExpression) VisitWorkflowPost(n *Workflow) error {
	rule.workflow = nil
	return nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RuleExpression) VisitJobPre(n *Job) error {
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

	if n.RunsOn != nil {
		for _, l := range n.RunsOn.Labels {
			rule.checkString(l)
		}
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

	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children
func (rule *RuleExpression) VisitJobPost(n *Job) error {
	// 'environment' and 'outputs' sections are evaluated after all steps are run
	if n.Environment != nil {
		rule.checkString(n.Environment.Name)
		rule.checkString(n.Environment.URL)
	}
	for _, output := range n.Outputs {
		rule.checkString(output.Value)
	}

	rule.matrixTy = nil
	rule.stepsTy = nil
	rule.needsTy = nil

	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RuleExpression) VisitStep(n *Step) error {
	rule.checkIfCondition(n.If)
	rule.checkString(n.Name)

	var spec *String
	switch e := n.Exec.(type) {
	case *ExecRun:
		rule.checkScriptString(e.Run)
		rule.checkString(e.Shell)
		rule.checkString(e.WorkingDirectory)
	case *ExecAction:
		rule.checkString(e.Uses)
		for n, i := range e.Inputs {
			if strings.HasPrefix(e.Uses.Value, "actions/github-script@") && n == "script" {
				rule.checkScriptString(i.Value)
			} else {
				rule.checkString(i.Value)
			}
		}
		rule.checkString(e.Entrypoint)
		rule.checkString(e.Args)
		spec = e.Uses
	}

	rule.checkEnv(n.Env)
	rule.checkBool(n.ContinueOnError)
	rule.checkFloat(n.TimeoutMinutes)

	if n.ID != nil {
		// Step ID is case insensitive
		id := strings.ToLower(n.ID.Value)
		rule.stepsTy.Props[id] = &ObjectType{
			Props: map[string]ExprType{
				"outputs":    rule.getActionOutputsType(spec),
				"conclusion": StringType{},
				"outcome":    StringType{},
			},
			StrictProps: true,
		}
	}

	return nil
}

func (rule *RuleExpression) getActionOutputsType(spec *String) *ObjectType {
	if spec == nil {
		return NewMapObjectType(StringType{}) // outputs.<output name>
	}

	if strings.HasPrefix(spec.Value, "./") {
		meta, err := rule.localActions.FindMetadata(spec.Value)
		if err != nil {
			rule.error(spec.Pos, err.Error())
			return NewMapObjectType(StringType{}) // outputs.<output name>
		}
		if meta == nil {
			return NewMapObjectType(StringType{}) // outputs.<output name>
		}

		return typeOfActionOutputs(meta)
	}

	// When the action run at this step is a popular action, we know what outputs are set by it.
	// Set the output names to `steps.{step_id}.outputs.{name}`.
	if meta, ok := PopularActions[spec.Value]; ok {
		return typeOfActionOutputs(meta)
	}

	return NewMapObjectType(StringType{}) // outputs.<output name>
}

func (rule *RuleExpression) checkOneExpression(s *String, what string) ExprType {
	// checkString is not available since it checks types for embedding values into a string
	if s == nil {
		return nil
	}
	ts := rule.checkExprsIn(s.Value, s.Pos, s.Quoted, false)
	if len(ts) != 1 {
		// This case should be unreachable since only one ${{ }} is included is checked by parser
		rule.errorf(s.Pos, "one ${{ }} expression should be included in %q value but got %d expressions", what, len(ts))
		return nil
	}
	return ts[0].ty
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
	case *ArrayType, AnyType:
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
		rule.errorf(pos, "type of expression at %q must be number but found type %s", what, ty.String())
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
		if ty, ok := rule.checkArrayExpression(cs.Expression, what).(*ArrayType); ok {
			rule.checkObjectTy(ty.Elem, cs.Expression.Pos, what)
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
	// This document is actually wrong. I confirmed that any strings without surrounding in ${{ }}
	// are evaluated.
	//
	// - run: echo 'run'
	//   if: '!false'
	// - run: echo 'not run'
	//   if: '!true'
	// - run: echo 'run'
	//   if: false || true
	// - run: echo 'run'
	//   if: true && true
	// - run: echo 'not run'
	//   if: true && false

	var condTy ExprType
	if strings.Contains(str.Value, "${{") && strings.Contains(str.Value, "}}") {
		if ts := rule.checkString(str); len(ts) == 1 {
			s := strings.TrimSpace(str.Value)
			if strings.HasPrefix(s, "${{") && strings.HasSuffix(s, "}}") {
				condTy = ts[0].ty
			}
		}
	} else {
		src := str.Value + "}}" // }} is necessary since lexer lexes it as end of tokens
		line, col := str.Pos.Line, str.Pos.Col

		p := NewExprParser()
		expr, err := p.Parse(NewExprLexer(src))
		if err != nil {
			rule.exprError(err, line, col)
			return
		}

		condTy = rule.checkSemanticsOfExprNode(expr, line, col, false)
	}

	if condTy != nil && !(BoolType{}).Assignable(condTy) {
		rule.errorf(str.Pos, "\"if\" condition should be type \"bool\" but got type %q", condTy.String())
	}
}

func (rule *RuleExpression) checkTemplateEvaluatedType(ts []typedExpr) {
	for _, t := range ts {
		switch t.ty.(type) {
		case *ObjectType, *ArrayType, NullType:
			rule.errorf(&t.pos, "object, array, and null values should not be evaluated in template with ${{ }} but evaluating the value of type %s", t.ty)
		}
	}
}

func (rule *RuleExpression) checkString(str *String) []typedExpr {
	if str == nil {
		return nil
	}
	ts := rule.checkExprsIn(str.Value, str.Pos, str.Quoted, false)
	rule.checkTemplateEvaluatedType(ts)
	return ts
}

func (rule *RuleExpression) checkScriptString(str *String) []typedExpr {
	if str == nil {
		return nil
	}
	ts := rule.checkExprsIn(str.Value, str.Pos, str.Quoted, true)
	rule.checkTemplateEvaluatedType(ts)
	return ts
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

func (rule *RuleExpression) checkExprsIn(s string, pos *Pos, quoted bool, checkUntrusted bool) []typedExpr {
	// TODO: Line number is not correct when the string contains newlines.

	line, col := pos.Line, pos.Col
	if quoted {
		col++ // when the string is quoted like 'foo' or "foo", column should be incremented
	}
	offset := 0
	ts := []typedExpr{}
	for {
		idx := strings.Index(s, "${{")
		if idx == -1 {
			break
		}

		start := idx + 3 // 3 means removing "${{"
		s = s[start:]
		offset += start
		col := col + offset

		ty, offsetAfter := rule.checkSemantics(s, line, col, checkUntrusted)
		if ty == nil || offsetAfter == 0 {
			return nil
		}
		ts = append(ts, typedExpr{ty, Pos{line, col - 3}})

		s = s[offsetAfter:]
		offset += offsetAfter
	}

	return ts
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
		rule.checkExprsIn(v.Value, v.Pos(), false, false)
	default:
		panic("unreachable")
	}
}

func (rule *RuleExpression) exprError(err *ExprError, lineBase, colBase int) {
	pos := convertExprLineColToPos(err.Line, err.Column, lineBase, colBase)
	rule.error(pos, err.Message)
}

func (rule *RuleExpression) checkSemanticsOfExprNode(expr ExprNode, line, col int, checkUntrusted bool) ExprType {
	c := NewExprSemanticsChecker(checkUntrusted)
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

	return ty
}

func (rule *RuleExpression) checkSemantics(src string, line, col int, checkUntrusted bool) (ExprType, int) {
	l := NewExprLexer(src)
	p := NewExprParser()
	expr, err := p.Parse(l)
	if err != nil {
		rule.exprError(err, line, col)
		return nil, l.Offset()
	}
	return rule.checkSemanticsOfExprNode(expr, line, col, checkUntrusted), l.Offset()
}

func (rule *RuleExpression) calcNeedsType(job *Job) *ObjectType {
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#needs-context
	o := NewStrictObjectType()
	rule.populateDependantNeedsTypes(o, job, job)
	return o
}

func (rule *RuleExpression) populateDependantNeedsTypes(out *ObjectType, job *Job, root *Job) {
	for _, id := range job.Needs {
		i := strings.ToLower(id.Value) // ID is case insensitive
		if i == root.ID.Value {
			continue // When cyclic dependency exists. This does not happen normally.
		}
		if _, ok := out.Props[i]; ok {
			continue // Already added
		}

		j, ok := rule.workflow.Jobs[i]
		if !ok {
			continue
		}

		outputs := NewStrictObjectType()
		for name := range j.Outputs {
			outputs.Props[name] = StringType{}
		}

		out.Props[i] = &ObjectType{
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
	ty := rule.checkObjectExpression(expr, "matrix")
	if ty == nil {
		return NewObjectType()
	}
	matTy, ok := ty.(*ObjectType)
	if !ok {
		return NewObjectType()
	}

	// Consider properties in include section elements since 'include' section adds matrix values
	incTy, ok := matTy.Props["include"]
	if ok {
		delete(matTy.Props, "include")
		if a, ok := incTy.(*ArrayType); ok {
			if o, ok := a.Elem.(*ObjectType); ok {
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
		if ty, ok := rule.checkOneExpression(m.Include.Expression, "include").(*ArrayType); ok {
			if ret, ok := o.Fuse(ty.Elem).(*ObjectType); ok {
				return ret
			}
		}
		return NewObjectType()
	}

	for _, combi := range m.Include.Combinations {
		if combi.Expression != nil {
			ty := rule.checkOneExpression(m.Include.Expression, "matrix combination at element of include section")
			if ty == nil {
				continue
			}
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
		if a, ok := rule.checkArrayExpression(r.Expression, "matrix row").(*ArrayType); ok {
			return a
		}
		return AnyType{}
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
		return &ObjectType{Props: m, StrictProps: true}
	case *RawYAMLArray:
		if len(v.Elems) == 0 {
			return &ArrayType{AnyType{}, false}
		}
		elem := guessTypeOfRawYAMLValue(v.Elems[0])
		for _, v := range v.Elems[1:] {
			elem = elem.Fuse(guessTypeOfRawYAMLValue(v))
		}
		return &ArrayType{elem, false}
	case *RawYAMLString:
		return guessTypeFromString(v.Value)
	default:
		panic("unreachable")
	}
}

func guessTypeFromString(s string) ExprType {
	// Note that keywords are case sensitive. TRUE, FALSE, NULL are invalid named value.
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

func convertExprLineColToPos(line, col, lineBase, colBase int) *Pos {
	// Line and column in ExprError are 1-based
	return &Pos{
		Line: line - 1 + lineBase,
		Col:  col - 1 + colBase,
	}
}

func typeOfActionOutputs(meta *ActionMetadata) *ObjectType {
	// Some action sets outputs dynamically. Such outputs are not defined in action.yml. actionlint
	// cannot check such outputs statically so it allows any props (#18)
	if meta.SkipOutputs {
		return NewObjectType()
	}
	ty := NewObjectType()
	for n := range meta.Outputs {
		ty.Props[n] = AnyType{}
	}
	ty.StrictProps = true
	return ty
}
