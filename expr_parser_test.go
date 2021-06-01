package actionlint

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseExpressionSyntaxOK(t *testing.T) {
	testCases := []struct {
		what     string
		input    string
		expected ExprNode
	}{
		// simple expressions
		{
			what:     "null literal",
			input:    "null",
			expected: &NullNode{},
		},
		{
			what:     "boolean literal true",
			input:    "true",
			expected: &BoolNode{Value: true},
		},
		{
			what:     "boolean literal false",
			input:    "false",
			expected: &BoolNode{Value: false},
		},
		{
			what:     "integer literal",
			input:    "711",
			expected: &IntNode{Value: 711},
		},
		{
			what:     "negative integer literal",
			input:    "-10",
			expected: &IntNode{Value: -10},
		},
		{
			what:     "zero integer literal",
			input:    "0",
			expected: &IntNode{Value: 0},
		},
		{
			what:     "hex integer literal",
			input:    "0x1f",
			expected: &IntNode{Value: 0x1f},
		},
		{
			what:     "negative hex integer literal",
			input:    "-0xaf",
			expected: &IntNode{Value: -0xaf},
		},
		{
			what:     "hex integer zero",
			input:    "0x0",
			expected: &IntNode{Value: 0x0},
		},
		{
			what:     "float literal",
			input:    "1234.567",
			expected: &FloatNode{Value: 1234.567},
		},
		{
			what:     "float literal smaller than 1",
			input:    "0.567",
			expected: &FloatNode{Value: 0.567},
		},
		{
			what:     "float literal zero",
			input:    "0.0",
			expected: &FloatNode{Value: 0.0},
		},
		{
			what:     "negative float literal",
			input:    "-1234.567",
			expected: &FloatNode{Value: -1234.567},
		},
		{
			what:     "float literal with exponent part",
			input:    "12e3",
			expected: &FloatNode{Value: 12e3},
		},
		{
			what:     "float literal with negative exponent part",
			input:    "-99e-1",
			expected: &FloatNode{Value: -99e-1},
		},
		{
			what:     "float literal with fraction and exponent part",
			input:    "1.2e3",
			expected: &FloatNode{Value: 1.2e3},
		},
		{
			what:     "float literal with fraction and negative exponent part",
			input:    "-0.123e-12",
			expected: &FloatNode{Value: -0.123e-12},
		},
		{
			what:     "float zero value with exponent part",
			input:    "0e3",
			expected: &FloatNode{Value: 0e3},
		},
		{
			what:     "string literal",
			input:    "'hello, world'",
			expected: &StringNode{Value: "hello, world"},
		},
		{
			what:     "empty string literal",
			input:    "''",
			expected: &StringNode{Value: ""},
		},
		{
			what:     "string literal with escapes",
			input:    "'''hello''world'''",
			expected: &StringNode{Value: "'hello'world'"},
		},
		{
			what:     "string literal with non-ascii chars",
			input:    "'こんにちは＼(^o^)／世界😊'",
			expected: &StringNode{Value: "こんにちは＼(^o^)／世界😊"},
		},
		{
			what:     "variable",
			input:    "github",
			expected: &VariableNode{Name: "github"},
		},
		{
			what:  "func call",
			input: "success()",
			expected: &FuncCallNode{
				Callee: "success",
				Args:   []ExprNode{},
			},
		},
		{
			what:  "func call with 1 argument",
			input: "fromJSON(object)",
			expected: &FuncCallNode{
				Callee: "fromJSON",
				Args: []ExprNode{
					&VariableNode{Name: "object"},
				},
			},
		},
		{
			what:  "func call with multiple arguments",
			input: "contains('hello, world', 'o, w')",
			expected: &FuncCallNode{
				Callee: "contains",
				Args: []ExprNode{
					&StringNode{Value: "hello, world"},
					&StringNode{Value: "o, w"},
				},
			},
		},
		{
			what:  "index access",
			input: "obj['key']",
			expected: &IndexAccessNode{
				Operand: &VariableNode{Name: "obj"},
				Index:   &StringNode{Value: "key"},
			},
		},
		{
			what:  "index access with variable",
			input: "obj[a.b]",
			expected: &IndexAccessNode{
				Operand: &VariableNode{Name: "obj"},
				Index: &ObjectDerefNode{
					Receiver: &VariableNode{Name: "a"},
					Property: "b",
				},
			},
		},
		{
			what:  "< operator",
			input: "0 < 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindLess,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "<= operator",
			input: "0 <= 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindLessEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "> operator",
			input: "0 > 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindGreater,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  ">= operator",
			input: "0 >= 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindGreaterEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "== operator",
			input: "0 == 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "!= operator",
			input: "0 != 1",
			expected: &CompareOpNode{
				Kind:  CompareOpNodeKindNotEq,
				Left:  &IntNode{Value: 0},
				Right: &IntNode{Value: 1},
			},
		},
		{
			what:  "&& operator",
			input: "true && false",
			expected: &LogicalOpNode{
				Kind:  LogicalOpNodeKindAnd,
				Left:  &BoolNode{Value: true},
				Right: &BoolNode{Value: false},
			},
		},
		{
			what:  "|| operator",
			input: "true || false",
			expected: &LogicalOpNode{
				Kind:  LogicalOpNodeKindOr,
				Left:  &BoolNode{Value: true},
				Right: &BoolNode{Value: false},
			},
		},
		{
			what:     "nested value",
			input:    "(42)",
			expected: &IntNode{Value: 42},
		},
		{
			what:     "very nested value",
			input:    "((((((((((((((((((42))))))))))))))))))",
			expected: &IntNode{Value: 42},
		},
		{
			what:  "object property dereference",
			input: "a.b",
			expected: &ObjectDerefNode{
				Receiver: &VariableNode{Name: "a"},
				Property: "b",
			},
		},
		{
			what:  "nested object property dereference",
			input: "a.b.c.d",
			expected: &ObjectDerefNode{
				Property: "d",
				Receiver: &ObjectDerefNode{
					Property: "c",
					Receiver: &ObjectDerefNode{
						Property: "b",
						Receiver: &VariableNode{Name: "a"},
					},
				},
			},
		},
		{
			what:  "array element dereference",
			input: "a.*",
			expected: &ArrayDerefNode{
				Receiver: &VariableNode{Name: "a"},
			},
		},
		{
			what:  "nested array element dereference",
			input: "a.*.*.*",
			expected: &ArrayDerefNode{
				Receiver: &ArrayDerefNode{
					Receiver: &ArrayDerefNode{
						Receiver: &VariableNode{Name: "a"},
					},
				},
			},
		},
		// compound expressions
		{
			what:  "logical expressions",
			input: "0 == 0.1 && a < b || x >= !y && true != false",
			expected: &LogicalOpNode{
				Kind: LogicalOpNodeKindOr,
				Left: &LogicalOpNode{
					Kind: LogicalOpNodeKindAnd,
					Left: &CompareOpNode{
						Kind:  CompareOpNodeKindEq,
						Left:  &IntNode{Value: 0},
						Right: &FloatNode{Value: 0.1},
					},
					Right: &CompareOpNode{
						Kind:  CompareOpNodeKindLess,
						Left:  &VariableNode{Name: "a"},
						Right: &VariableNode{Name: "b"},
					},
				},
				Right: &LogicalOpNode{
					Kind: LogicalOpNodeKindAnd,
					Left: &CompareOpNode{
						Kind: CompareOpNodeKindGreaterEq,
						Left: &VariableNode{Name: "x"},
						Right: &NotOpNode{
							Operand: &VariableNode{Name: "y"},
						},
					},
					Right: &CompareOpNode{
						Kind:  CompareOpNodeKindNotEq,
						Left:  &BoolNode{Value: true},
						Right: &BoolNode{Value: false},
					},
				},
			},
		},
		{
			what:  "logical expressions with nested expressions",
			input: "(0 == 0.1) && (a < b || x >= !y) && (true != false)",
			expected: &LogicalOpNode{
				Kind: LogicalOpNodeKindAnd,
				Left: &CompareOpNode{
					Kind:  CompareOpNodeKindEq,
					Left:  &IntNode{Value: 0},
					Right: &FloatNode{Value: 0.1},
				},
				Right: &LogicalOpNode{
					Kind: LogicalOpNodeKindAnd,
					Left: &LogicalOpNode{
						Kind: LogicalOpNodeKindOr,
						Left: &CompareOpNode{
							Kind:  CompareOpNodeKindLess,
							Left:  &VariableNode{Name: "a"},
							Right: &VariableNode{Name: "b"},
						},
						Right: &CompareOpNode{
							Kind: CompareOpNodeKindGreaterEq,
							Left: &VariableNode{Name: "x"},
							Right: &NotOpNode{
								Operand: &VariableNode{Name: "y"},
							},
						},
					},
					Right: &CompareOpNode{
						Kind:  CompareOpNodeKindNotEq,
						Left:  &BoolNode{Value: true},
						Right: &BoolNode{Value: false},
					},
				},
			},
		},
		{
			what:  "logical expressions with more nested expressions",
			input: "((0 == 0.1) && (a < b || x >= !y)) && (true != false)",
			expected: &LogicalOpNode{
				Kind: LogicalOpNodeKindAnd,
				Left: &LogicalOpNode{
					Kind: LogicalOpNodeKindAnd,
					Left: &CompareOpNode{
						Kind:  CompareOpNodeKindEq,
						Left:  &IntNode{Value: 0},
						Right: &FloatNode{Value: 0.1},
					},
					Right: &LogicalOpNode{
						Kind: LogicalOpNodeKindOr,
						Left: &CompareOpNode{
							Kind:  CompareOpNodeKindLess,
							Left:  &VariableNode{Name: "a"},
							Right: &VariableNode{Name: "b"},
						},
						Right: &CompareOpNode{
							Kind: CompareOpNodeKindGreaterEq,
							Left: &VariableNode{Name: "x"},
							Right: &NotOpNode{
								Operand: &VariableNode{Name: "y"},
							},
						},
					},
				},
				Right: &CompareOpNode{
					Kind:  CompareOpNodeKindNotEq,
					Left:  &BoolNode{Value: true},
					Right: &BoolNode{Value: false},
				},
			},
		},
		{
			what:  "nested function calls",
			input: "!contains(some.value, 'foo') && endsWith(join(x.*.y, ', '), 'bar')",
			expected: &LogicalOpNode{
				Kind: LogicalOpNodeKindAnd,
				Left: &NotOpNode{
					Operand: &FuncCallNode{
						Callee: "contains",
						Args: []ExprNode{
							&ObjectDerefNode{
								Receiver: &VariableNode{Name: "some"},
								Property: "value",
							},
							&StringNode{Value: "foo"},
						},
					},
				},
				Right: &FuncCallNode{
					Callee: "endsWith",
					Args: []ExprNode{
						&FuncCallNode{
							Callee: "join",
							Args: []ExprNode{
								&ObjectDerefNode{
									Receiver: &ArrayDerefNode{
										Receiver: &VariableNode{Name: "x"},
									},
									Property: "y",
								},
								&StringNode{Value: ", "},
							},
						},
						&StringNode{Value: "bar"},
					},
				},
			},
		},
		{
			what:  "nested function calls with nested expressions",
			input: "!((contains((some.value), ('foo'))) && (endsWith((join((x.*.y), (', '))), ('bar'))))",
			expected: &NotOpNode{
				Operand: &LogicalOpNode{
					Kind: LogicalOpNodeKindAnd,
					Left: &FuncCallNode{
						Callee: "contains",
						Args: []ExprNode{
							&ObjectDerefNode{
								Receiver: &VariableNode{Name: "some"},
								Property: "value",
							},
							&StringNode{Value: "foo"},
						},
					},
					Right: &FuncCallNode{
						Callee: "endsWith",
						Args: []ExprNode{
							&FuncCallNode{
								Callee: "join",
								Args: []ExprNode{
									&ObjectDerefNode{
										Receiver: &ArrayDerefNode{
											Receiver: &VariableNode{Name: "x"},
										},
										Property: "y",
									},
									&StringNode{Value: ", "},
								},
							},
							&StringNode{Value: "bar"},
						},
					},
				},
			},
		},
		{
			what:  "nested dereferences",
			input: "contains(github.event['issue'].labels.*.name, 'bug')",
			expected: &FuncCallNode{
				Callee: "contains",
				Args: []ExprNode{
					&ObjectDerefNode{
						Property: "name",
						Receiver: &ArrayDerefNode{
							Receiver: &ObjectDerefNode{
								Property: "labels",
								Receiver: &IndexAccessNode{
									Index: &StringNode{Value: "issue"},
									Operand: &ObjectDerefNode{
										Property: "event",
										Receiver: &VariableNode{Name: "github"},
									},
								},
							},
						},
					},
					&StringNode{Value: "bug"},
				},
			},
		},
		{
			what:  "nested dereferences with nested expressions",
			input: "(((((github.event)['issue']).labels).*).name)",
			expected: &ObjectDerefNode{
				Property: "name",
				Receiver: &ArrayDerefNode{
					Receiver: &ObjectDerefNode{
						Property: "labels",
						Receiver: &IndexAccessNode{
							Index: &StringNode{Value: "issue"},
							Operand: &ObjectDerefNode{
								Property: "event",
								Receiver: &VariableNode{Name: "github"},
							},
						},
					},
				},
			},
		},
	}

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(VariableNode{}),
		cmpopts.IgnoreUnexported(NullNode{}),
		cmpopts.IgnoreUnexported(BoolNode{}),
		cmpopts.IgnoreUnexported(IntNode{}),
		cmpopts.IgnoreUnexported(FloatNode{}),
		cmpopts.IgnoreUnexported(StringNode{}),
		cmpopts.IgnoreUnexported(ObjectDerefNode{}),
		cmpopts.IgnoreUnexported(ArrayDerefNode{}),
		cmpopts.IgnoreUnexported(IndexAccessNode{}),
		cmpopts.IgnoreUnexported(NotOpNode{}),
		cmpopts.IgnoreUnexported(CompareOpNode{}),
		cmpopts.IgnoreUnexported(LogicalOpNode{}),
		cmpopts.IgnoreUnexported(FuncCallNode{}),
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			l := NewExprLexer()
			tok, _, err := l.Lex(tc.input + "}}")
			if err != nil {
				t.Fatal(err)
			}

			p := NewExprParser()
			n, err := p.Parse(tok)
			if err != nil {
				t.Fatal("Parse error:", err)
			}

			if !cmp.Equal(tc.expected, n, opts...) {
				t.Fatalf("wanted:\n%#v\n\nbut got:\n%#v\n\ndiff:\n%s\n", tc.expected, n, cmp.Diff(tc.expected, n, opts...))
			}
		})
	}
}

