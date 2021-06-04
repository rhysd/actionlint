package actionlint

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestExprSemanticsCheckOK(t *testing.T) {
	testCases := []struct {
		what     string
		input    string
		expected ExprType
		funcs    map[string][]*FuncSignature
		matrix   *ObjectType
		steps    *ObjectType
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
			expected: BuiltinGlobalVariableTypes["github"].Type,
		},
		{
			what:     "object property dereference",
			input:    "test().bar.piyo",
			expected: BoolType{},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ObjectType{
							Props: map[string]ExprType{
								"bar": &ObjectType{
									Props: map[string]ExprType{
										"piyo": BoolType{},
									},
								},
							},
						},
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
			expected: &ArrayDerefType{Elem: BoolType{}},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ObjectType{
							Props: map[string]ExprType{
								"bar": &ArrayType{Elem: BoolType{}},
							},
						},
					},
				},
			},
		},
		{
			what:     "filter object property by array element dereference",
			input:    "test().foo.*.bar.piyo",
			expected: &ArrayDerefType{Elem: StringType{}},
			funcs: map[string][]*FuncSignature{
				"test": {
					{
						Name: "test",
						Ret: &ObjectType{
							Props: map[string]ExprType{
								"foo": &ArrayType{
									Elem: &ObjectType{
										Props: map[string]ExprType{
											"bar": &ObjectType{
												Props: map[string]ExprType{
													"piyo": StringType{},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			what:     "array element dereference with any type",
			input:    "github.event.labels.*.name",
			expected: &ArrayDerefType{Elem: AnyType{}},
		},
		{
			what:     "nested array element dereference",
			input:    "github.event.issues.*.labels.*.name",
			expected: &ArrayDerefType{Elem: AnyType{}},
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
						Ret: &ObjectType{
							Props: map[string]ExprType{
								"bar": &ObjectType{
									Props: map[string]ExprType{
										"piyo": BoolType{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			what:     "object property index access with any type",
			input:    "env['FOOO']",
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
						Ret: &ObjectType{
							Props: map[string]ExprType{
								"foo": &ArrayType{
									Elem: &ObjectType{
										Props: map[string]ExprType{
											"bar": StringType{},
										},
									},
								},
							},
						},
					},
				},
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
			matrix: &ObjectType{
				Props: map[string]ExprType{
					"foooo": StringType{},
				},
				StrictProps: true,
			},
		},
		{
			what:     "step output value with typed steps outputs",
			input:    "steps.foo.outputs",
			expected: NewObjectType(),
			steps: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ObjectType{
						Props: map[string]ExprType{
							"outputs":    NewObjectType(),
							"conclusion": StringType{},
							"outcome":    StringType{},
						},
						StrictProps: true,
					},
				},
				StrictProps: true,
			},
		},
		{
			what:     "step conclusion with typed steps outputs",
			input:    "steps.foo.conclusion",
			expected: StringType{},
			steps: &ObjectType{
				Props: map[string]ExprType{
					"foo": &ObjectType{
						Props: map[string]ExprType{
							"outputs":    NewObjectType(),
							"conclusion": StringType{},
							"outcome":    StringType{},
						},
						StrictProps: true,
					},
				},
				StrictProps: true,
			},
		},
	}

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(AnyType{}),
		cmpopts.IgnoreUnexported(NullType{}),
		cmpopts.IgnoreUnexported(NumberType{}),
		cmpopts.IgnoreUnexported(BoolType{}),
		cmpopts.IgnoreUnexported(StringType{}),
		cmpopts.IgnoreUnexported(ObjectType{}),
		cmpopts.IgnoreUnexported(ArrayType{}),
		cmpopts.IgnoreUnexported(ArrayDerefType{}),
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			l := NewExprLexer()
			tok, _, err := l.Lex(tc.input + "}}")
			if err != nil {
				t.Fatal("Lex error:", err)
			}

			p := NewExprParser()
			e, err := p.Parse(tok)
			if err != nil {
				t.Fatal("Parse error:", tc.input)
			}

			c := NewExprSemanticsChecker()
			if tc.funcs != nil {
				c.funcs = tc.funcs
			}
			if tc.matrix != nil {
				c.UpdateMatrix(tc.matrix)
			}
			if tc.steps != nil {
				c.UpdateSteps(tc.steps)
			}
			ty, errs := c.Check(e)
			if len(errs) > 0 {
				t.Fatal("semantics check failed:", errs)
			}

			if !cmp.Equal(tc.expected, ty, opts...) {
				t.Fatalf("wanted: %s\nbut got:%s\ndiff:\n%s", tc.expected.String(), ty.String(), cmp.Diff(tc.expected, ty, opts...))
			}
		})
	}
}

func TestExprBuiltinFunctionSignatures(t *testing.T) {
	for name, sigs := range BuiltinFuncSignatures {
		if len(sigs) == 0 {
			t.Errorf("overload candidates of %q should not be empty", name)
		}
		for i, sig := range sigs {
			if name != sig.Name {
				t.Errorf("name of %dth overload is different from its key: name=%q vs key=%q", i+1, sig.Name, name)
			}
			if sig.VariableLengthParams && len(sig.Params) == 0 {
				t.Errorf("number of arguments of %dth overload of %q must not be empty because VariableLengthParams is set to true", i+1, name)
			}
		}
	}
}

func TestExprBuiltinGlobalVariableTypes(t *testing.T) {
	for name, g := range BuiltinGlobalVariableTypes {
		if name != g.Name {
			t.Errorf("name of global variable is different from its key: name=%q vs key=%q", g.Name, name)
		}
	}
}

func TestExprSemanticsCheckError(t *testing.T) {
	testCases := []struct {
		what     string
		input    string
		expected []string
		funcs    map[string][]*FuncSignature
		matrix   *ObjectType
		steps    *ObjectType
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
							Elem: &ObjectType{
								Props: map[string]ExprType{
									"foo": BoolType{},
								},
								StrictProps: true,
							},
						},
					},
				},
			},
		},
		{
			what:  "array element is not object for filtering array dereference",
			input: "test().*.bar",
			expected: []string{
				"object proprety filter \"bar\" of array element dereference must be type of object but got \"string\"",
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
				"receiver of array element dereference must be type of array but got \"bool\"",
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
				"number of arguments is wrong. function \"contains(string, string) -> bool\" takes 2 parameters but 1 arguments are provided",
				"number of arguments is wrong. function \"contains(array<any>, any) -> bool\" takes 2 parameters but 1 arguments are provided",
			},
		},
		{
			what:  "wrong number of arguments at function call for variable length parameters",
			input: "hashFiles()",
			expected: []string{
				"number of arguments is wrong. function \"hashFiles(string...) -> string\" takes at least 1 parameters but 0 arguments are provided",
			},
		},
		{
			what:  "wrong type at parameter",
			input: "startsWith('foo', 10)",
			expected: []string{
				"2nd argument of function call is not assignable. \"number\" cannot be assigned to \"string\"",
			},
		},
		{
			what:  "wrong type at parameter of overloaded function",
			input: "contains('foo', 10)",
			expected: []string{
				"2nd argument of function call is not assignable. \"number\" cannot be assigned to \"string\"",
				"1st argument of function call is not assignable. \"string\" cannot be assigned to \"array<any>\"",
			},
		},
		{
			what:  "wrong type at rest parameter",
			input: "hashFiles(10)",
			expected: []string{
				"1st argument of function call is not assignable. \"number\" cannot be assigned to \"string\"",
			},
		},
		{
			what:  "wrong type at rest parameter part2",
			input: "hashFiles('foo', 10)",
			expected: []string{
				"2nd argument of function call is not assignable. \"number\" cannot be assigned to \"string\"",
			},
		},
		{
			what:  "less arguments for format() builtin function call",
			input: "format('format {0} {1} {2}', 'foo')",
			expected: []string{
				"format string \"format {0} {1} {2}\" contains 3 placeholders but 1 arguments are given to format",
			},
		},
		{
			what:  "more arguments for format() builtin function call",
			input: "format('format {0} {1} {2}', 'foo', 1, true, null)",
			expected: []string{
				"format string \"format {0} {1} {2}\" contains 3 placeholders but 4 arguments are given to format",
			},
		},
		{
			what:  "operand of ! operator is not bool",
			input: "!42",
			expected: []string{
				"type of operand of ! operator \"number\" is not assignable to type \"bool\"",
			},
		},
		{
			what:  "left operand of && operator is not bool",
			input: "42 && true",
			expected: []string{
				"type of left operand of && operator \"number\" is not assignable to type \"bool\"",
			},
		},
		{
			what:  "right operand of && operator is not bool",
			input: "true && 42",
			expected: []string{
				"type of right operand of && operator \"number\" is not assignable to type \"bool\"",
			},
		},
		{
			what:  "left operand of || operator is not bool",
			input: "42 || true",
			expected: []string{
				"type of left operand of || operator \"number\" is not assignable to type \"bool\"",
			},
		},
		{
			what:  "right operand of || operator is not bool",
			input: "true || 42",
			expected: []string{
				"type of right operand of || operator \"number\" is not assignable to type \"bool\"",
			},
		},
		{
			what:  "undefined matrix value",
			input: "matrix.bar",
			expected: []string{
				"property \"bar\" is not defined in object type {foo: any}",
			},
			matrix: &ObjectType{
				Props: map[string]ExprType{
					"foo": AnyType{},
				},
				StrictProps: true,
			},
		},
		{
			what:  "type mismatch in matrix value",
			input: "startsWith('hello', matrix.foo)",
			expected: []string{
				"2nd argument of function call is not assignable. \"null\" cannot be assigned to \"string\"",
			},
			matrix: &ObjectType{
				Props: map[string]ExprType{
					"foo": NullType{},
				},
				StrictProps: true,
			},
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
			steps: &ObjectType{
				Props: map[string]ExprType{
					"bar": &ObjectType{
						Props: map[string]ExprType{
							"outputs":    NewObjectType(),
							"conclusion": StringType{},
							"outcome":    StringType{},
						},
						StrictProps: true,
					},
				},
				StrictProps: true,
			},
		},
		{
			what:  "invalid property in step object",
			input: "steps.bar.foo",
			expected: []string{
				"property \"foo\" is not defined in object type {", // order of prop types in object type changes randomly so we cannot check it easily
			},
			steps: &ObjectType{
				Props: map[string]ExprType{
					"bar": &ObjectType{
						Props: map[string]ExprType{
							"outputs":    NewObjectType(),
							"conclusion": StringType{},
							"outcome":    StringType{},
						},
						StrictProps: true,
					},
				},
				StrictProps: true,
			},
		},
		{
			what:  "step output value without typed steps outputs",
			input: "steps.foo.outputs",
			expected: []string{
				"property \"foo\" is not defined in object type {}",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			l := NewExprLexer()
			tok, _, err := l.Lex(tc.input + "}}")
			if err != nil {
				t.Fatal("Lex error:", err)
			}

			p := NewExprParser()
			e, err := p.Parse(tok)
			if err != nil {
				t.Fatal("Parse error:", tc.input)
			}

			c := NewExprSemanticsChecker()
			if tc.funcs != nil {
				c.funcs = tc.funcs // Set functions for testing
			}
			if tc.matrix != nil {
				c.UpdateMatrix(tc.matrix)
			}
			if tc.steps != nil {
				c.UpdateSteps(tc.steps)
			}
			_, errs := c.Check(e)
			if len(errs) != len(tc.expected) {
				t.Fatalf("semantics check should report %d errors but got %d errors %#v", len(tc.expected), len(errs), errs)
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
