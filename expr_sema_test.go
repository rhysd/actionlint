package actionlint

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/google/go-cmp/cmp"
)

func TestExprSemanticsCheckOK(t *testing.T) {
	testCases := []struct {
		what          string
		input         string
		expected      ExprType
		funcs         map[string][]*FuncSignature
		matrix        *ObjectType
		steps         *ObjectType
		needs         *ObjectType
		inputs        *ObjectType
		secrets       *ObjectType
		jobs          *ObjectType
		availContexts []string
		availSPFuncs  []string
		configVars    []string
	}{
		{
			what:     "null",
			input:    "null",
			expected: NullType{},
		},
		{
			what:     "bool",
			input:    "true",
			expected: BoolType{},
		},
		{
			what:     "integer",
			input:    "42",
			expected: NumberType{},
		},
		{
			what:     "float",
			input:    "-3.14e16",
			expected: NumberType{},
		},
		{
			what:     "string",
			input:    "'this is string'",
			expected: StringType{},
		},
		{
			what:     "variable",
			input:    "github",
			expected: BuiltinGlobalVariableTypes["github"],
		},
		{
			what:     "object property dereference",
			input:    "test().bar.piyo",
			expected: BoolType{},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewObjectType(map[string]ExprType{
							"bar": NewObjectType(map[string]ExprType{
								"piyo": BoolType{},
							}),
						}),
					},
				},
			},
		},
		{
			what:     "object property dereference of global variable",
			input:    "job.container.network",
			expected: StringType{},
		},
		{
			what:     "object property dereference for any type",
			input:    "github.event.labels",
			expected: AnyType{},
		},
		{
			what:     "array element dereference",
			input:    "test().bar.*",
			expected: &ArrayType{BoolType{}, true},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewObjectType(map[string]ExprType{
							"bar": &ArrayType{Elem: BoolType{}},
						}),
					},
				},
			},
		},
		{
			what:     "filter object property by array element dereference",
			input:    "test().foo.*.bar.piyo",
			expected: &ArrayType{StringType{}, true},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewObjectType(map[string]ExprType{
							"foo": &ArrayType{
								Elem: NewObjectType(map[string]ExprType{
									"bar": NewObjectType(map[string]ExprType{
										"piyo": StringType{},
									}),
								}),
							},
						}),
					},
				},
			},
		},
		{
			what:     "filter strict object property by array element dereference",
			input:    "test().foo.*.bar.piyo",
			expected: &ArrayType{StringType{}, true},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewStrictObjectType(map[string]ExprType{
							"foo": &ArrayType{
								Elem: NewStrictObjectType(map[string]ExprType{
									"bar": NewStrictObjectType(map[string]ExprType{
										"piyo": StringType{},
									}),
								}),
							},
						}),
					},
				},
			},
		},
		// TODO: Add strictprops
		{
			what:     "array element dereference with any type",
			input:    "github.event.labels.*.name",
			expected: &ArrayType{AnyType{}, true},
		},
		{
			what:     "nested array element dereference",
			input:    "github.event.issues.*.labels.*.name",
			expected: &ArrayType{AnyType{}, true},
		},
		{
			what:     "array element dereference with object type",
			input:    "github.*.name",
			expected: &ArrayType{AnyType{}, true},
		},
		{
			what:     "function call",
			input:    "contains('hello', 'll')",
			expected: BoolType{},
		},
		{
			what:     "function call overload",
			input:    "contains(github.event.labels, 'foo')",
			expected: BoolType{},
		},
		{
			what:     "function call zero arguments",
			input:    "always()",
			expected: BoolType{},
		},
		{
			what:     "function call variable length parameters",
			input:    "format('hello {0} {1}', 42, true)",
			expected: StringType{},
		},
		{
			what:     "object property index access",
			input:    "test()['bar']['piyo']",
			expected: BoolType{},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewObjectType(map[string]ExprType{
							"bar": NewObjectType(map[string]ExprType{
								"piyo": BoolType{},
							}),
						}),
					},
				},
			},
		},
		{
			what:     "object property index access with any type",
			input:    "github.event['FOOO']",
			expected: AnyType{},
		},
		{
			what:     "array element dereference",
			input:    "test()[0]",
			expected: BoolType{},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret:  &ArrayType{Elem: BoolType{}},
					},
				},
			},
		},
		{
			what:     "array element index access with any type fallback",
			input:    "github.event.labels[0]",
			expected: AnyType{},
		},
		{
			what:     "index access to dereferenced array",
			input:    "test().foo.*.bar[0]",
			expected: StringType{},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewObjectType(map[string]ExprType{
							"foo": &ArrayType{
								Elem: NewObjectType(map[string]ExprType{
									"bar": StringType{},
								}),
							},
						}),
					},
				},
			},
		},
		{
			what:     "coerce array dereference into array at function parameter",
			input:    "contains(test().*.x, 10)",
			expected: BoolType{},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ArrayType{
							Elem: NewObjectType(map[string]ExprType{
								"x": NumberType{},
							}),
						},
					},
				},
				"contains": BuiltinFuncSignatures["contains"],
			},
		},
		{
			what:     "index access to dereferenced array with any type fallback",
			input:    "github.event.labels.*.name[0]",
			expected: AnyType{},
		},
		{
			what:     "! operator",
			input:    "!true",
			expected: BoolType{},
		},
		{
			what:     "! operator with non-bool operand",
			input:    "!null",
			expected: BoolType{},
		},
		{
			what:     "< operator",
			input:    "0 < 1",
			expected: BoolType{},
		},
		{
			what:     "<= operator",
			input:    "0 <= 1",
			expected: BoolType{},
		},
		{
			what:     "> operator",
			input:    "0 > 1",
			expected: BoolType{},
		},
		{
			what:     ">= operator",
			input:    "0 >= 1",
			expected: BoolType{},
		},
		{
			what:     "== operator",
			input:    "0 == 1",
			expected: BoolType{},
		},
		{
			what:     "!= operator",
			input:    "0 != 1",
			expected: BoolType{},
		},
		{
			what:     "&& operator",
			input:    "true && false",
			expected: BoolType{},
		},
		{
			what:     "|| operator",
			input:    "true || false",
			expected: BoolType{},
		},
		{
			what:     "&& operator with non-bool operands",
			input:    "10 && 'foo'",
			expected: StringType{},
		},
		{
			what:     "|| operator with non-bool operands",
			input:    "'foo' || 42",
			expected: StringType{},
		},
		{
			what:  "coercing two objects on && operator",
			input: "foo() && bar()",
			expected: NewStrictObjectType(map[string]ExprType{
				"foo": NumberType{},
				"bar": BoolType{},
			}),
			funcs: map[string][]*FuncSignature{
				"foo": {
					{
						Name: "foo",
						Ret: NewStrictObjectType(map[string]ExprType{
							"foo": NumberType{},
						}),
					},
				},
				"bar": {
					{
						Name: "bar",
						Ret: NewStrictObjectType(map[string]ExprType{
							"bar": BoolType{},
						}),
					},
				},
			},
		},
		{
			what:  "coercing two objects on || operator",
			input: "foo() || bar()",
			expected: NewStrictObjectType(map[string]ExprType{
				"foo": NumberType{},
				"bar": BoolType{},
			}),
			funcs: map[string][]*FuncSignature{
				"foo": {
					{
						Name: "foo",
						Ret: NewStrictObjectType(map[string]ExprType{
							"foo": NumberType{},
						}),
					},
				},
				"bar": {
					{
						Name: "bar",
						Ret: NewStrictObjectType(map[string]ExprType{
							"bar": BoolType{},
						}),
					},
				},
			},
		},
		{
			what:     "== operator with loose equality check",
			input:    "true == 1.1",
			expected: BoolType{},
		},
		{
			what:     "arguments of format() is not checked when first argument is not a literal",
			input:    "format(github.action, 1, 2, 3)",
			expected: StringType{},
		},
		{
			what:     "matrix value with typed matrix values",
			input:    "matrix.foooo",
			expected: StringType{},
			matrix: NewStrictObjectType(map[string]ExprType{
				"foooo": StringType{},
			}),
		},
		{
			what:     "step output value with typed steps outputs",
			input:    "steps.foo.outputs",
			expected: NewEmptyObjectType(),
			steps: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"outputs":    NewEmptyObjectType(),
					"conclusion": StringType{},
					"outcome":    StringType{},
				}),
			}),
		},
		{
			what:     "step conclusion with typed steps outputs",
			input:    "steps.foo.conclusion",
			expected: StringType{},
			steps: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"outputs":    NewEmptyObjectType(),
					"conclusion": StringType{},
					"outcome":    StringType{},
				}),
			}),
		},
		{
			what:     "needs context object",
			input:    "needs",
			expected: NewEmptyStrictObjectType(),
		},
		{
			what:     "output string in needs context object",
			input:    "needs.foo.outputs.out1",
			expected: StringType{},
			needs: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"outputs": NewStrictObjectType(map[string]ExprType{
						"out1": StringType{},
						"out2": StringType{},
					}),
					"result": StringType{},
				}),
			}),
		},
		{
			what:     "result in needs context object",
			input:    "needs.foo.result",
			expected: StringType{},
			needs: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"outputs": NewStrictObjectType(map[string]ExprType{
						"out1": StringType{},
						"out2": StringType{},
					}),
					"result": StringType{},
				}),
			}),
		},
		{
			what:     "number is coerced into string",
			input:    "startsWith('42foo', 42)",
			expected: BoolType{},
		},
		{
			what:     "string is coerced into bool",
			input:    "!'hello'",
			expected: BoolType{},
		},
		{
			what:     "coerce number into bool",
			input:    "!42",
			expected: BoolType{},
		},
		{
			what:     "coerce null into bool",
			input:    "!null",
			expected: BoolType{},
		},
		{
			what:     "coerce string into bool",
			input:    "!'hello'",
			expected: BoolType{},
		},
		{
			what:     "case insensitive comparison for object property",
			input:    "test().foo-Bar_PIYO",
			expected: NullType{},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewObjectType(map[string]ExprType{
							"foo-bar_piyo": NullType{},
						}),
					},
				},
			},
		},
		{
			what:     "case insensitive comparison for function name",
			input:    "toJSON(fromjson(toJson(github)))",
			expected: StringType{},
		},
		{
			what:     "case insensitive comparison for context name",
			input:    "JOB.CONTAINER.NETWORK",
			expected: StringType{},
		},
		{
			what:     "case insensitive comparison for context name in index",
			input:    "JOB['CONTAINER']['NETWORK']",
			expected: StringType{},
		},
		{
			what:     "format() function arguments validation",
			input:    "format('{0}{0}{0} {1}{2}{1} {1}{2}{1}{2} {0} {1}{1}{1} {2}{2}{2} {0}{0}{0}{0} {0}', 1, 'foo', true)",
			expected: StringType{},
		},
		{
			what:     "braces not for placeholders in format string of format() call",
			input:    "format('{0} {} {x} {', 1)",
			expected: StringType{},
		},
		{
			what:     "map object dereference",
			input:    "env.FOO",
			expected: StringType{},
		},
		{
			what:     "map object dreference on array filter",
			input:    "test().*.foo",
			expected: &ArrayType{NumberType{}, true},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ArrayType{
							Elem: NewMapObjectType(NumberType{}),
						},
					},
				},
			},
		},
		{
			what:     "map object dreference on object filter",
			input:    "test().*.foo",
			expected: &ArrayType{NumberType{}, true},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: NewMapObjectType(
							NewMapObjectType(NumberType{}),
						),
					},
				},
			},
		},
		{
			what:     "loose object at object filter",
			input:    "github.event.*.foo",
			expected: &ArrayType{AnyType{}, true},
		},
		{
			what:     "strict object which has at least one object element at object filter",
			input:    "github.*.foo",
			expected: &ArrayType{AnyType{}, true},
		},
		{
			what:     "map object index access with string literal",
			input:    "env['FOO']",
			expected: StringType{},
		},
		{
			what:     "map object index access with dynamic value",
			input:    "env[github.action]",
			expected: StringType{},
		},
		{
			what:     "nested object in map object",
			input:    "job.services.my_service.network",
			expected: StringType{},
		},
		{
			what:     "narrowed inputs object",
			input:    "inputs.hello",
			expected: NumberType{},
			inputs: NewStrictObjectType(map[string]ExprType{
				"hello": NumberType{},
			}),
		},
		{
			what:     "narrowed secrets object",
			input:    "secrets.token",
			expected: StringType{},
			secrets: NewStrictObjectType(map[string]ExprType{
				"token": StringType{},
			}),
		},
		{
			what:     "default secrets object",
			input:    "secrets.any_value",
			expected: StringType{},
		},
		{
			what:     "automatically supplied secret",
			input:    "secrets.github_token",
			expected: StringType{},
			secrets: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
		},
		{
			what:     "automatically supplied secret",
			input:    "secrets.ACTIONS_STEP_DEBUG",
			expected: StringType{},
			secrets: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
		},
		{
			what:     "automatically supplied secret",
			input:    "secrets.ACTIONS_RUNNER_DEBUG",
			expected: StringType{},
			secrets: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
		},
		{
			what:     "jobs object",
			input:    "jobs.some_job",
			expected: NewEmptyObjectType(),
			jobs: NewStrictObjectType(map[string]ExprType{
				"some_job": NewEmptyObjectType(),
			}),
		},
		{
			what:          "available context",
			input:         "strategy.fail-fast",
			expected:      BoolType{},
			availContexts: []string{"strategy"},
		},
		{
			what:          "available contexts",
			input:         "strategy.fail-fast",
			expected:      BoolType{},
			availContexts: []string{"matrix", "strategy"},
		},
		{
			what:         "special function",
			input:        "success()",
			expected:     BoolType{},
			availSPFuncs: []string{"success"},
		},
		{
			what:         "non-special function",
			input:        "contains('hello, world', 'o, w')",
			expected:     BoolType{},
			availSPFuncs: []string{"always"},
		},
		{
			what:         "case-insensitive special function name",
			input:        "hashFiles('aaa.txt')",
			expected:     StringType{},
			availSPFuncs: []string{"hashfiles"},
		},
		{
			what:     "configuration variable",
			input:    "vars.SOME_VARIABLE",
			expected: StringType{},
		},
		{
			what:       "known configuration variable",
			input:      "vars.SOME_VARIABLE",
			expected:   StringType{},
			configVars: []string{"SOME_VARIABLE"},
		},
		{
			what:       "configuration variable name is case insensitive",
			input:      "vars.SOME_VARIABLE",
			expected:   StringType{},
			configVars: []string{"some_variable"},
		},
		{
			what:     "narrow type of && operator by assumed value (#384)",
			input:    "('foo' && 10) || 20",
			expected: NumberType{},
		},
		{
			what:     "narrow type of || operator by assumed value (#384)",
			input:    "('foo' || 10) && 20",
			expected: NumberType{},
		},
		{
			what:     "narrow type of nested && operator",
			input:    "((('foo' && true) || false) && 10) || 20",
			expected: NumberType{},
		},
		{
			what:     "narrow type of nested || operator",
			input:    "((('foo' || true) && false) || 10) && 20",
			expected: NumberType{},
		},
		{
			what:     "don't narrow type on nested && operator",
			input:    "('foo' && 10) && 20",
			expected: StringType{},
		},
		{
			what:     "don't narrow type on nested || operator",
			input:    "('foo' || 10) || 20",
			expected: StringType{},
		},
		{
			what:     "narrowed || operator at LHS and && operator at RHS",
			input:    "('foo' || 10) && (('foo' && 10) || 20)",
			expected: NumberType{},
		},
		{
			what:     "narrowed && operator at LHS and || operator at RHS",
			input:    "('foo' && 10) || (('foo' || 10) && 20)",
			expected: NumberType{},
		},
		{
			what:     "not operator negates && operator type narrowing",
			input:    "!('foo' && 10) && 20",
			expected: NumberType{},
		},
		{
			what:     "not operator negates || operator type narrowing",
			input:    "!('foo' || 10) || 20",
			expected: NumberType{},
		},
		{
			what:     "double not operators does nothing on type narrowing",
			input:    "!!('foo' || 10) && 20",
			expected: NumberType{},
		},
		{
			what:     "escaped braces in format string",
			input:    "format('hello {{1}} {0}', 42)",
			expected: StringType{},
		},
		{
			what:     "format specifier is escaped",
			input:    "format('hello {{{0}', 'world')", // First {{ is escaped. {0} is not escaped
			expected: StringType{},
		},
		{
			what:  "fromJSON with JSON constant value",
			input: `fromJSON('{"foo":true,"bar":["foo", 12.3],"piyo":null}')`,
			expected: NewStrictObjectType(map[string]ExprType{
				"foo":  BoolType{},
				"bar":  &ArrayType{Elem: StringType{}}, // Element type was merged
				"piyo": NullType{},
			}),
		},
	}

	allSPFuncs := []string{}
	for f := range SpecialFunctionNames {
		allSPFuncs = append(allSPFuncs, f)
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			p := NewExprParser()
			e, err := p.Parse(NewExprLexer(tc.input + "}}"))
			if err != nil {
				t.Fatal("Parse error:", tc.input)
			}

			c := NewExprSemanticsChecker(false, nil)
			c.SetContextAvailability([]string{"github", "job", "jobs", "matrix", "steps", "needs", "env", "inputs", "secrets", "vars", "runner"})
			if tc.funcs != nil {
				c.funcs = tc.funcs
			}
			if tc.matrix != nil {
				c.UpdateMatrix(tc.matrix)
			}
			if tc.steps != nil {
				c.UpdateSteps(tc.steps)
			}
			if tc.needs != nil {
				c.UpdateNeeds(tc.needs)
			}
			if tc.inputs != nil {
				c.UpdateInputs(tc.inputs)
			}
			if tc.secrets != nil {
				c.UpdateSecrets(tc.secrets)
			}
			if tc.jobs != nil {
				c.UpdateJobs(tc.jobs)
			}
			if len(tc.availContexts) > 0 {
				c.SetContextAvailability(tc.availContexts)
			}
			if len(tc.availSPFuncs) > 0 {
				c.SetSpecialFunctionAvailability(tc.availSPFuncs)
			} else {
				c.SetSpecialFunctionAvailability(allSPFuncs)
			}
			ty, errs := c.Check(e)
			if len(errs) > 0 {
				t.Fatal("semantics check failed:", errs)
			}

			if diff := cmp.Diff(tc.expected, ty); diff != "" {
				t.Fatalf("wanted: %s\nbut got:%s\ndiff:\n%s", tc.expected.String(), ty.String(), diff)
			}
		})
	}
}

