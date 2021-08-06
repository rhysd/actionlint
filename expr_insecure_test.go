package actionlint

import (
	"regexp"
	"strings"
	"testing"
)

// From https://securitylab.github.com/research/github-actions-untrusted-input/
var testAllUntrustedInputs = []string{
	"github.event.issue.title",
	"github.event.issue.body",
	"github.event.pull_request.title",
	"github.event.pull_request.body",
	"github.event.comment.body",
	"github.event.review.body",
	"github.event.review_comment.body",
	"github.event.pages.*.page_name",
	"github.event.commits.*.message",
	"github.event.head_commit.message",
	"github.event.head_commit.author.email",
	"github.event.head_commit.author.name",
	"github.event.commits.*.author.email",
	"github.event.commits.*.author.name",
	"github.event.pull_request.head.ref",
	"github.event.pull_request.head.label",
	"github.event.pull_request.head.repo.default_branch",
	"github.head_ref",
}

func TestExprInsecureBuiltinUntrustedInputs(t *testing.T) {
	for _, input := range testAllUntrustedInputs {
		cur := BuiltinUntrustedInputs
		for _, name := range strings.Split(input, ".") {
			if m, ok := cur[name]; ok {
				cur = m
				continue
			}
			if m, ok := cur["*"]; ok {
				cur = m
				continue
			}
			t.Fatalf("%s in %s does not match to builtin untrusted inputs map: %v", name, input, cur)
		}
		if cur != nil {
			t.Fatalf("%s does not reach end of builtin untrusted inputs map: %v", input, cur)
		}
	}

	re := regexp.MustCompile(`^[a-z_]+$`)
	var rec func(m UntrustedInputMap, path []string)
	rec = func(m UntrustedInputMap, path []string) {
		for k, v := range m {
			p := append(path, k)
			if k != "*" && !re.MatchString(k) {
				t.Errorf("%v does not match to ^[a-z_]+$", k)
			}
			if v != nil {
				if len(v) == 0 {
					t.Errorf("%v must not be empty. use nil for end of branch", p)
				}
				rec(v, p)
			}
		}
	}

	rec(BuiltinUntrustedInputs, []string{})
}

func testCheckUntrustedInput(t *testing.T, input string) []*ExprError {
	expr, err := NewExprParser().Parse(NewExprLexer(input + "}}"))
	if err != nil {
		t.Fatal(err)
	}
	c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
	VisitExprNode(expr, func(n ExprNode, entering bool) {
		if entering {
			return
		}
		c.OnNode(n)
	})
	return c.Errs()
}

func TestExprInsecureDetectUntrustedValueAccess(t *testing.T) {
	for _, input := range testAllUntrustedInputs {
		t.Run(input, func(t *testing.T) {
			access := strings.ReplaceAll(input, "*", "foo")
			errs := testCheckUntrustedInput(t, access)
			if len(errs) != 1 {
				t.Fatalf("wanted only one error but got %v", errs)
			}
			e := errs[0]
			if !strings.Contains(e.Message, access) {
				t.Errorf("unexpected error: %v", e)
			}
			if e.Line != 1 || e.Column != 1 {
				t.Errorf("position is unexpected. wanted (1, 1) but got (%d, %d)", e.Line, e.Column)
			}
		})
	}
}

func TestExprInsecureDetectMultipleUntrustedValues(t *testing.T) {
	args := make([]string, 0, len(testAllUntrustedInputs))
	for _, i := range testAllUntrustedInputs {
		args = append(args, strings.ReplaceAll(i, "*", "foo"))
	}
	expr := "someFunc(" + strings.Join(args, ", ") + ")"
	errs := testCheckUntrustedInput(t, expr)
	if len(errs) != len(args) {
		t.Fatalf("# of args %d v.s. # of errs %d. errs: %v", len(args), len(errs), errs)
	}
	for i, err := range errs {
		arg := args[i]
		if !strings.Contains(err.Message, arg) {
			t.Errorf("%q is not contained in error: %v", arg, err)
		}
	}
}

func TestExprInsecureDetectNoUntrustedValue(t *testing.T) {
	inputs := []string{
		"0",
		"'foo'",
		"matrix.foo",
		"matrix.github.event.issue.title",
		"matrix.event.issue.title",
		"github",
		"github.event.issue",
		"github.event.commits.foo",
		"github.event.foo.body",
		"github[x].issue.title",
		"foo(github.event, pull_request.body)",
		"foo(github.event, github.pull_request.body)",
		// TODO: More tests
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			errs := testCheckUntrustedInput(t, input)
			if len(errs) > 0 {
				t.Fatalf("%d error(s) occurred: %v", len(errs), errs)
			}
		})
	}
}

// TODO: TestExprInsecureDetectUntrustedValueInExpr
