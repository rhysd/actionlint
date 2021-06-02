package actionlint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestExprSemanticsCheckOK(t *testing.T) {
	testCases := []struct {
		what     string
		input    string
		expected ExprType
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
			expected: NewObjectType(),
		},
		{
			what:     "object property dereference",
			input:    "github.issue.labels",
			expected: AnyType{},
		},
		{
			what:     "array element dereference",
			input:    "github.issue.labels.*.name",
			expected: &ArrayDerefType{Elem: AnyType{}},
		},
		{
			what:     "nested array element dereference",
			input:    "github.issues.*.labels.*.name",
			expected: &ArrayDerefType{Elem: AnyType{}},
		},
		{
			what:     "function call",
			input:    "contains('hello', 'll')",
			expected: BoolType{},
		},
		{
			what:     "function call overload",
			input:    "contains(github.issue.labels, 'foo')",
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
			input:    "github['issue']",
			expected: AnyType{},
		},
		{
			what:     "array element index access",
			input:    "github.issue.labels[0]",
			expected: AnyType{},
		},
		{
			what:     "index access to dereferenced array",
			input:    "github.issue.labels.*.name[0]",
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

func TestBuiltinFunctionSignatures(t *testing.T) {
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
