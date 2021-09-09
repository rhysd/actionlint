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
	"github.event.discussion.title",
	"github.event.discussion.body",
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
			if k == "*" {
				if len(m) != 1 {
					t.Errorf("%v has * key but it also has other keys in %v", k, p)
				}
			} else if !re.MatchString(k) {
				t.Errorf("%v does not match to ^[a-z_]+$ in %v", k, p)
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

func testRunTrustedInputsCheckerForNode(t *testing.T, c *UntrustedInputChecker, input string) {
	n, err := NewExprParser().Parse(NewExprLexer(input + "}}"))
	if err != nil {
		t.Fatal(err)
	}
	VisitExprNode(n, func(n, p ExprNode, entering bool) {
		if !entering {
			c.OnNodeLeave(n)
		}
	})
}

func TestExprInsecureDetectUntrustedValue(t *testing.T) {
	type testCase struct {
		input string
		want  []string
	}

	tests := []testCase{}
	for _, input := range testAllUntrustedInputs {
		tests = append(tests, testCase{input, []string{input}})
		props := strings.Split(input, ".")
		{
			// github.foo.*.bar -> github['foo'].*['bar']
			var b strings.Builder
			b.WriteString(props[0])
			for _, p := range props[1:] {
				if p == "*" {
					b.WriteString(".*")
				} else {
					b.WriteString("['")
					b.WriteString(p)
					b.WriteString("']")
				}
			}
			tests = append(tests, testCase{b.String(), []string{input}})
		}

		if strings.Contains(input, ".*") {
			// Add both array dereference version and array index access version
			tests = append(tests, testCase{strings.ReplaceAll(input, ".*", "[0]"), []string{input}})
			{
				// github.foo.*.bar -> github['foo'][0]['bar']
				var b strings.Builder
				b.WriteString(props[0])
				for _, p := range props[1:] {
					if p == "*" {
						b.WriteString("[0]")
					} else {
						b.WriteString("['")
						b.WriteString(p)
						b.WriteString("']")
					}
				}
				tests = append(tests, testCase{b.String(), []string{input}})
			}
		}
	}

	tests = append(tests,
		testCase{
			"github.event.issue.body || github.event.issue.title",
			[]string{
				"github.event.issue.body",
				"github.event.issue.title",
			},
		},
		testCase{
			"github.event.issue.body.foo.bar",
			[]string{
				"github.event.issue.body",
			},
		},
		testCase{
			"github.event.issue.body[0]",
			[]string{
				"github.event.issue.body",
			},
		},
		testCase{
			"matrix.foo[github.event.issue.title].bar[github.event.issue.body]",
			[]string{
				"github.event.issue.body",
				"github.event.issue.title",
			},
		},
		testCase{
			"github.event.pages[github.event.issue.title].page_name",
			[]string{
				"github.event.issue.title",
				"github.event.pages.*.page_name",
			},
		},
		testCase{
			"github.event.pages[foo[github.event.issue.title]].page_name",
			[]string{
				"github.event.issue.title",
				"github.event.pages.*.page_name",
			},
		},
		testCase{
			"github.event.issue.body[github.event.issue.title][github.head_ref]",
			[]string{
				"github.head_ref",
				"github.event.issue.title",
				"github.event.issue.body",
			},
		},
		testCase{
			"github.event.pages[format('0')].page_name",
			[]string{
				"github.event.pages.*.page_name",
			},
		},
		testCase{
			"github.event.pages[matrix.page_num].page_name",
			[]string{
				"github.event.pages.*.page_name",
			},
		},
		testCase{
			"github.event.pages[github.event.commits[github.event.issue.title].author.name].page_name",
			[]string{
				"github.event.issue.title",
				"github.event.commits.*.author.name",
				"github.event.pages.*.page_name",
			},
		},
		testCase{
			"github.event.pages[format('{0}', github.event.issue.title)].page_name",
			[]string{
				"github.event.issue.title",
				"github.event.pages.*.page_name",
			},
		},
		testCase{
			"contains(github.event.pages.*.page_name.*.foo, github.event.issue.title)",
			[]string{
				"github.event.pages.*.page_name",
				"github.event.issue.title",
			},
		},
	)

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
			testRunTrustedInputsCheckerForNode(t, c, tc.input)
			errs := c.Errs()
			if len(tc.want) != len(errs) {
				t.Fatalf("wanted %d error(s) but got %v", len(tc.want), errs)
			}
			for i, err := range errs {
				want := tc.want[i]
				if !strings.Contains(err.Message, want) {
					t.Errorf("%q is not contained in error message: %v", want, err)
				}
				if err.Line != 1 {
					t.Error("line should be 1 but got", err.Line)
				}
			}
		})
	}
}