func TestExprSemanticsCheckError(t *testing.T) {
	testCases := []struct {
		what       string
		input      string
		expected   []string
		funcs      map[string][]*FuncSignature
		matrix     *ObjectType
		steps      *ObjectType
		needs      *ObjectType
		availCtx   []string
		availSP    []string
		configVars []string
	}{
		{
			what:  "undefined variable",
			input: "fooooo",
			expected: []string{
				"undefined variable \"fooooo\"",
			},
		},
		{
			what:  "receiver of object dereference is not an object",
			input: "true.foo",
			expected: []string{
				"receiver of object dereference \"foo\" must be type of object but got \"bool\"",
			},
		},
		{
			what:  "strict prop check",
			input: "github.foo",
			expected: []string{
				"property \"foo\" is not defined in object type",
			},
		},
		{
			what:  "strict prop check at object property filter for array dereference",
			input: "test().*.bar",
			expected: []string{
				"property \"bar\" is not defined in object type",
			},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ArrayType{
							Elem: NewStrictObjectType(map[string]ExprType{
								"foo": BoolType{},
							}),
						},
					},
				},
			},
		},
		{
			what:  "array element is not object for filtering array dereference",
			input: "test().*.bar",
			expected: []string{
				"property filtered by \"bar\" at object filtering must be type of object but got \"string\"",
			},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ArrayType{
							Elem: StringType{},
						},
					},
				},
			},
		},
		{
			what:  "receiver of array dereference is not an array",
			input: "true.*",
			expected: []string{
				"receiver of object filtering `.*` must be type of array or object but got \"bool\"",
			},
		},
		{
			what:  "receiver of object filter is not an object which has object element",
			input: "env.*",
			expected: []string{
				"elements of object at receiver of object filtering `.*` must be type of object but got \"string\". the type of receiver was \"{string => string}\"",
			},
		},
		{
			what:  "receiver of object filter is an object which has no object element",
			input: "runner.*",
			expected: []string{
				"cannot be filtered by object filtering `.*` since it has no object element",
			},
		},
		{
			what:  "index access to invalid value (number)",
			input: "true[0]",
			expected: []string{
				"index access operand must be type of object or array but got \"bool\"",
			},
		},
		{
			what:  "index access to invalid value (string)",
			input: "true['hello']",
			expected: []string{
				"index access operand must be type of object or array but got \"bool\"",
			},
		},
		{
			what:  "index access to array with not a number",
			input: "test()['hi']",
			expected: []string{
				"index access of array must be type of number but got \"string\"",
			},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ArrayType{
							Elem: StringType{},
						},
					},
				},
			},
		},
		{
			what:  "index access to array dereference with not a number",
			input: "test().*['hi']",
			expected: []string{
				"index access of array must be type of number but got \"string\"",
			},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ArrayType{
							Elem: StringType{},
						},
					},
				},
			},
		},
		{
			what:  "index access to object with not a string",
			input: "env[0]",
			expected: []string{
				"property access of object must be type of string but got \"number\"",
			},
		},
		{
			what:  "strict prop check with string literal index access to object",
			input: "github['fooooo']",
			expected: []string{
				"property \"fooooo\" is not defined in object type",
			},
		},
		{
			what:  "undefined function",
			input: "foooo()",
			expected: []string{
				"undefined function \"foooo\"",
			},
		},
		{
			what:  "wrong number of arguments at function call",
			input: "contains('foo')",
			expected: []string{
				"number of arguments is wrong. function \"contains(string, string) -> bool\" takes 2 parameters but 1 arguments are given",
				"number of arguments is wrong. function \"contains(array<any>, any) -> bool\" takes 2 parameters but 1 arguments are given",
			},
		},
		{
			what:  "wrong number of arguments at function call for variable length parameters",
			input: "hashFiles()",
			expected: []string{
				"number of arguments is wrong. function \"hashFiles(string...) -> string\" takes at least 1 parameters but 0 arguments are given",
			},
		},
		{
			what:  "wrong type at parameter",
			input: "startsWith('foo', null)",
			expected: []string{
				"2nd argument of function call is not assignable. \"null\" cannot be assigned to \"string\"",
			},
		},
		{
			what:  "wrong type at parameter of overloaded function",
			input: "contains('foo', null)",
			expected: []string{
				"2nd argument of function call is not assignable. \"null\" cannot be assigned to \"string\"",
				"1st argument of function call is not assignable. \"string\" cannot be assigned to \"array<any>\"",
			},
		},
		{
			what:  "wrong type at rest parameter",
			input: "hashFiles(null)",
			expected: []string{
				"1st argument of function call is not assignable. \"null\" cannot be assigned to \"string\"",
			},
		},
		{
			what:  "wrong type at rest parameter part2",
			input: "hashFiles('foo', null)",
			expected: []string{
				"2nd argument of function call is not assignable. \"null\" cannot be assigned to \"string\"",
			},
		},
		{
			what:  "less arguments for format() builtin function call",
			input: "format('format {0} {1}', 'foo')",
			expected: []string{
				"format string \"format {0} {1}\" contains placeholder {1} but only 1 arguments are given to format",
			},
		},
		{
			what:  "more arguments for format() builtin function call",
			input: "format('format {0} {1} {2}', 'foo', 1, true, null)",
			expected: []string{
				"format string \"format {0} {1} {2}\" does not contain placeholder {3}. remove argument which is unused in the format string",
			},
		},
		{
			what:  "unused placeholder in format string of format()",
			input: "format('format {0} {2}', 1, 2, 3)",
			expected: []string{
				"format string \"format {0} {2}\" does not contain placeholder {1}. remove argument which is unused in the format string",
			},
		},
		{
			what:  "missing placeholder and less argument at the same time in format string of format()",
			input: "format('format {0} {2}', 1, 2)",
			expected: []string{
				"format string \"format {0} {2}\" does not contain placeholder {1}. remove argument which is unused in the format string",
				"format string \"format {0} {2}\" contains placeholder {2} but only 2 arguments are given to format",
			},
		},
		{
			what:  "zero format arguments for format() call",
			input: "format('hi')",
			expected: []string{
				"takes at least 2 parameters but 1 arguments are given",
			},
		},
		{
			what:  "function name of format() call check is case insensitive",
			input: "Format('{0}', 1, 2)",
			expected: []string{
				`format string "{0}" does not contain placeholder {1}`,
			},
		},
		{
			what:  "format specifier is escaped",
			input: "format('hello {{0}}', 'world')",
			expected: []string{
				"does not contain placeholder {0}",
			},
		},
		{
			what:  "format specifier is still escaped",
			input: "format('hello {{{{0}}', 'world')", // First {{ is escaped. {{0}} is still escaped
			expected: []string{
				"does not contain placeholder {0}",
			},
		},
		{
			what:  "undefined matrix value",
			input: "matrix.bar",
			expected: []string{
				"property \"bar\" is not defined in object type {foo: any}",
			},
			matrix: NewStrictObjectType(map[string]ExprType{
				"foo": AnyType{},
			}),
		},
		{
			what:  "type mismatch in matrix value",
			input: "startsWith('hello', matrix.foo)",
			expected: []string{
				"2nd argument of function call is not assignable. \"null\" cannot be assigned to \"string\"",
			},
			matrix: NewStrictObjectType(map[string]ExprType{
				"foo": NullType{},
			}),
		},
		{
			what:  "matrix value with untyped matrix values",
			input: "matrix.foooo",
			expected: []string{
				"property \"foooo\" is not defined in object type {}",
			},
		},
		{
			what:  "undefined step id",
			input: "steps.foo",
			expected: []string{
				"property \"foo\" is not defined in object type {bar: ", // order of prop types in object type changes randomly so we cannot check it easily
			},
			steps: NewStrictObjectType(map[string]ExprType{
				"bar": NewStrictObjectType(map[string]ExprType{
					"outputs":    NewEmptyObjectType(),
					"conclusion": StringType{},
					"outcome":    StringType{},
				}),
			}),
		},
		{
			what:  "invalid property in step object",
			input: "steps.bar.foo",
			expected: []string{
				"property \"foo\" is not defined in object type {", // order of prop types in object type changes randomly so we cannot check it easily
			},
			steps: NewStrictObjectType(map[string]ExprType{
				"bar": NewStrictObjectType(map[string]ExprType{
					"outputs":    NewEmptyObjectType(),
					"conclusion": StringType{},
					"outcome":    StringType{},
				}),
			}),
		},
		{
			what:  "step output value without typed steps outputs",
			input: "steps.foo.outputs",
			expected: []string{
				"property \"foo\" is not defined in object type {}",
			},
		},
		{
			what:  "undefined job id in needs context",
			input: "needs.bar",
			expected: []string{
				"property \"bar\" is not defined in object type ",
			},
			needs: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"outputs": NewStrictObjectType(map[string]ExprType{
						"out1": StringType{},
						"out2": StringType{},
					}),
					"result": StringType{},
				}),
			}),
		},
		{
			what:  "undefined output in needs context",
			input: "needs.foo.outputs.out3",
			expected: []string{
				"property \"out3\" is not defined in object type ",
			},
			needs: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"outputs": NewStrictObjectType(map[string]ExprType{
						"out1": StringType{},
						"out2": StringType{},
					}),
					"result": StringType{},
				}),
			}),
		},
		{
			what:  "undefined prop in needs context",
			input: "needs.foo.bar",
			expected: []string{
				"property \"bar\" is not defined in object type ",
			},
			needs: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"outputs": NewStrictObjectType(map[string]ExprType{
						"out1": StringType{},
						"out2": StringType{},
					}),
					"result": StringType{},
				}),
			}),
		},
		{
			what:  "undefined prop in untyped needs context",
			input: "needs.foo",
			expected: []string{
				"property \"foo\" is not defined in object type ",
			},
		},
		{
			what:  "bool literal in upper case",
			input: "TRUE",
			expected: []string{
				"undefined variable \"TRUE\"",
			},
		},
		{
			what:  "null literal in upper case",
			input: "NULL",
			expected: []string{
				"undefined variable \"NULL\"",
			},
		},
		{
			what:  "inputs context accepts nothing by default",
			input: "inputs.hello",
			expected: []string{
				"property \"hello\" is not defined in object type {}",
			},
		},
		{
			what:  "available context",
			input: "env.FOO",
			expected: []string{
				"context \"env\" is not allowed here. available context is \"github\"",
			},
			availCtx: []string{"github"},
		},
		{
			what:  "available contexts",
			input: "github.event.labels",
			expected: []string{
				"context \"github\" is not allowed here. available contexts are \"env\", \"matrix\"",
			},
			availCtx: []string{"env", "matrix"},
		},
		{
			what:  "no available context",
			input: "github",
			expected: []string{
				"context \"github\" is not allowed here. no context is available",
			},
			availCtx: []string{},
		},
		{
			what:  "no special function allowed",
			input: "success()",
			expected: []string{
				"calling function \"success\" is not allowed here. \"success\" is only available in ",
			},
			availSP: []string{},
		},
		{
			what:  "special function",
			input: "always()",
			expected: []string{
				"calling function \"always\" is not allowed here. \"always\" is only available in ",
			},
			availSP: []string{"fail", "success"},
		},
		{
			what:  "case-insensitive special function name",
			input: "hashFiles('aaa.txt')",
			expected: []string{
				"calling function \"hashFiles\" is not allowed here. \"hashFiles\" is only available in ",
			},
			availSP: []string{"fail"},
		},
		{
			what:  "no configuration variable is allowed",
			input: "vars.UNKNOWN_VARIABLE",
			expected: []string{
				"no configuration variable is allowed since the variables list is empty",
			},
			configVars: []string{},
		},
		{
			what:  "unknown configuration variable",
			input: "vars.UNKNOWN_VARIABLE",
			expected: []string{
				"undefined configuration variable \"UNKNOWN_VARIABLE\".",
			},
			configVars: []string{"FOO_BAR"},
		},
		{
			what:  "config variable naming convention",
			input: "vars.FOO-BAR",
			expected: []string{
				"configuration variable name \"FOO-BAR\" can only contain alphabets, decimal numbers, and '_'.",
			},
		},
		{
			what:  "config variable name cannot start with GITHUB_",
			input: "vars.GITHUB_FOOOOOOOO",
			expected: []string{
				"must not start with the GITHUB_ prefix",
			},
		},
		{
			what:  "broken JSON value at fromJSON argument",
			input: `fromJSON('{"foo": true')`,
			expected: []string{
				"broken JSON string is passed to fromJSON() at offset 12",
			},
		},
	}

	allSP := []string{}
	for f := range SpecialFunctionNames {
		allSP = append(allSP, f)
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			p := NewExprParser()
			e, err := p.Parse(NewExprLexer(tc.input + "}}"))
			if err != nil {
				t.Fatal("Parse error:", tc.input)
			}

			c := NewExprSemanticsChecker(false, tc.configVars)
			if tc.funcs != nil {
				c.funcs = tc.funcs // Set functions for testing
			}
			if tc.matrix != nil {
				c.UpdateMatrix(tc.matrix)
			}
			if tc.steps != nil {
				c.UpdateSteps(tc.steps)
			}
			if tc.needs != nil {
				c.UpdateNeeds(tc.needs)
			}
			if tc.availCtx != nil {
				c.SetContextAvailability(tc.availCtx)
			} else {
				c.SetContextAvailability([]string{"github", "job", "jobs", "matrix", "steps", "needs", "env", "inputs", "secrets", "vars", "runner"})
			}
			if tc.availSP != nil {
				c.SetSpecialFunctionAvailability(tc.availSP)
			} else {
				c.SetSpecialFunctionAvailability(allSP)
			}

			_, errs := c.Check(e)
			if len(errs) != len(tc.expected) {
				t.Fatalf("semantics check should report %d errors but got %d errors: %v", len(tc.expected), len(errs), errs)
			}
		LoopErrs:
			for _, err := range errs {
				for _, want := range tc.expected {
					if strings.Contains(err.Error(), want) {
						continue LoopErrs
					}
				}
				t.Fatalf("error %q did not match any expected error messages %#v", err.Error(), tc.expected)
			}
		})
	}
}

