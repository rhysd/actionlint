package actionlint

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLexOneToken(t *testing.T) {
	testCases := []struct {
		what  string
		input string
		kind  TokenKind
	}{
		{
			what:  "identifier",
			input: "foo",
			kind:  TokenKindIdent,
		},
		{
			what:  "identifier with _",
			input: "foo_bar",
			kind:  TokenKindIdent,
		},
		{
			what:  "identifier with -",
			input: "foo_bar",
			kind:  TokenKindIdent,
		},
		{
			what:  "identifier with _ and -",
			input: "foo_bar-piyo",
			kind:  TokenKindIdent,
		},
		{
			what:  "_",
			input: "_",
			kind:  TokenKindIdent,
		},
		{
			what:  "_-",
			input: "_-",
			kind:  TokenKindIdent,
		},
		{
			what:  "null",
			input: "null",
			kind:  TokenKindIdent,
		},
		{
			what:  "bool",
			input: "true",
			kind:  TokenKindIdent,
		},
		{
			what:  "string",
			input: "'hello world'",
			kind:  TokenKindString,
		},
		{
			what:  "empty string",
			input: "''",
			kind:  TokenKindString,
		},
		{
			what:  "string with escapes",
			input: "'''hello''world'''",
			kind:  TokenKindString,
		},
		{
			what:  "string with braces",
			input: "'braces {in} string {{is}} ok!'",
			kind:  TokenKindString,
		},
		{
			what:  "string with non-ascii chars",
			input: "'こんにちは＼(^o^)／世界😊'",
			kind:  TokenKindString,
		},
		{
			what:  "int",
			input: "42",
			kind:  TokenKindInt,
		},
		{
			what:  "zero",
			input: "0",
			kind:  TokenKindInt,
		},
		{
			what:  "negative int",
			input: "-42",
			kind:  TokenKindInt,
		},
		{
			what:  "negative zero",
			input: "-0",
			kind:  TokenKindInt,
		},
		{
			what:  "hex int",
			input: "0x1e",
			kind:  TokenKindInt,
		},
		{
			what:  "hex int part2",
			input: "0xf",
			kind:  TokenKindInt,
		},
		{
			what:  "hex int part3",
			input: "0xa",
			kind:  TokenKindInt,
		},
		{
			what:  "negative hex int",
			input: "-0x1e",
			kind:  TokenKindInt,
		},
		{
			what:  "hex zero",
			input: "0x0",
			kind:  TokenKindInt,
		},
		{
			what:  "float",
			input: "1.0",
			kind:  TokenKindFloat,
		},
		{
			what:  "float smaller than 1",
			input: "0.123",
			kind:  TokenKindFloat,
		},
		{
			what:  "float zero",
			input: "0.0",
			kind:  TokenKindFloat,
		},
		{
			what:  "float exp part",
			input: "1.0e3",
			kind:  TokenKindFloat,
		},
		{
			what:  "float exp part with upper E",
			input: "1.0E3",
			kind:  TokenKindFloat,
		},
		{
			what:  "float negative exp part",
			input: "1.0e-99",
			kind:  TokenKindFloat,
		},
		{
			what:  "float negative exp part with upper E",
			input: "1.0E-99",
			kind:  TokenKindFloat,
		},
		{
			what:  "float zero with negative exp part",
			input: "0.0e-99",
			kind:  TokenKindFloat,
		},
		{
			what:  "int with exp part",
			input: "3e42",
			kind:  TokenKindFloat,
		},
		{
			what:  "int zero with exp part",
			input: "0e42",
			kind:  TokenKindFloat,
		},
		{
			what:  "int with negative exp part",
			input: "3e-9",
			kind:  TokenKindFloat,
		},
		{
			what:  "left paren",
			input: "(",
			kind:  TokenKindLeftParen,
		},
		{
			what:  "right paren",
			input: ")",
			kind:  TokenKindRightParen,
		},
		{
			what:  "left bracket",
			input: "[",
			kind:  TokenKindLeftBracket,
		},
		{
			what:  "right bracket",
			input: "]",
			kind:  TokenKindRightBracket,
		},
		{
			what:  "dot operator",
			input: ".",
			kind:  TokenKindDot,
		},
		{
			what:  "not operator",
			input: "!",
			kind:  TokenKindNot,
		},
		{
			what:  "less",
			input: "<",
			kind:  TokenKindLess,
		},
		{
			what:  "less equal",
			input: "<=",
			kind:  TokenKindLessEq,
		},
		{
			what:  "greater",
			input: ">",
			kind:  TokenKindGreater,
		},
		{
			what:  "greater equal",
			input: ">=",
			kind:  TokenKindGreaterEq,
		},
		{
			what:  "equal operator",
			input: "==",
			kind:  TokenKindEq,
		},
		{
			what:  "not equal operator",
			input: "!=",
			kind:  TokenKindNotEq,
		},
		{
			what:  "and operator",
			input: "&&",
			kind:  TokenKindAnd,
		},
		{
			what:  "or operator",
			input: "||",
			kind:  TokenKindOr,
		},
		{
			what:  "array access",
			input: "*",
			kind:  TokenKindStar,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			tokens, offset, err := LexExpression(tc.input + "}}")
			if err != nil {
				t.Fatal("error while lexing:", err)
			}
			if len(tokens) != 2 {
				t.Fatal("wanted token", tc.kind.String(), "followed by End token but got", tokens)
			}
			if tokens[1].Kind != TokenKindEnd {
				t.Fatal("wanted End token at end but got", tokens[1])
			}
			tok := tokens[0]
			if tok.Kind != tc.kind {
				t.Fatal("wanted token", tc.kind.String(), "but got", tok)
			}
			if tok.Value != tc.input {
				t.Fatalf("wanted token value %#v but got %#v", tc.input, tok.Value)
			}
			if offset != len(tc.input)+len("}}") {
				t.Fatal("wanted", len(tc.input)+len("}}"), "but got", offset, "tokens:", tokens)
			}
		})
	}
}

