package actionlint

import (
	"encoding/json"
	"fmt"
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

// parseFormatFuncSpecifiers parses the format string passed to `format()` calls.
// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/evaluate-expressions-in-workflows-and-actions#format
func parseFormatFuncSpecifiers(f string, n int) map[int]struct{} {
	ret := make(map[int]struct{}, n)
	start := -1
	for i, r := range f {
		if r == '{' {
			if start == i {
				start = -1 // When the '{' is escaped like '{{'
			} else {
				start = i + 1 // `+ 1` because `i` points char '{'
			}
		} else if start >= 0 {
			if '0' <= r && r <= '9' {
				continue
			}
			if r == '}' && start < i {
				i, _ := strconv.Atoi(f[start:i])
				ret[i] = struct{}{}
			}
			start = -1 // Done
		}
	}
	return ret
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
// https://docs.github.com/en/actions/learn-github-actions/expressions#functions
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
		// When the second parameter is omitted, values are concatenated with ','.
		{
			Name: "join",
			Ret:  StringType{},
			Params: []ExprType{
				&ArrayType{Elem: StringType{}},
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
// documented at https://docs.github.com/en/actions/learn-github-actions/contexts
var BuiltinGlobalVariableTypes = map[string]ExprType{
	// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/accessing-contextual-information-about-workflow-runs#github-context
	"github": NewStrictObjectType(map[string]ExprType{
		"action":                    StringType{},
		"action_path":               StringType{}, // Note: Composite actions only
		"action_ref":                StringType{},
		"action_repository":         StringType{},
		"action_status":             StringType{}, // Note: Composite actions only
		"actor":                     StringType{},
		"actor_id":                  StringType{},
		"api_url":                   StringType{},
		"artifact_cache_size_limit": NumberType{}, // Note: Undocumented
		"base_ref":                  StringType{},
		"env":                       StringType{},
		"event":                     NewEmptyObjectType(), // Note: Stricter type check for this payload would be possible
		"event_name":                StringType{},
		"event_path":                StringType{},
		"graphql_url":               StringType{},
		"head_ref":                  StringType{},
		"job":                       StringType{},
		"output":                    StringType{}, // Note: Undocumented
		"path":                      StringType{},
		"ref":                       StringType{},
		"ref_name":                  StringType{},
		"ref_protected":             BoolType{},
		"ref_type":                  StringType{},
		"repository":                StringType{},
		"repository_id":             StringType{},
		"repository_owner":          StringType{},
		"repository_owner_id":       StringType{},
		"repository_visibility":     StringType{}, // Note: Undocumented
		"repositoryurl":             StringType{}, // repositoryUrl
		"retention_days":            NumberType{},
		"run_attempt":               StringType{},
		"run_id":                    StringType{},
		"run_number":                StringType{},
		"secret_source":             StringType{},
		"server_url":                StringType{},
		"sha":                       StringType{},
		"state":                     StringType{}, // Note: Undocumented
		"step_summary":              StringType{}, // Note: Undocumented
		"token":                     StringType{},
		"triggering_actor":          StringType{},
		"workflow":                  StringType{},
		"workflow_ref":              StringType{},
		"workflow_sha":              StringType{},
		"workspace":                 StringType{},
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#env-context
	"env": NewMapObjectType(StringType{}), // env.<env_name>
	// https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
	"job": NewStrictObjectType(map[string]ExprType{
		"container": NewStrictObjectType(map[string]ExprType{
			"id":      StringType{},
			"network": StringType{},
		}),
		"services": NewMapObjectType(
			NewStrictObjectType(map[string]ExprType{
				"id":      StringType{}, // job.services.<service id>.id
				"network": StringType{},
				"ports":   NewMapObjectType(StringType{}),
			}),
		),
		"status": StringType{},
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
	"steps": NewEmptyStrictObjectType(), // This value will be updated contextually
	// https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
	"runner": NewStrictObjectType(map[string]ExprType{
		"name":        StringType{},
		"os":          StringType{},
		"arch":        StringType{},
		"temp":        StringType{},
		"tool_cache":  StringType{},
		"debug":       StringType{},
		"environment": StringType{}, // https://github.com/github/docs/issues/32443
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#secrets-context
	"secrets": NewMapObjectType(StringType{}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#strategy-context
	"strategy": NewObjectType(map[string]ExprType{
		"fail-fast":    BoolType{},
		"job-index":    NumberType{},
		"job-total":    NumberType{},
		"max-parallel": NumberType{},
	}),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#matrix-context
	"matrix": NewEmptyStrictObjectType(), // This value will be updated contextually
	// https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context
	"needs": NewEmptyStrictObjectType(), // This value will be updated contextually
	// https://docs.github.com/en/actions/learn-github-actions/contexts#inputs-context
	// https://docs.github.com/en/actions/learn-github-actions/reusing-workflows
	"inputs": NewEmptyStrictObjectType(),
	// https://docs.github.com/en/actions/learn-github-actions/contexts#vars-context
	"vars": NewMapObjectType(StringType{}), // vars.<var_name>
}

// Semantics checker

// ExprSemanticsChecker is a semantics checker for expression syntax. It checks types of values
// in given expression syntax tree. It additionally checks other semantics like arguments of
// format() built-in function. To know the details of the syntax, see
//
// - https://docs.github.com/en/actions/learn-github-actions/contexts
// - https://docs.github.com/en/actions/learn-github-actions/expressions
type ExprSemanticsChecker struct {
	funcs                 map[string][]*FuncSignature
	vars                  map[string]ExprType
	errs                  []*ExprError
	varsCopied            bool
	githubVarCopied       bool
	untrusted             *UntrustedInputChecker
	availableContexts     []string
	availableSpecialFuncs []string
	configVars            []string
}

// NewExprSemanticsChecker creates new ExprSemanticsChecker instance. When checkUntrustedInput is
// set to true, the checker will make use of possibly untrusted inputs error.
func NewExprSemanticsChecker(checkUntrustedInput bool, configVars []string) *ExprSemanticsChecker {
	c := &ExprSemanticsChecker{
		funcs:           BuiltinFuncSignatures,
		vars:            BuiltinGlobalVariableTypes,
		varsCopied:      false,
		githubVarCopied: false,
		configVars:      configVars,
	}
	if checkUntrustedInput {
		c.untrusted = NewUntrustedInputChecker(BuiltinUntrustedInputs)
	}
	return c
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

func (sema *ExprSemanticsChecker) ensureGithubVarCopied() {
	if sema.githubVarCopied {
		return
	}
	sema.ensureVarsCopied()

	sema.vars["github"] = sema.vars["github"].DeepCopy()
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

// UpdateSecrets updates 'secrets' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateSecrets(ty *ObjectType) {
	sema.ensureVarsCopied()

	// Merges automatically supplied secrets with manually defined secrets.
	// ACTIONS_STEP_DEBUG and ACTIONS_RUNNER_DEBUG seem supplied from caller of the workflow (#130)
	copied := NewStrictObjectType(map[string]ExprType{
		"github_token":         StringType{},
		"actions_step_debug":   StringType{},
		"actions_runner_debug": StringType{},
	})
	for n, v := range ty.Props {
		copied.Props[n] = v
	}
	sema.vars["secrets"] = copied
}

// UpdateInputs updates 'inputs' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateInputs(ty *ObjectType) {
	sema.ensureVarsCopied()
	o := sema.vars["inputs"].(*ObjectType)
	if len(o.Props) == 0 && o.IsStrict() {
		sema.vars["inputs"] = ty
		return
	}
	// When both `workflow_call` and `workflow_dispatch` are the triggers of the workflow, `inputs` context can be used
	// by both events. To cover both cases, merge `inputs` contexts into one object type. (#263)
	sema.vars["inputs"] = o.Merge(ty)
}

// UpdateDispatchInputs updates 'github.event.inputs' and 'inputs' objects to given object type.
// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_dispatch
func (sema *ExprSemanticsChecker) UpdateDispatchInputs(ty *ObjectType) {
	sema.UpdateInputs(ty)

	// Update `github.event.inputs`.
	// Unlike `inputs.*`, type of `github.event.inputs.*` is always string unlike `inputs.*`. We need
	// to create a new type from `ty` (e.g. {foo: boolean, bar: number} -> {foo: string, bar: string})

	p := make(map[string]ExprType, len(ty.Props))
	for n := range ty.Props {
		p[n] = StringType{}
	}
	ty = NewStrictObjectType(p)

	sema.ensureGithubVarCopied()
	sema.vars["github"].(*ObjectType).Props["event"].(*ObjectType).Props["inputs"] = ty
}

// UpdateJobs updates 'jobs' context object to given object type.
func (sema *ExprSemanticsChecker) UpdateJobs(ty *ObjectType) {
	sema.ensureVarsCopied()
	sema.vars["jobs"] = ty
}

// SetContextAvailability sets available context names while semantics checks. Some contexts limit
// where they can be used.
// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
//
// Elements of 'avail' parameter must be in lower case to check context names in case-insensitive.
//
// If this method is not called before checks, ExprSemanticsChecker considers any contexts are
// available by default.
// Available contexts for workflow keys can be obtained from actionlint.ContextAvailability.
func (sema *ExprSemanticsChecker) SetContextAvailability(avail []string) {
	sema.availableContexts = avail
}

func (sema *ExprSemanticsChecker) checkAvailableContext(n *VariableNode) {
	ctx := strings.ToLower(n.Name)
	for _, c := range sema.availableContexts {
		if c == ctx {
			return
		}
	}

	var notes string
	switch len(sema.availableContexts) {
	case 0:
		notes = "no context is available here"
	case 1:
		notes = "available context is " + quotes(sema.availableContexts)
	default:
		notes = "available contexts are " + quotes(sema.availableContexts)
	}
	sema.errorf(
		n,
		"context %q is not allowed here. %s. see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details",
		n.Name,
		notes,
	)
}

// SetSpecialFunctionAvailability sets names of available special functions while semantics checks.
// Some functions limit where they can be used.
// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
//
// Elements of 'avail' parameter must be in lower case to check function names in case-insensitive.
//
// If this method is not called before checks, ExprSemanticsChecker considers no special function is
// allowed by default. Allowed functions can be obtained from actionlint.SpecialFunctionNames global
// constant.
//
// Available function names for workflow keys can be obtained from actionlint.ContextAvailability.
func (sema *ExprSemanticsChecker) SetSpecialFunctionAvailability(avail []string) {
	sema.availableSpecialFuncs = avail
}

func (sema *ExprSemanticsChecker) checkSpecialFunctionAvailability(n *FuncCallNode) {
	f := strings.ToLower(n.Callee)

	allowed, ok := SpecialFunctionNames[f]
	if !ok {
		return // This function is not special
	}

	for _, sp := range sema.availableSpecialFuncs {
		if sp == f {
			return
		}
	}

	sema.errorf(
		n,
		"calling function %q is not allowed here. %q is only available in %s. see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details",
		n.Callee,
		n.Callee,
		quotes(allowed),
	)
}

func (sema *ExprSemanticsChecker) visitUntrustedCheckerOnEnterNode(n ExprNode) {
	if sema.untrusted != nil {
		sema.untrusted.OnVisitNodeEnter(n)
	}
}

func (sema *ExprSemanticsChecker) visitUntrustedCheckerOnLeaveNode(n ExprNode) {
	if sema.untrusted != nil {
		sema.untrusted.OnVisitNodeLeave(n)
	}
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

	sema.checkAvailableContext(n)
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
		if ty.Mapped != nil {
			if v, ok := n.Receiver.(*VariableNode); ok && v.Name == "vars" {
				sema.checkConfigVariables(n)
			}
			return ty.Mapped
		}
		if ty.IsStrict() {
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
			} else if et.Mapped != nil {
				elem = et.Mapped
			} else if et.IsStrict() {
				sema.errorf(n, "property %q is not defined in object type %s as element of filtered array", n.Property, et.String())
			}
			return &ArrayType{elem, true}
		default:
			sema.errorf(
				n,
				"property filtered by %q at object filtering must be type of object but got %q",
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

func (sema *ExprSemanticsChecker) checkConfigVariables(n *ObjectDerefNode) {
	// https://docs.github.com/en/actions/learn-github-actions/variables#naming-conventions-for-configuration-variables
	if strings.HasPrefix(n.Property, "github_") {
		sema.errorf(
			n,
			"configuration variable name %q must not start with the GITHUB_ prefix (case insensitive). note: see the convention at https://docs.github.com/en/actions/learn-github-actions/variables#naming-conventions-for-configuration-variables",
			n.Property,
		)
		return
	}
	for _, r := range n.Property {
		// Note: `n.Property` was already converted to lower case by parser
		// Note: First character cannot be number, but it was already checked by parser
		if '0' <= r && r <= '9' || 'a' <= r && r <= 'z' || r == '_' {
			continue
		}
		sema.errorf(
			n,
			"configuration variable name %q can only contain alphabets, decimal numbers, and '_'. note: see the convention at https://docs.github.com/en/actions/learn-github-actions/variables#naming-conventions-for-configuration-variables",
			n.Property,
		)
		return
	}

	if sema.configVars == nil {
		return
	}
	if len(sema.configVars) == 0 {
		sema.errorf(
			n,
			"no configuration variable is allowed since the variables list is empty in actionlint.yaml. you may forget adding the variable %q to the list",
			n.Property,
		)
		return
	}

	for _, v := range sema.configVars {
		if strings.EqualFold(v, n.Property) {
			return
		}
	}

	sema.errorf(
		n,
		"undefined configuration variable %q. defined configuration variables in actionlint.yaml are %s",
		n.Property,
		sortedQuotes(sema.configVars),
	)
}

func (sema *ExprSemanticsChecker) checkArrayDeref(n *ArrayDerefNode) ExprType {
	switch ty := sema.check(n.Receiver).(type) {
	case AnyType:
		return &ArrayType{AnyType{}, true}
	case *ArrayType:
		ty.Deref = true
		return ty
	case *ObjectType:
		// Object filtering is available for objects, not only arrays (#66)

		if ty.Mapped != nil {
			// For map object or loose object at receiver of .*
			switch mty := ty.Mapped.(type) {
			case AnyType:
				return &ArrayType{AnyType{}, true}
			case *ObjectType:
				return &ArrayType{mty, true}
			default:
				sema.errorf(n, "elements of object at receiver of object filtering `.*` must be type of object but got %q. the type of receiver was %q", mty.String(), ty.String())
				return AnyType{}
			}
		}

		// For strict object at receiver of .*
		found := false
		for _, t := range ty.Props {
			if _, ok := t.(*ObjectType); ok {
				found = true
				break
			}
		}
		if !found {
			sema.errorf(n, "object type %q cannot be filtered by object filtering `.*` since it has no object element", ty.String())
			return AnyType{}
		}

		return &ArrayType{AnyType{}, true}
	default:
		sema.errorf(n, "receiver of object filtering `.*` must be type of array or object but got %q", ty.String())
		return AnyType{}
	}
}

func (sema *ExprSemanticsChecker) checkIndexAccess(n *IndexAccessNode) ExprType {
	// Note: Index must be visited before Index to make UntrustedInputChecker work correctly even if
	// the expression has some nest like foo[aaa.bbb].bar. Nest happens in top-down order and
	// properties/indices access check is done in bottom-up order. So, as far as we visit nested
	// index nodes before visiting operand, the index is recursively checked first.
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
				if ty.Mapped != nil {
					return ty.Mapped
				}
				if ty.IsStrict() {
					sema.errorf(n, "property %q is not defined in object type %s", lit.Value, ty.String())
				}
			}
			if ty.Mapped != nil {
				return ty.Mapped
			}
			return AnyType{} // Fallback
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
			"number of arguments is wrong. function %q takes %s%d parameters but %d arguments are given",
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

func (sema *ExprSemanticsChecker) checkBuiltinFuncCall(n *FuncCallNode, sig *FuncSignature) ExprType {
	sema.checkSpecialFunctionAvailability(n)

	// Special checks for specific built-in functions
	switch strings.ToLower(n.Callee) {
	case "format":
		lit, ok := n.Args[0].(*StringNode)
		if !ok {
			return sig.Ret
		}
		l := len(n.Args) - 1 // -1 means removing first format string argument

		holders := parseFormatFuncSpecifiers(lit.Value, l)

		for i := 0; i < l; i++ {
			if _, ok := holders[i]; !ok {
				sema.errorf(n, "format string %q does not contain placeholder {%d}. remove argument which is unused in the format string", lit.Value, i)
				continue
			}
			delete(holders, i) // forget it to check unused placeholders
		}

		for i := range holders {
			sema.errorf(n, "format string %q contains placeholder {%d} but only %d arguments are given to format", lit.Value, i, l)
		}
	case "fromjson":
		lit, ok := n.Args[0].(*StringNode)
		if !ok {
			return sig.Ret
		}
		var v any
		err := json.Unmarshal([]byte(lit.Value), &v)
		if err == nil {
			return typeOfJSONValue(v)
		}
		if s, ok := err.(*json.SyntaxError); ok {
			sema.errorf(lit, "broken JSON string is passed to fromJSON() at offset %d: %s", s.Offset, s)
		}
	}

	return sig.Ret
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
			return sema.checkBuiltinFuncCall(n, sig)
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

func validateCompareOpOperands(op CompareOpNodeKind, l, r ExprType) bool {
	// Comparison behavior: https://docs.github.com/en/actions/learn-github-actions/expressions#operators
	switch op {
	case CompareOpNodeKindEq, CompareOpNodeKindNotEq:
		switch l := l.(type) {
		case AnyType, NullType:
			return true
		case NumberType, BoolType, StringType:
			switch r.(type) {
			case *ObjectType, *ArrayType:
				// These are coerced to NaN hence the comparison result is always false
				return false
			default:
				return true
			}
		case *ObjectType:
			switch r.(type) {
			case *ObjectType, NullType, AnyType:
				return true
			default:
				return false
			}
		case *ArrayType:
			switch r := r.(type) {
			case *ArrayType:
				return validateCompareOpOperands(op, l.Elem, r.Elem)
			case NullType, AnyType:
				return true
			default:
				return false
			}
		default:
			panic("unreachable")
		}
	case CompareOpNodeKindLess, CompareOpNodeKindLessEq, CompareOpNodeKindGreater, CompareOpNodeKindGreaterEq:
		// null, bool, array, and object cannot be compared with these operators
		switch l.(type) {
		case AnyType, NumberType, StringType:
			switch r.(type) {
			case NullType, BoolType, *ObjectType, *ArrayType:
				return false
			default:
				return true
			}
		case NullType, BoolType, *ObjectType, *ArrayType:
			return false
		default:
			panic("unreachable")
		}
	default:
		return true
	}
}

func (sema *ExprSemanticsChecker) checkCompareOp(n *CompareOpNode) ExprType {
	l := sema.check(n.Left)
	r := sema.check(n.Right)

	if !validateCompareOpOperands(n.Kind, l, r) {
		sema.errorf(n, "%q value cannot be compared to %q value with %q operator", l.String(), r.String(), n.Kind.String())
	}

	return BoolType{}
}

// checkWithNarrowing checks type of given expression with type narrowing. Type narrowing narrows
// down the type of the expression by assuming its value. For example, `l && r` is typed as
// `typeof(l) | typeof(r)` usually. However when the expression is assumed to be true, its type can
// be narrowed down to `typeof(r)`.
// This analysis is useful to make type checking more accurate. For example, `some_var && 60 || 20`
// can be typed as `number` instead of `typeof(some_var) | number`. (#384)
func (sema *ExprSemanticsChecker) checkWithNarrowing(n ExprNode, isTruthy bool) ExprType {
	switch n := n.(type) {
	case *LogicalOpNode:
		switch n.Kind {
		case LogicalOpNodeKindAnd:
			// When `l && r` is true, narrow its type to `typeof(r)`
			if isTruthy {
				sema.check(n.Left)
				return sema.check(n.Right)
			}
		case LogicalOpNodeKindOr:
			// When `l || r` is false, narrow its type to `typeof(r)`
			if !isTruthy {
				sema.check(n.Left)
				return sema.check(n.Right)
			}
		}
		return sema.checkLogicalOp(n)
	case *NotOpNode:
		return sema.checkWithNarrowing(n.Operand, !isTruthy)
	default:
		return sema.check(n)
	}
}

func (sema *ExprSemanticsChecker) checkLogicalOp(n *LogicalOpNode) ExprType {
	switch n.Kind {
	case LogicalOpNodeKindAnd:
		// When `l` is false in `l && r`, its type is `typeof(l)`. Otherwise `typeof(r)`.
		// Narrow the type of LHS expression by assuming its value is falsy.
		return sema.checkWithNarrowing(n.Left, false).Merge(sema.check(n.Right))
	case LogicalOpNodeKindOr:
		// When `l` is true in `l || r`, its type is `typeof(l)`. Otherwise `typeof(r).
		// Narrow the type of LHS expression by assuming its value is truthy.
		return sema.checkWithNarrowing(n.Left, true).Merge(sema.check(n.Right))
	default:
		sema.check(n.Left)
		sema.check(n.Right)
		return AnyType{}
	}
}

func (sema *ExprSemanticsChecker) check(expr ExprNode) ExprType {
	sema.visitUntrustedCheckerOnEnterNode(expr)
	defer sema.visitUntrustedCheckerOnLeaveNode(expr) // Call this method in bottom-up order

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

// Check checks semantics of given expression syntax tree. It returns the type of the expression as
// the first return value when the check was successfully done. And it returns all errors found
// while checking the expression as the second return value.
func (sema *ExprSemanticsChecker) Check(expr ExprNode) (ExprType, []*ExprError) {
	sema.errs = []*ExprError{}
	if sema.untrusted != nil {
		sema.untrusted.Init()
	}
	ty := sema.check(expr)
	errs := sema.errs
	if sema.untrusted != nil {
		sema.untrusted.OnVisitEnd()
		errs = append(errs, sema.untrusted.Errs()...)
	}
	return ty, errs
}