func TestExprCompareOperandsCheck(t *testing.T) {
	// Matrix of operator -> lhs type -> rhs type -> result
	// Result `true` means the comparison is allowed. `false` means the comparison causes an error.
	matrix := map[string]map[string]map[string]bool{
		"==": {
			"number": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   true,
				"null":   true,
				"object": false,
				"array":  false,
			},
			"string": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   true,
				"null":   true,
				"object": false,
				"array":  false,
			},
			"bool": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   true,
				"null":   true,
				"object": false,
				"array":  false,
			},
			"null": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   true,
				"null":   true,
				"object": true,
				"array":  true,
			},
			"object": {
				"any":    true,
				"number": false,
				"string": false,
				"bool":   false,
				"null":   true,
				"object": true,
				"array":  false,
			},
			"array": {
				"any":      true,
				"number":   false,
				"string":   false,
				"bool":     false,
				"null":     true,
				"object":   false,
				"array":    true,
				"array_2d": false,
			},
			"array_2d": {
				"any":      true,
				"number":   false,
				"string":   false,
				"bool":     false,
				"null":     true,
				"object":   false,
				"array":    false,
				"array_2d": true,
			},
			"any": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   true,
				"null":   true,
				"object": true,
				"array":  true,
			},
		},
		"<": {
			"number": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   false,
				"null":   false,
				"object": false,
				"array":  false,
			},
			"string": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   false,
				"null":   false,
				"object": false,
				"array":  false,
			},
			"bool": {
				"any":    false,
				"number": false,
				"string": false,
				"bool":   false,
				"null":   false,
				"object": false,
				"array":  false,
			},
			"null": {
				"any":    false,
				"number": false,
				"string": false,
				"bool":   false,
				"null":   false,
				"object": false,
				"array":  false,
			},
			"any": {
				"any":    true,
				"number": true,
				"string": true,
				"bool":   false,
				"null":   false,
				"object": false,
				"array":  false,
			},
			"object": {
				"any":    false,
				"number": false,
				"string": false,
				"bool":   false,
				"null":   false,
				"object": false,
				"array":  false,
			},
			"array": {
				"any":      false,
				"number":   false,
				"string":   false,
				"bool":     false,
				"null":     false,
				"object":   false,
				"array":    false,
				"array_2d": false,
			},
		},
	}

	for op, expr := range matrix {
		for lhs, rest := range expr {
			for rhs, ok := range rest {
				input := lhs + " " + op + " " + rhs
				t.Run(input, func(t *testing.T) {
					p := NewExprParser()
					e, err := p.Parse(NewExprLexer(input + "}}"))
					if err != nil {
						t.Fatal("Parse error:", input)
					}

					c := NewExprSemanticsChecker(false, nil)
					c.vars = map[string]ExprType{
						"number": NumberType{},
						"string": StringType{},
						"bool":   BoolType{},
						"object": NewEmptyObjectType(),
						"array":  &ArrayType{Elem: NumberType{}},
						"array_2d": &ArrayType{
							Elem: &ArrayType{Elem: NumberType{}},
						},
						"any": AnyType{},
					}
					c.SetContextAvailability([]string{
						"any",
						"number",
						"string",
						"bool",
						"null",
						"object",
						"array",
						"array_2d",
					})

					ty, errs := c.Check(e)
					if ok {
						if len(errs) > 0 {
							t.Fatal("semantics check failed:", errs)
						}
						if _, ok := ty.(BoolType); !ok {
							t.Fatalf("wanted bool type but have %q", ty)
						}
					} else {
						if len(errs) != 1 {
							t.Fatal("more than one error occurred", errs)
						}
						msg := errs[0].Error()

						toType := func(n string) string {
							switch n {
							case "array":
								return "array<number>"
							case "array_2d":
								return "array<array<number>>"
							default:
								return n
							}
						}
						want := fmt.Sprintf("%q value cannot be compared to %q value with %q operator", toType(lhs), toType(rhs), op)

						if !strings.Contains(msg, want) {
							t.Fatalf("error message %q doesn't contain expected message %q", msg, want)
						}
					}
				})
			}
		}
	}
}