func TestParseExpressionSyntaxError(t *testing.T) {
	testCases := []struct {
		what     string
		input    string
		expected string
	}{
		{
			what:     "remaining inputs",
			input:    "42 foo bar",
			expected: "parser did not reach end of input after parsing expression",
		},
		{
			what:     "missing operand in || operator",
			input:    "true ||",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand in && operator",
			input:    "true &&",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand in < operator",
			input:    "0 <",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand in <= operator",
			input:    "0 <=",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand in > operator",
			input:    "0 >",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand in >= operator",
			input:    "0 >=",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand in == operator",
			input:    "0 ==",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand in != operator",
			input:    "0 !=",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand after ! operator",
			input:    "!",
			expected: "unexpected end of input",
		},
		{
			what:     "missing operand after . operator",
			input:    "foo.",
			expected: "unexpected end of input",
		},
		{
			what:     "ident must come after .",
			input:    "foo.42",
			expected: "unexpected token \"INTEGER\" while parsing object property dereference",
		},
		{
			what:     "broken index access part1",
			input:    "foo[0",
			expected: "unexpected end of input",
		},
		{
			what:     "broken index access part2",
			input:    "foo[",
			expected: "unexpected end of input",
		},
		{
			what:     "unexpected closing at index access",
			input:    "foo[0)",
			expected: "unexpected token \")\" while parsing closing bracket ']' for index access",
		},
		{
			what:     "starting with invalid token",
			input:    "[",
			expected: "unexpected token \"[\" while parsing variable access, function call, null, bool, int, float or string",
		},
		{
			what:     "missing closing ) for nested expression",
			input:    "(a",
			expected: "unexpected end of input",
		},
		{
			what:     "invalid token at closing nested expression",
			input:    "(a]",
			expected: "unexpected token \"]\" while parsing closing ')'",
		},
		{
			what:     "unexpected end of input while function call part1",
			input:    "foo(",
			expected: "unexpected end of input",
		},
		{
			what:     "unexpected end of input while function call part2",
			input:    "foo(0",
			expected: "unexpected end of input",
		},
		{
			what:     "unexpected end of input while function call part3",
			input:    "foo(0,",
			expected: "unexpected end of input",
		},
		{
			what:     "unexpected end of input while function call part4",
			input:    "foo(0, a",
			expected: "unexpected end of input",
		},
		{
			what:     "unexpected closing at function call",
			input:    "foo(0]",
			expected: "unexpected token \"]\" while parsing arguments of function call",
		},
		{
			what:     "error while parsing nested expression",
			input:    "([",
			expected: "unexpected token \"[\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			l := NewExprLexer()
			tok, _, err := l.Lex(tc.input + "}}")
			if err != nil {
				t.Fatal(err)
			}
			p := NewExprParser()
			_, err = p.Parse(tok)
			if err == nil {
				t.Fatal("Parse error did not occur:", tc.input)
			}

			if !strings.Contains(err.Error(), tc.expected) {
				t.Fatalf("error message %q does not contain expected string %q", err.Error(), tc.expected)
			}
		})
	}
}

func TestParseExpressionNumberLiteralsError(t *testing.T) {
	testCases := []struct {
		what string
		tok  *Token
	}{
		{
			what: "integer literal",
			tok: &Token{
				Kind:  TokenKindInt,
				Value: "abc",
			},
		},
		{
			what: "float literal",
			tok: &Token{
				Kind:  TokenKindFloat,
				Value: "abc",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			ts := []*Token{
				tc.tok,
				&Token{Kind: TokenKindEnd},
			}
			p := NewExprParser()
			_, err := p.Parse(ts)
			if err == nil {
				t.Fatal("Parse error did not occur:", tc.tok.Value)
			}
			want := fmt.Sprintf("parsing invalid %s", tc.what)
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error message %q does not contain %q", err.Error(), want)
			}
		})
	}
}