func TestLexExpression(t *testing.T) {
	testCases := []struct {
		what   string
		input  string
		tokens []TokenKind
		values []string
	}{
		{
			what:  "property dereference",
			input: "github.action_path",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindDot,
				TokenKindIdent,
			},
			values: []string{
				"github",
				".",
				"action_path",
			},
		},
		{
			what:  "property dereference with -",
			input: "job.services.foo-bar.id",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindDot,
				TokenKindIdent,
				TokenKindDot,
				TokenKindIdent,
				TokenKindDot,
				TokenKindIdent,
			},
			values: []string{
				"job",
				".",
				"services",
				".",
				"foo-bar",
				".",
				"id",
			},
		},
		{
			what:  "index syntax",
			input: "github['sha']",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindLeftBracket,
				TokenKindString,
				TokenKindRightBracket,
			},
			values: []string{
				"github",
				"[",
				"'sha'",
				"]",
			},
		},
		{
			what:  "array elements dereference",
			input: "labels.*.name",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindDot,
				TokenKindStar,
				TokenKindDot,
				TokenKindIdent,
			},
			values: []string{
				"labels",
				".",
				"*",
				".",
				"name",
			},
		},
		{
			what:  "startsWith",
			input: "startsWith('hello, world', 'hello')",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindLeftParen,
				TokenKindString,
				TokenKindComma,
				TokenKindString,
				TokenKindRightParen,
			},
			values: []string{
				"startsWith",
				"(",
				"'hello, world'",
				",",
				"'hello'",
				")",
			},
		},
		{
			what:  "join",
			input: "join(labels.*.name, ', ')",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindLeftParen,
				TokenKindIdent,
				TokenKindDot,
				TokenKindStar,
				TokenKindDot,
				TokenKindIdent,
				TokenKindComma,
				TokenKindString,
				TokenKindRightParen,
			},
			values: []string{
				"join",
				"(",
				"labels",
				".",
				"*",
				".",
				"name",
				",",
				"', '",
				")",
			},
		},
		{
			what:  "success",
			input: "success()",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindLeftParen,
				TokenKindRightParen,
			},
			values: []string{
				"success",
				"(",
				")",
			},
		},
		{
			what:  "operator twice",
			input: "!!success()",
			tokens: []TokenKind{
				TokenKindNot,
				TokenKindNot,
				TokenKindIdent,
				TokenKindLeftParen,
				TokenKindRightParen,
			},
			values: []string{
				"!",
				"!",
				"success",
				"(",
				")",
			},
		},
		{
			what:  "nested expression",
			input: "((a || b) && (c || d))",
			tokens: []TokenKind{
				TokenKindLeftParen,
				TokenKindLeftParen,
				TokenKindIdent,
				TokenKindOr,
				TokenKindIdent,
				TokenKindRightParen,
				TokenKindAnd,
				TokenKindLeftParen,
				TokenKindIdent,
				TokenKindOr,
				TokenKindIdent,
				TokenKindRightParen,
				TokenKindRightParen,
			},
			values: []string{
				"(",
				"(",
				"a",
				"||",
				"b",
				")",
				"&&",
				"(",
				"c",
				"||",
				"d",
				")",
				")",
			},
		},
		{
			what:  "equal expression",
			input: "0 == 1",
			tokens: []TokenKind{
				TokenKindInt,
				TokenKindEq,
				TokenKindInt,
			},
			values: []string{
				"0",
				"==",
				"1",
			},
		},
		{
			what: "skip white spaces",
			input: " 	a .b .c	x ( 'foo bar' ,	42	) [ true ]	",
			tokens: []TokenKind{
				TokenKindIdent,
				TokenKindDot,
				TokenKindIdent,
				TokenKindDot,
				TokenKindIdent,
				TokenKindIdent,
				TokenKindLeftParen,
				TokenKindString,
				TokenKindComma,
				TokenKindInt,
				TokenKindRightParen,
				TokenKindLeftBracket,
				TokenKindIdent,
				TokenKindRightBracket,
			},
			values: []string{
				"a",
				".",
				"b",
				".",
				"c",
				"x",
				"(",
				"'foo bar'",
				",",
				"42",
				")",
				"[",
				"true",
				"]",
			},
		},
		{
			what:   "empty expression",
			input:  "",
			tokens: []TokenKind{},
			values: []string{},
		},
	}

	for _, tc := range testCases {
		if len(tc.tokens) != len(tc.values) {
			panic(tc)
		}
		t.Run(tc.what, func(t *testing.T) {
			tokens, offset, err := LexExpression(tc.input + "}}")
			if err != nil {
				t.Fatal("error while lexing:", err)
			}
			if len(tokens) != len(tc.tokens)+1 {
				t.Fatal("wanted tokens", tc.tokens, "followed by End token but got", tokens)
			}
			last := tokens[len(tokens)-1]
			if last.Kind != TokenKindEnd {
				t.Fatal("wanted End token at end but got", last)
			}

			tokens = tokens[:len(tokens)-1]

			kinds := make([]TokenKind, 0, len(tokens))
			values := make([]string, 0, len(tokens))
			for _, t := range tokens {
				kinds = append(kinds, t.Kind)
				values = append(values, t.Value)
			}

			if !cmp.Equal(kinds, tc.tokens) {
				t.Errorf("wanted token kinds %#v but got %#v", tc.tokens, kinds)
			}
			if !cmp.Equal(values, tc.values) {
				t.Errorf("wanted values %#v but got %#v", tc.values, values)
			}

			if offset != len(tc.input)+len("}}") {
				t.Fatal("wanted offset", len(tc.input)+len("}}"), "but got", offset, "tokens:", tokens)
			}
		})
	}
}