func TestExprBuiltinFunctionSignatures(t *testing.T) {
	for name, sigs := range BuiltinFuncSignatures {
		if len(sigs) == 0 {
			t.Errorf("overload candidates of %q should not be empty", name)
		}
		{
			ok := true
			for _, r := range name {
				if !unicode.IsLower(r) {
					ok = false
					break
				}
			}
			if !ok {
				t.Errorf("name of function must be in lower case to check in case insensitive: %q", name)
			}
		}
		for i, sig := range sigs {
			if name != strings.ToLower(sig.Name) {
				t.Errorf("name of %dth overload is different from its key: name=%q vs key=%q", i+1, sig.Name, name)
			}
			if sig.VariableLengthParams && len(sig.Params) == 0 {
				t.Errorf("number of arguments of %dth overload of %q must not be empty because VariableLengthParams is set to true", i+1, name)
			}
		}
	}
}

func TestExprSemanticsCheckerUpdateMatrix(t *testing.T) {
	c := NewExprSemanticsChecker(false, nil)
	ty := NewEmptyObjectType()
	prev := c.vars["matrix"]
	c.UpdateMatrix(ty)
	if c.vars["matrix"] == prev {
		t.Fatalf("Global variables map was not copied")
	}
	prev = c.vars["matrix"]
	c.UpdateMatrix(ty)
	if c.vars["matrix"] != prev {
		t.Fatalf("Global variables map was copied when calling UpdateMatrix again")
	}
}

