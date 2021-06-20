package actionlint

import (
	"fmt"
	"strings"
)

func ordinal(i int) string {
	suffix := "th"
	switch i % 10 {
	case 1:
		if i%100 != 11 {
			suffix = "st"
		}
	case 2:
		if i%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if i%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", i, suffix)
}

// Functions

// FuncSignature is a signature of function, which holds return and arguments types.
type FuncSignature struct {
	// Name is a name of the function.
	Name string
	// Ret is a return type of the function.
	Ret ExprType
	// Params is a list of parameter types of the function. The final element of this list might
	// be repeated as variable length arguments.
	Params []ExprType
	// VariableLengthParams is a flag to handle variable length parameters. When this flag is set to
	// true, it means that the last type of params might be specified multiple times (including zero
	// times). Setting true implies length of Params is more than 0.
	VariableLengthParams bool
}

func (sig *FuncSignature) String() string {
	ts := make([]string, 0, len(sig.Params))
	for _, p := range sig.Params {
		ts = append(ts, p.String())
	}
	elip := ""
	if sig.VariableLengthParams {
		elip = "..."
	}
	return fmt.Sprintf("%s(%s%s) -> %s", sig.Name, strings.Join(ts, ", "), elip, sig.Ret.String())
}

// BuiltinFuncSignatures is a set of all builtin function signatures. All function names are in
// lower case because function names are compared in case insensitive.
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#functions
var BuiltinFuncSignatures = map[string][]*FuncSignature{
	"contains": {
		{
			Name: "contains",
			Ret:  BoolType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
		{
			Name: "contains",
			Ret:  BoolType{},
			Params: []ExprType{
				&ArrayType{Elem: AnyType{}},
				AnyType{},
			},
		},
	},
	"startswith": {
		{
			Name: "startsWith",
			Ret:  BoolType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
	},
	"endswith": {
		{
			Name: "endsWith",
			Ret:  BoolType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
	},
	"format": {
		{
			Name: "format",
			Ret:  StringType{},
			Params: []ExprType{
				StringType{},
				AnyType{}, // variable length
			},
			VariableLengthParams: true,
		},
	},
	"join": {
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				&ArrayType{Elem: StringType{}},
				StringType{},
			},
		},
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
		// When the second parameter is omitted, values are concatenated with ','.
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				&ArrayType{Elem: StringType{}},
			},
		},
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				StringType{},
			},
		},
	},
	"tojson": {{
		Name: "toJSON",
		Ret:  StringType{},
		Params: []ExprType{
			AnyType{},
		},
	}},
	"fromjson": {{
		Name: "fromJSON",
		Ret:  AnyType{},
		Params: []ExprType{
			StringType{},
		},
	}},
	"hashfiles": {{
		Name: "hashFiles",
		Ret:  StringType{},
		Params: []ExprType{
			StringType{},
		},
		VariableLengthParams: true,
	}},
	"success": {{
		Name:   "success",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
	"always": {{
		Name:   "always",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
	"cancelled": {{
		Name:   "cancelled",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
	"failure": {{
		Name:   "failure",
		Ret:    BoolType{},
		Params: []ExprType{},
	}},
}

// Global variables

// BuiltinGlobalVariableTypes defines types of all global variables. All context variables are
// documented at https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
var BuiltinGlobalVariableTypes = map[string]ExprType{
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#github-context
	"github": &ObjectType{
		Props: map[string]ExprType{
			"action":           StringType{},
			"action_path":      StringType{},
			"actor":            StringType{},
			"base_ref":         StringType{},
			"event":            NewObjectType(), // Note: Stricter type check for this payload would be possible
			"event_name":       StringType{},
			"event_path":       StringType{},
			"head_ref":         StringType{},
			"job":              StringType{},
			"ref":              StringType{},
			"repository":       StringType{},
			"repository_owner": StringType{},
			"run_id":           StringType{},
			"run_number":       StringType{},
			"sha":              StringType{},
			"token":            StringType{},
			"workflow":         StringType{},
			"workspace":        StringType{},
			// These are not documented but actually exist
			"action_ref":        StringType{},
			"action_repository": StringType{},
			"api_url":           StringType{},
			"env":               StringType{},
			"graphql_url":       StringType{},
			"path":              StringType{},
			"repositoryurl":     StringType{}, // repositoryUrl
			"retention_days":    NumberType{},
			"server_url":        StringType{},
		},
		StrictProps: true,
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#env-context
	"env": NewObjectType(),
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#job-context
	"job": &ObjectType{
		Props: map[string]ExprType{
			"container": &ObjectType{
				Props: map[string]ExprType{
					"id":      StringType{},
					"network": StringType{},
				},
				StrictProps: true,
			},
			"services": NewObjectType(),
			"status":   StringType{},
		},
		StrictProps: true,
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#steps-context
	"steps": NewStrictObjectType(),
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#runner-context
	"runner": &ObjectType{
		Props: map[string]ExprType{
			"os":         StringType{},
			"temp":       StringType{},
			"tool_cache": StringType{},
			// These are not documented but actually exist
			"workspace": StringType{},
		},
		StrictProps: true,
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	"secrets": NewObjectType(),
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	"strategy": &ObjectType{
		Props: map[string]ExprType{
			"fail-fast":    BoolType{},
			"job-index":    NumberType{},
			"job-total":    NumberType{},
			"max-parallel": NumberType{},
		},
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	"matrix": NewStrictObjectType(),
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#needs-context
	"needs": NewStrictObjectType(),
}

// Semantics checker

// ExprSemanticsChecker is a semantics checker for expression syntax. It checks types of values
// in given expression syntax tree. It additionally checks other semantics like arguments of
// format() built-in function. To know the details of the syntax, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
type ExprSemanticsChecker struct {
	funcs      map[string][]*FuncSignature
	vars       map[string]ExprType
	errs       []*ExprError
	varsCopied bool
}

// NewExprSemanticsChecker creates new ExprSemanticsChecker instance.
func NewExprSemanticsChecker() *ExprSemanticsChecker {
	return &ExprSemanticsChecker{
		funcs:      BuiltinFuncSignatures,
		vars:       BuiltinGlobalVariableTypes,
		varsCopied: false,
	}
}

func errorAtExpr(e ExprNode, msg string) *ExprError {
	t := e.Token()
	return &ExprError{
		Message: msg,
		Offset:  t.Offset,
		Line:    t.Line,
		Column:  t.Column,
	}
}

func errorfAtExpr(e ExprNode, format string, args ...interface{}) *ExprError {
	return errorAtExpr(e, fmt.Sprintf(format, args...))
}

func (sema *ExprSemanticsChecker) errorf(e ExprNode, format string, args ...interface{}) {
	sema.errs = append(sema.errs, errorfAtExpr(e, format, args...))
}

func (sema *ExprSemanticsChecker) ensureVarsCopied() {
	if sema.varsCopied {
		return
	}

	// Make shallow copy of current variables map not to pollute global variable
	copied := make(map[string]ExprType, len(sema.vars))
	for k, v := range sema.vars {
		copied[k] = v
	}
	sema.vars = copied
	sema.varsCopied = true
}

// UpdateMatrix updates matrix object to given object type. Since matrix values change according to
// 'matrix' section of job configuration, the type needs to be updated.
func (sema *ExprSemanticsChecker) UpdateMatrix(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["matrix"] = ty
}

// UpdateSteps updates 'steps' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateSteps(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["steps"] = ty
}

// UpdateNeeds updates 'needs' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateNeeds(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["needs"] = ty
}

func (sema *ExprSemanticsChecker) checkVariable(n *VariableNode) ExprType {
	v, ok := sema.vars[n.Name]
	if !ok {
		ss := make([]string, 0, len(sema.vars))
		for n := range sema.vars {
			ss = append(ss, n)
		}
		sema.errorf(n, "undefined variable %q. available variables are %s", n.Token().Value, sortedQuotes(ss))
		return AnyType{}
	}

	return v
}

func (sema *ExprSemanticsChecker) checkObjectDeref(n *ObjectDerefNode) ExprType {
	switch ty := sema.check(n.Receiver).(type) {
	case AnyType:
		return AnyType{}
	case *ObjectType:
		if t, ok := ty.Props[n.Property]; ok {
			return t
		}
		if ty.StrictProps {
			sema.errorf(n, "property %q is not defined in object type %s", n.Property, ty.String())
		}
		return AnyType{}
	case *ArrayType:
		if !ty.Deref {
			sema.errorf(n, "receiver of object dereference %q must be type of object but got %q", n.Property, ty.String())
			return AnyType{}
		}
		switch et := ty.Elem.(type) {
		case AnyType:
			// When element type is any, map the any type to any. Reuse `ty`
			return ty
		case *ObjectType:
			// Map element type of delererenced array
			var elem ExprType = AnyType{}
			if t, ok := et.Props[n.Property]; ok {
				elem = t
			} else if et.StrictProps {
				sema.errorf(n, "property %q is not defined in object type %s as element of filtered array", n.Property, et.String())
			}
			return &ArrayType{elem, true}
		default:
			sema.errorf(
				n,
				"object property filter %q of array element dereference must be type of object but got %q",
				n.Property,
				ty.Elem.String(),
			)
			return AnyType{}
		}
	default:
		sema.errorf(n, "receiver of object dereference %q must be type of object but got %q", n.Property, ty.String())
		return AnyType{}
	}
}

func (sema *ExprSemanticsChecker) checkArrayDeref(n *ArrayDerefNode) ExprType {
	switch ty := sema.check(n.Receiver).(type) {
	case AnyType:
		return &ArrayType{AnyType{}, true}
	case *ArrayType:
		ty.Deref = true
		return ty
	default:
		sema.errorf(n, "receiver of array element dereference must be type of array but got %q", ty.String())
		return AnyType{}
	}
}

func (sema *ExprSemanticsChecker) checkIndexAccess(n *IndexAccessNode) ExprType {
	idx := sema.check(n.Index)
	switch ty := sema.check(n.Operand).(type) {
	case AnyType:
		return AnyType{}
	case *ArrayType:
		switch idx.(type) {
		case AnyType, NumberType:
			return ty.Elem
		default:
			sema.errorf(n.Index, "index access of array must be type of number but got %q", idx.String())
			return AnyType{}
		}
	case *ObjectType:
		switch idx.(type) {
		case AnyType:
			return AnyType{}
		case StringType:
			// Index access with string literal like foo['bar']
			if lit, ok := n.Index.(*StringNode); ok {
				if prop, ok := ty.Props[lit.Value]; ok {
					return prop
				}
				if ty.StrictProps {
					sema.errorf(n, "property %q is not defined in object type %s", lit.Value, ty.String())
				}
			}
			return AnyType{}
		default:
			sema.errorf(n.Index, "property access of object must be type of string but got %q", idx.String())
			return AnyType{}
		}
	default:
		sema.errorf(n, "index access operand must be type of object or array but got %q", ty.String())
		return AnyType{}
	}
}

func checkFuncSignature(n *FuncCallNode, sig *FuncSignature, args []ExprType) *ExprError {
	lp, la := len(sig.Params), len(args)
	if sig.VariableLengthParams && (lp > la) || !sig.VariableLengthParams && lp != la {
		atLeast := ""
		if sig.VariableLengthParams {
			atLeast = "at least "
		}
		return errorfAtExpr(
			n,
			"number of arguments is wrong. function %q takes %s%d parameters but %d arguments are provided",
			sig.String(),
			atLeast,
			lp,
			la,
		)
	}

	for i := 0; i < len(sig.Params); i++ {
		p, a := sig.Params[i], args[i]
		if !p.Assignable(a) {
			return errorfAtExpr(
				n.Args[i],
				"%s argument of function call is not assignable. %q cannot be assigned to %q. called function type is %q",
				ordinal(i+1),
				a.String(),
				p.String(),
				sig.String(),
			)
		}
	}

	// Note: Unlike many languages, this check does not allow 0 argument for the variable length
	// parameter since it is useful for checking hashFiles() and format().
	if sig.VariableLengthParams {
		rest := args[lp:]
		p := sig.Params[lp-1]
		for i, a := range rest {
			if !p.Assignable(a) {
				return errorfAtExpr(
					n.Args[lp+i],
					"%s argument of function call is not assignable. %q cannot be assigned to %q. called function type is %q",
					ordinal(lp+i+1),
					a.String(),
					p.String(),
					sig.String(),
				)
			}
		}
	}

	return nil
}

func (sema *ExprSemanticsChecker) checkBuiltinFunctionCall(n *FuncCallNode, sig *FuncSignature) {
	switch n.Callee {
	case "format":
		lit, ok := n.Args[0].(*StringNode)
		if !ok {
			return
		}

		// Count number of placeholders in format string
		c := 0
		for i := 0; ; i++ {
			p := fmt.Sprintf("{%d}", i)
			if !strings.Contains(lit.Value, p) {
				break
			}
			c++
		}

		// -1 means removing first format string argument
		if len(n.Args)-1 != c {
			sema.errorf(
				n,
				"format string %q contains %d placeholders but %d arguments are given to format",
				lit.Value,
				c,
				len(n.Args)-1,
			)
		}
	}
}

func (sema *ExprSemanticsChecker) checkFuncCall(n *FuncCallNode) ExprType {
	// Check function name in case insensitive. For example, toJson and toJSON are the same function.
	callee := strings.ToLower(n.Callee)
	sigs, ok := sema.funcs[callee]
	if !ok {
		ss := make([]string, 0, len(sema.funcs))
		for n := range sema.funcs {
			ss = append(ss, n)
		}
		sema.errorf(n, "undefined function %q. available functions are %s", n.Callee, sortedQuotes(ss))
		return AnyType{}
	}

	tys := make([]ExprType, 0, len(n.Args))
	for _, a := range n.Args {
		tys = append(tys, sema.check(a))
	}

	// Check all overloads
	errs := []*ExprError{}
	for _, sig := range sigs {
		err := checkFuncSignature(n, sig, tys)
		if err == nil {
			// When one of overload pass type check, overload was resolved correctly
			sema.checkBuiltinFunctionCall(n, sig)
			return sig.Ret
		}
		errs = append(errs, err)
	}

	// All candidates failed
	sema.errs = append(sema.errs, errs...)

	return AnyType{}
}

func (sema *ExprSemanticsChecker) checkNotOp(n *NotOpNode) ExprType {
	ty := sema.check(n.Operand)
	if !(BoolType{}).Assignable(ty) {
		sema.errorf(n, "type of operand of ! operator %q is not assignable to type \"bool\"", ty.String())
	}
	return BoolType{}
}

func (sema *ExprSemanticsChecker) checkCompareOp(n *CompareOpNode) ExprType {
	sema.check(n.Left)
	sema.check(n.Right)
	// Note: Comparing values is very loose. Any value can be compared with any value without an
	// error.
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#operators
	return BoolType{}
}

func (sema *ExprSemanticsChecker) checkLogicalOp(n *LogicalOpNode) ExprType {
	lty := sema.check(n.Left)
	rty := sema.check(n.Right)
	if !(BoolType{}).Assignable(lty) {
		sema.errorf(n, "type of left operand of %s operator %q is not assignable to type \"bool\"", n.Kind.String(), lty.String())
	}
	if !(BoolType{}).Assignable(rty) {
		sema.errorf(n, "type of right operand of %s operator %q is not assignable to type \"bool\"", n.Kind.String(), rty.String())
	}
	return BoolType{}
}

func (sema *ExprSemanticsChecker) check(expr ExprNode) ExprType {
	switch e := expr.(type) {
	case *VariableNode:
		return sema.checkVariable(e)
	case *NullNode:
		return NullType{}
	case *BoolNode:
		return BoolType{}
	case *StringNode:
		return StringType{}
	case *IntNode, *FloatNode:
		return NumberType{}
	case *ObjectDerefNode:
		return sema.checkObjectDeref(e)
	case *ArrayDerefNode:
		return sema.checkArrayDeref(e)
	case *IndexAccessNode:
		return sema.checkIndexAccess(e)
	case *FuncCallNode:
		return sema.checkFuncCall(e)
	case *NotOpNode:
		return sema.checkNotOp(e)
	case *CompareOpNode:
		return sema.checkCompareOp(e)
	case *LogicalOpNode:
		return sema.checkLogicalOp(e)
	default:
		panic("unreachable")
	}
}

// Check checks sematics of given expression syntax tree. It returns the type of the expression as
// the first return value when the check was successfully done. And it returns all errors found
// while checking the expression as the second return value.
func (sema *ExprSemanticsChecker) Check(expr ExprNode) (ExprType, []*ExprError) {
	sema.errs = []*ExprError{}
	ty := sema.check(expr)
	return ty, sema.errs
}