func TestExprInsecureAllUntrustedValuesAtOnce(t *testing.T) {
	args := make([]string, 0, len(testAllUntrustedInputs))
	for _, i := range testAllUntrustedInputs {
		args = append(args, i)
		if strings.Contains(i, ".*") {
			args = append(args, strings.ReplaceAll(i, ".*", "[0]")) // also add index access version
		}
	}
	// Generate function call with all untrusted inputs as its arguments
	expr := "someFunc(" + strings.Join(args, ", ") + ")"

	c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
	testRunTrustedInputsCheckerForNode(t, c, expr)
	errs := c.Errs()

	if len(errs) != len(args) {
		t.Fatalf("# of args %d v.s. # of errs %d. errs: %v", len(args), len(errs), errs)
	}

	col := 0
	for i, err := range errs {
		arg := args[i]
		if !strings.Contains(arg, "[0]") && !strings.Contains(err.Message, arg) {
			t.Errorf("%q is not contained in error: %v", arg, err)
		}
		if err.Line != 1 {
			t.Error("line should be 1 but got", err.Line)
		}
		if err.Column <= col {
			t.Errorf("column should be greater than %d but got %d", col, err.Column)
		}
		col = err.Column
	}
}

func TestExprInsecureInitState(t *testing.T) {
	c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
	testRunTrustedInputsCheckerForNode(t, c, "github.event.issue.title")
	if len(c.Errs()) == 0 {
		t.Fatal("no error occurred")
	}

	c.Init()
	if len(c.Errs()) != 0 {
		t.Fatal(c.Errs())
	}

	testRunTrustedInputsCheckerForNode(t, c, "github.event.issue.title")
	if len(c.Errs()) == 0 {
		t.Fatal("no error occurred")
	}
}

func TestExprInsecureNoUntrustedValue(t *testing.T) {
	inputs := []string{
		"0",
		"true",
		"null",
		"'github.event.issue.title'",
		"matrix.foo",
		"matrix.github.event.issue.title",
		"matrix.event.issue.title",
		"github",
		"github.event.issue",
		"github.event.commits.foo.message",
		"github.event.commits[0]",
		"github.event.commits.*",
		"github.event.commits.*.foo",
		"github.event.foo.body",
		"github[x].issue.title",
		"github.event[foo].title",
		"github.event.issue[0].title",
		"foo(github.event, pull_request.body)",
		"foo(github.event, github.pull_request.body)",
		"github.event[pull_request.body]",
		"github[event.pull_request].body",
		"github[github.event.pull_request].body",
		"github.event.*.body",
		"matrix.foo[github.event.pages].page_name",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
			testRunTrustedInputsCheckerForNode(t, c, input)
			if errs := c.Errs(); len(errs) > 0 {
				t.Fatalf("%d error(s) occurred: %v", len(errs), errs)
			}
		})
	}
}

func TestExprInsecureCustomizedUntrustedInputMapping(t *testing.T) {
	testCases := []struct {
		mapping UntrustedInputMap
		input   string
		want    string
	}{
		{
			mapping: UntrustedInputMap{
				"github": nil,
			},
			input: "github.event.issue.title",
			want:  `"github"`,
		},
		{
			mapping: UntrustedInputMap{
				"github": {
					"foo": {
						"*": nil,
					},
				},
			},
			input: "github.foo[0]",
			want:  `"github.foo.*"`,
		},
		{
			mapping: UntrustedInputMap{
				"github": {
					"foo": {
						"*": nil,
					},
				},
			},
			input: "github.foo.*",
			want:  `"github.foo.*"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			c := NewUntrustedInputChecker(tc.mapping)
			testRunTrustedInputsCheckerForNode(t, c, tc.input)
			errs := c.Errs()
			if len(errs) != 1 {
				t.Fatalf("1 error was wanted but got %d error(s)", len(errs))
			}
			err := errs[0]
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%q was wanted to be contained in error message %q", tc.want, err.Error())
			}
		})
	}
}