func TestExprSemanticsCheckerUpdateSteps(t *testing.T) {
	c := NewExprSemanticsChecker(false, nil)
	ty := NewEmptyObjectType()
	prev := c.vars["steps"]
	c.UpdateSteps(ty)
	if c.vars["steps"] == prev {
		t.Fatalf("Global variables map was not copied")
	}
	prev = c.vars["steps"]
	c.UpdateSteps(ty)
	if c.vars["steps"] != prev {
		t.Fatalf("Global variables map was copied when calling UpdateSteps again")
	}
}

func TestExprSematincsCheckerUpdateDispatchInputsVarType(t *testing.T) {
	ty := NewStrictObjectType(map[string]ExprType{"foo": NullType{}})
	c := NewExprSemanticsChecker(false, nil)
	c.UpdateDispatchInputs(ty)
	o := c.vars["github"].(*ObjectType).Props["event"].(*ObjectType).Props["inputs"].(*ObjectType)
	if _, ok := o.Props["foo"]; !ok {
		t.Error("Local github.event.inputs is not updated", o)
	}
	if !o.IsStrict() {
		t.Error("Local github.event.inputs is not strict")
	}
	for n, ty := range o.Props {
		if _, ok := ty.(StringType); !ok {
			t.Errorf("Type of %q input is not string: %s", n, ty)
		}
	}

	// Check global value is not polluted
	o = BuiltinGlobalVariableTypes["github"].(*ObjectType).Props["event"].(*ObjectType)
	if _, ok := o.Props["inputs"]; ok {
		t.Error("Global github.event.inputs exists", o)
	}
}