func TestLexExprError(t *testing.T) {
	testCases := []struct {
		what  string
		input string
		want  string
	}{
		{
			what:  "unknown char",
			input: "?",
			want:  "unexpected character '?' while lexing expression",
		},
		{
			what:  "unexpected EOF",
			input: "42",
			want:  "unexpected EOF while lexing expression",
		},
		{
			what:  "empty string",
			input: "",
			want:  "unexpected EOF while lexing expression",
		},
		{
			what:  "broken string literal",
			input: "'foo bar",
			want:  "unexpected EOF while lexing end of string literal",
		},
		{
			what:  "broken string literal with escape",
			input: "'foo bar''",
			want:  "unexpected EOF while lexing end of string literal",
		},
		{
			what:  "invalid char after -",
			input: "-a",
			want:  "unexpected character 'a' while lexing number after -",
		},
		{
			what:  "EOF after -",
			input: "-",
			want:  "unexpected EOF while lexing number after -",
		},
		{
			what:  "invalid char after 0",
			input: "0d",
			want:  "unexpected character 'd' while lexing number after 0",
		},
		{
			what:  "invalid char in fraction part of float",
			input: "1.e1",
			want:  "unexpected character 'e' while lexing fraction part of float number",
		},
		{
			what:  "invalid char in exponent part of float",
			input: "1.0e_",
			want:  "unexpected character '_' while lexing exponent part of float number",
		},
		{
			what:  "no number after - in exponent part",
			input: "1.0e-",
			want:  "unexpected EOF while lexing exponent part of float number",
		},
		{
			what:  "invalid char in hex int",
			input: "0xg",
			want:  "unexpected character 'g' while lexing hex integer",
		},
		{
			what:  "unexpected EOF after 0x",
			input: "0x",
			want:  "unexpected EOF while lexing hex integer",
		},
		{
			what:  "invalid char at end of input",
			input: "'in {string} it {{is}} ok'}_",
			want:  "unexpected character '_' while lexing end marker",
		},
		{
			what:  "invalid char after =",
			input: "=3",
			want:  "unexpected character '3' while lexing == operator",
		},
		{
			what:  "invalid char after &",
			input: "&3",
			want:  "unexpected character '3' while lexing && operator",
		},
		{
			what:  "invalid char after |",
			input: "|3",
			want:  "unexpected character '3' while lexing || operator",
		},
		{
			what:  "unexpected EOF while lexing int",
			input: "0x",
			want:  "unexpected EOF while lexing hex integer",
		},
		{
			what:  "unexpected EOF while lexing fraction of float",
			input: "0.",
			want:  "unexpected EOF while lexing fraction part of float number",
		},
		{
			what:  "unexpected EOF while lexing exponent of float",
			input: "0.1e",
			want:  "unexpected EOF while lexing exponent part of float number",
		},
		{
			what:  "unexpected EOF while lexing end marker",
			input: "}",
			want:  "unexpected EOF while lexing end marker }}",
		},
		{
			what:  "unexpected EOF while lexing == operator",
			input: "=",
			want:  "unexpected EOF while lexing == operator",
		},
		{
			what:  "unexpected EOF while lexing && operator",
			input: "&",
			want:  "unexpected EOF while lexing && operator",
		},
		{
			what:  "unexpected EOF while lexing || operator",
			input: "|",
			want:  "unexpected EOF while lexing || operator",
		},
		{
			what:  "empty string",
			input: "",
			want:  "unexpected EOF while lexing expression",
		},
		{
			what:  "special note for string literals with double quotes",
			input: "\"hello\"",
			want:  "do you mean string literals? only single quotes are available for string delimiter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			_, _, err := LexExpression(tc.input)
			if err == nil {
				t.Fatal("error did not occur")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Error message %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestLexTokenPos(t *testing.T) {
	input := "foo( true && 0.1234, github.issue )"
	want := []int{0, 3, 5, 10, 13, 19, 21, 27, 28, 34, 35}

	ts, _, err := LexExpression(input + "}}")
	if err != nil {
		t.Fatal("error while lexing:", err)
	}

	if len(ts) != len(want) {
		t.Fatalf("length of inputs mismatch. want=%d, have=%d", len(want), len(ts))
	}

	for i := 0; i < len(ts); i++ {
		if ts[i].Offset != want[i] {
			t.Errorf("%dth token offsets mismatch. want=%d, have=%d", i+1, want[i], ts[i].Offset)
		}
		if ts[i].Line != 1 {
			t.Errorf("%dth token line should be 1 but got %d", i+1, ts[i].Line)
		}
		if ts[i].Column != want[i]+1 {
			t.Errorf("%dth token column should be %d but got %d", i+1, want[i]+1, ts[i].Column)
		}
	}
}
