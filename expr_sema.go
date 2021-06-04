package actionlint

import (
	"fmt"
	"sort"
	"strconv"
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

// Types

// ExprType is interface for types of values in expression.
type ExprType interface {
	// String returns string representation of the type.
	String() string
	// Assignable returns if other type can be assignable to the type.
	Assignable(other ExprType) bool
	// Equals returns if the type is equal to the other type.
	Equals(other ExprType) bool
}

// AnyType represents type which can be any type. It also indicates that a value of the type cannot
// be type-checked since it's type cannot be known statically.
type AnyType struct{}

func (ty AnyType) String() string {
	return "any"
}

// Assignable returns if other type can be assignable to the type.
func (ty AnyType) Assignable(_ ExprType) bool {
	return true
}

// Equals returns if the type is equal to the other type.
func (ty AnyType) Equals(other ExprType) bool {
	return true
}

// NullType is type for null value.
type NullType struct{}

func (ty NullType) String() string {
	return "null"
}

// Assignable returns if other type can be assignable to the type.
func (ty NullType) Assignable(other ExprType) bool {
	switch other.(type) {
	case NullType, AnyType:
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty NullType) Equals(other ExprType) bool {
	_, ok := other.(NullType)
	return ok
}

// NumberType is type for number values such as integer or float.
type NumberType struct{}

func (ty NumberType) String() string {
	return "number"
}

// Equals returns if the type is equal to the other type.
func (ty NumberType) Equals(other ExprType) bool {
	_, ok := other.(NumberType)
	return ok
}

// Assignable returns if other type can be assignable to the type.
func (ty NumberType) Assignable(other ExprType) bool {
	// TODO: Is string of numbers corced into number?
	switch other.(type) {
	case NumberType, AnyType:
		return true
	default:
		return false
	}
}

// BoolType is type for boolean values.
type BoolType struct{}

func (ty BoolType) String() string {
	return "bool"
}

// Assignable returns if other type can be assignable to the type.
func (ty BoolType) Assignable(other ExprType) bool {
	// TODO: Is numbers corced into bool?
	switch other.(type) {
	case BoolType, AnyType:
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty BoolType) Equals(other ExprType) bool {
	_, ok := other.(BoolType)
	return ok
}

// StringType is type for string values.
type StringType struct{}

func (ty StringType) String() string {
	return "string"
}

// Assignable returns if other type can be assignable to the type.
func (ty StringType) Assignable(other ExprType) bool {
	// TODO: Is numbers corced into string?
	switch other.(type) {
	case StringType, AnyType:
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty StringType) Equals(other ExprType) bool {
	_, ok := other.(StringType)
	return ok
}

// ObjectType is type for objects, which can hold key-values.
type ObjectType struct {
	// Props is map from properties name to their type.
	Props map[string]ExprType
	// StrictProps is flag to check if the properties should be checked strictly. When this flag
	// is set to true, it means that other than properties defined in Props field are not permitted
	// and will cause type error. When this flag is set to false, accessing to unknown properties
	// does not cause type error and will be deducted to any type.
	StrictProps bool
}

// NewObjectType creates new ObjectType instance.
func NewObjectType() *ObjectType {
	return &ObjectType{map[string]ExprType{}, false}
}

func (ty *ObjectType) String() string {
	len := len(ty.Props)
	if len == 0 {
		return "object"
	}
	ps := make([]string, 0, len)
	for n, t := range ty.Props {
		ps = append(ps, fmt.Sprintf("%s: %s", n, t.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(ps, "; "))
}

// Assignable returns if other type can be assignable to the type.
func (ty *ObjectType) Assignable(other ExprType) bool {
	switch other := other.(type) {
	case AnyType:
		return true
	case *ObjectType:
		for n, p1 := range ty.Props {
			if p2, ok := other.Props[n]; ok && !p1.Assignable(p2) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty *ObjectType) Equals(other ExprType) bool {
	panic("unimplemented")
}

// ArrayType is type for arrays.
type ArrayType struct {
	// Elem is type of element of the array.
	Elem ExprType
}

func (ty *ArrayType) String() string {
	return fmt.Sprintf("array<%s>", ty.Elem.String())
}

// Equals returns if the type is equal to the other type.
func (ty *ArrayType) Equals(other ExprType) bool {
	panic("unimplemented")
}

// Assignable returns if other type can be assignable to the type.
func (ty *ArrayType) Assignable(other ExprType) bool {
	// Note: ArrayType and ArrayDerefType are compatible
	switch other := other.(type) {
	case AnyType:
		return true
	case *ArrayType:
		return ty.Elem.Assignable(other.Elem)
	case *ArrayDerefType:
		return ty.Elem.Assignable(other.Elem)
	default:
		return false
	}
}

// ArrayDerefType is a type for array element dereference with '.*'. It is distinguished from ArrayType
// for type checker.
// For example, when type of 'a' is '{foo: {bar: int}[]}', 'a.*' is type of 'array_deref<{bar: int}>
// ' and 'a.*.b' is type of 'array_deref<int>'.
type ArrayDerefType struct {
	// Elem is type of element of dereferenced (and filtered) array.
	Elem ExprType
}

func (ty *ArrayDerefType) String() string {
	return fmt.Sprintf("array<%s>", ty.Elem.String())
}

// Assignable returns if other type can be assignable to the type.
func (ty *ArrayDerefType) Assignable(other ExprType) bool {
	// Note: ArrayType and ArrayDerefType are compatible
	switch other := other.(type) {
	case AnyType:
		return true
	case *ArrayType:
		return ty.Elem.Assignable(other.Elem)
	case *ArrayDerefType:
		return ty.Elem.Assignable(other.Elem)
	default:
		return false
	}
}

// Equals returns if the type is equal to the other type.
func (ty *ArrayDerefType) Equals(other ExprType) bool {
	panic("unimplemented")
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

// BuiltinFuncSignatures is a set of all builtin function signatures.
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
	"startsWith": {
		{
			Name: "startsWith",
			Ret:  BoolType{},
			Params: []ExprType{
				StringType{},
				StringType{},
			},
		},
	},
	"endsWith": {
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
	},
	"toJson": {{
		Name: "toJson",
		Ret:  StringType{},
		Params: []ExprType{
			AnyType{},
		},
	}},
	"fromJson": {{
		Name: "fromJson",
		Ret:  AnyType{},
		Params: []ExprType{
			StringType{},
		},
	}},
	"hashFiles": {{
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

// GlobalVariableType is type of global variable.
type GlobalVariableType struct {
	// Name is a name of the global variable.
	Name string
	// Type is a type of the global variable.
	Type ExprType
}

// BuiltinGlobalVariableTypes defines types of all global variables. All context variables are
// documented at https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
var BuiltinGlobalVariableTypes = map[string]*GlobalVariableType{
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#github-context
	"github": {
		Name: "github",
		Type: &ObjectType{
			Props: map[string]ExprType{
				"action":           StringType{},
				"action_path":      StringType{},
				"actor":            StringType{},
				"base_ref":         StringType{},
				"event":            NewObjectType(),
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
			},
			StrictProps: true,
		},
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#env-context
	"env": {
		Name: "env",
		Type: NewObjectType(),
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#job-context
	"job": {
		Name: "job",
		Type: &ObjectType{
			Props: map[string]ExprType{
				"container": &ObjectType{
					Props: map[string]ExprType{
						"id":      StringType{},
						"network": StringType{},
					},
				},
				"services": NewObjectType(),
				"status":   StringType{},
			},
			StrictProps: true,
		},
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#steps-context
	"steps": {
		Name: "steps",
		Type: NewObjectType(),
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#runner-context
	"runner": {
		Name: "runner",
		Type: &ObjectType{
			Props: map[string]ExprType{
				"os":         StringType{},
				"temp":       StringType{},
				"tool_cache": StringType{},
			},
			StrictProps: true,
		},
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	"secrets": {
		Name: "secrets",
		Type: NewObjectType(),
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	"strategy": {
		Name: "strategy",
		Type: &ObjectType{
			Props: map[string]ExprType{
				"fail-fast":    BoolType{},
				"job-index":    NumberType{},
				"job-total":    NumberType{},
				"max-parallel": NumberType{},
			},
		},
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	"matrix": {
		Name: "matrix",
		Type: NewObjectType(),
	},
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#needs-context
	"needs": {
		Name: "needs",
		Type: NewObjectType(),
	},
}

// Semantics checker

// ExprSemanticsChecker is a semantics checker for expression syntax. It checks types of values
// in given expression syntax tree. It additionally checks other semantics like arguments of
// format() built-in function. To know the details of the syntax, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
type ExprSemanticsChecker struct {
	funcs      map[string][]*FuncSignature
	vars       map[string]*GlobalVariableType
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

func (sema *ExprSemanticsChecker) ensureCopyVars() {
	if sema.varsCopied {
		return
	}

	// Make shallow copy of current variables map not to pollute global variable
	copied := make(map[string]*GlobalVariableType, len(sema.vars))
	for k, v := range sema.vars {
		copied[k] = v
	}
	sema.vars = copied
	sema.varsCopied = true
}

// UpdateMatrix updates matrix object to given object type. Since matrix values change according to
// 'matrix' section of job configuration, the type needs to be updated.
func (sema *ExprSemanticsChecker) UpdateMatrix(ty *ObjectType) {
	sema.ensureCopyVars()
	sema.vars["matrix"] = &GlobalVariableType{
		Name: "matrix",
		Type: ty,
	}
}

func (sema *ExprSemanticsChecker) UpdateSteps(ty *ObjectType) {
	sema.ensureCopyVars()
	sema.vars["steps"] = &GlobalVariableType{
		Name: "steps",
		Type: ty,
	}
}

func (sema *ExprSemanticsChecker) checkVariable(n *VariableNode) ExprType {
	global, ok := sema.vars[n.Name]
	if !ok {
		qs := make([]string, 0, len(sema.vars))
		for n := range sema.vars {
			qs = append(qs, strconv.Quote(n))
		}
		sort.Strings(qs)
		sema.errorf(n, "undefined variable %q. available variables are %s", n.Name, strings.Join(qs, ", "))
		return AnyType{}
	}

	return global.Type
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
	case *ArrayDerefType:
		switch et := ty.Elem.(type) {
		case AnyType:
			// When element type is any, map the any type to any. Reuse `ty`
			return ty
		case *ObjectType:
			// Map element type of delererenced array
			var elem ExprType = AnyType{}
			if t, ok := et.Props[n.Property]; ok {
				elem = t
			}
			if et.StrictProps {
				sema.errorf(n, "property %q is not defined in object type %s", n.Property, et.String())
			}
			return &ArrayDerefType{Elem: elem}
		default:
			sema.errorf(
				n,
				"object proprety filter %q of array element dereference must be type of object but got %q",
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
		return &ArrayDerefType{Elem: AnyType{}}
	case *ArrayType:
		return &ArrayDerefType{Elem: ty.Elem}
	case *ArrayDerefType:
		return &ArrayDerefType{Elem: ty.Elem}
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
	case *ArrayDerefType:
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
	sigs, ok := sema.funcs[n.Callee]
	if !ok {
		qs := make([]string, 0, len(sema.funcs))
		for n := range sema.funcs {
			qs = append(qs, strconv.Quote(n))
		}
		sort.Strings(qs)
		sema.errorf(n, "undefined function %q. available functions are %s", n.Callee, strings.Join(qs, ", "))
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