func TestExprSemanticsCheckerUpdateInputsMultipleTimes(t *testing.T) {
	tests := []struct {
		first  *ObjectType
		second *ObjectType
		want   *ObjectType
	}{
		{
			first: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			second: NewStrictObjectType(map[string]ExprType{
				"bar": NumberType{},
			}),
			want: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
				"bar": NumberType{},
			}),
		},
		{
			first: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			second: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
		},
		{
			first: NewStrictObjectType(map[string]ExprType{
				"foo": BoolType{},
			}),
			second: NewStrictObjectType(map[string]ExprType{
				"foo": NumberType{},
			}),
			want: NewStrictObjectType(map[string]ExprType{
				"foo": AnyType{},
			}),
		},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("%v then %v", tc.first, tc.second)
		t.Run(name, func(t *testing.T) {
			c := NewExprSemanticsChecker(false, nil)
			c.UpdateInputs(tc.first)
			c.UpdateDispatchInputs(tc.second)
			have := c.vars["inputs"]
			if diff := cmp.Diff(have, tc.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func testObjectPropertiesAreInLowerCase(t *testing.T, ty ExprType) {
	t.Helper()
	switch ty := ty.(type) {
	case *ObjectType:
		for n, ty := range ty.Props {
			for _, r := range n {
				if 'A' <= r && r <= 'Z' {
					t.Errorf("Property of object must not contain uppercase character because comparison is case insensitive but got %q", n)
					break
				}
			}
			testObjectPropertiesAreInLowerCase(t, ty)
		}
	case *ArrayType:
		testObjectPropertiesAreInLowerCase(t, ty.Elem)
	}
}

func TestBuiltinGlobalVariableTypesValidation(t *testing.T) {
	for ctx, ty := range BuiltinGlobalVariableTypes {
		for _, r := range ctx {
			if 'A' <= r && r <= 'Z' {
				t.Errorf("Context name must not contain uppercase character because comparison is case insensitive but got %q", ctx)
				break
			}
		}
		testObjectPropertiesAreInLowerCase(t, ty)
	}
}

func TestParseFormatSpecifiers(t *testing.T) {
	tests := []struct {
		what string
		in   string
		want []int // Specifiers in the `in` string
	}{
		{
			what: "empty input",
			in:   "",
		},
		{
			what: "no specifier",
			in:   "hello, world!",
		},
		{
			what: "single specifier",
			in:   "Hello{0}specifier",
			want: []int{0},
		},
		{
			what: "multiple specifiers",
			in:   "{0} {1}{2}x{3}}{4}!",
			want: []int{0, 1, 2, 3, 4},
		},
		{
			what: "many specifiers",
			in:   "{0}{1}{2}{3}{4}{5}{6}{7}{8}{9}{10}!",
			want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			what: "unordered",
			in:   "{2}foo{4} {0}{3}",
			want: []int{2, 4, 0, 3},
		},
		{
			what: "uncontiguous",
			in:   "{0} {2}foo{5} {1}",
			want: []int{0, 2, 5, 1},
		},
		{
			what: "unclosed",
			in:   "{12foo",
		},
		{
			what: "not digit",
			in:   "{hello}",
		},
		{
			what: "space in digits",
			in:   "{1 2}",
		},
		{
			what: "empty",
			in:   "{}",
		},
		{
			what: "specifier inside specifier",
			in:   "{1{0}2}",
			want: []int{0},
		},
		{
			what: "escaped",
			in:   "{{hello{{0}{{{{1}world}}",
		},
		{
			what: "after escaped",
			in:   "{{{{{0}",
			want: []int{0},
		},
		{
			what: "kuma-",
			in:   "{・{ᴥ}・}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			want := map[int]struct{}{}
			for _, i := range tc.want {
				want[i] = struct{}{}
			}
			have := parseFormatFuncSpecifiers(tc.in, len(tc.want))

			if diff := cmp.Diff(want, have); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

// vim: nofoldenable
