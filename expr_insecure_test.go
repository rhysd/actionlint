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
				cur = m.Children
				continue
			}
			if m, ok := cur["*"]; ok {
				cur = m.Children
				continue
			}
			t.Fatalf("%s in %s does not match to builtin untrusted inputs map: %v", name, input, cur)
		}
		if cur != nil {
			t.Fatalf("%s does not reach end of builtin untrusted inputs map: %v", input, cur)
		}
	}

	re := regexp.MustCompile(`^[a-z_]+$`)
	var rec func(m map[string]*UntrustedInputMap, path []string)
	rec = func(m map[string]*UntrustedInputMap, path []string) {
		for k, v := range m {
			p := append(path, k)
			if k == "*" {
				if len(m) != 1 {
					t.Errorf("%v has * key but it also has other keys in %v", k, p)
				}
			} else if !re.MatchString(k) {
				t.Errorf("%v does not match to ^[a-z_]+$ in %v", k, p)
			}
			rec(v.Children, p)
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
			c.OnVisitNodeLeave(n)
		}
	})
	c.OnVisitEnd()
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
			"contains(github.event.pages.*.page_name, github.event.issue.title)",
			[]string{
				"github.event.pages.*.page_name",
				"github.event.issue.title",
			},
		},
		testCase{
			"contains(github.event.*.body, github.event.*.*)",
			[]string{
				"github.event.",
				"github.event.",
			},
		},
	)

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
			testRunTrustedInputsCheckerForNode(t, c, tc.input)
			errs := c.Errs()
			if len(tc.want) != len(errs) {
				t.Fatalf("wanted %d error(s) but got %d error(s): %v", len(tc.want), len(errs), errs)
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
		//"github.event.commits.foo.message",
		"github.event.commits[0]",
		"github.event.commits.*",
		"github.event.commits.*.foo",
		"github.event.foo.body",
		"github[x].issue.title",
		"github.event[foo].title",
		"github.event.issue[0].title",
		"foo(github.event, pull_request.body)",
		"foo(github.event, github.pull_request.body)",
		"foo(github.event, bar().pull_request.body)",
		"github.event[pull_request.body]",
		"github[event.pull_request].body",
		"github[github.event.pull_request].body",
		"matrix.foo[github.event.pages].page_name",
		"github.event.issue.body.foo.bar",
		"github.event.issue.body[0]",
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
		mapping *UntrustedInputMap
		input   string
		want    string
	}{
		{
			mapping: NewUntrustedInputMap("foo"),
			input:   "foo",
			want:    `"foo"`,
		},
		{
			mapping: NewUntrustedInputMap("foo",
				NewUntrustedInputMap("bar",
					NewUntrustedInputMap("piyo"),
				),
			),
			input: "foo.bar.piyo",
			want:  `"foo.bar.piyo"`,
		},
		{
			mapping: NewUntrustedInputMap("github",
				NewUntrustedInputMap("foo",
					NewUntrustedInputMap("*"),
				),
			),
			input: "github.foo[0]",
			want:  `"github.foo.*"`,
		},
		{
			mapping: NewUntrustedInputMap("github",
				NewUntrustedInputMap("foo",
					NewUntrustedInputMap("*"),
				),
			),
			input: "github.foo.*",
			want:  `"github.foo.*"`,
		},
		{
			mapping: NewUntrustedInputMap("foo",
				NewUntrustedInputMap("bar",
					NewUntrustedInputMap("piyo"),
				),
			),
			input: "foo.*.piyo",
			want:  `"foo.bar.piyo"`,
		},
		{
			mapping: NewUntrustedInputMap("foo",
				NewUntrustedInputMap("bar",
					NewUntrustedInputMap("piyo"),
				),
			),
			input: "foo.*.piyo[0]",
			want:  `"foo.bar.piyo"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			roots := UntrustedInputSearchRoots{}
			roots.AddRoot(tc.mapping)
			c := NewUntrustedInputChecker(roots)
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

func TestExprInsecureDetectUntrustedObjectFiltering(t *testing.T) {
	tests := []struct {
		input    string
		detected []string
	}{
		{
			input: "github.event.*.body",
			detected: []string{
				"github.event.issue.body",
				"github.event.comment.body",
				"github.event.review.body",
				"github.event.review_comment.body",
				"github.event.pull_request.body",
			},
		},
		{
			input: "github.event.*.body[0]",
			detected: []string{
				"github.event.issue.body",
				"github.event.comment.body",
				"github.event.review.body",
				"github.event.review_comment.body",
				"github.event.pull_request.body",
			},
		},
		{
			input: "github.event.*.*.email",
			detected: []string{
				"github.event.head_commit.author.email",
			},
		},
		{
			input: "github.event.*.*.email[0]",
			detected: []string{
				"github.event.head_commit.author.email",
			},
		},
		{
			input: "github['event'].*.body",
			detected: []string{
				"github.event.issue.body",
				"github.event.comment.body",
				"github.event.review.body",
				"github.event.review_comment.body",
				"github.event.pull_request.body",
			},
		},
		{
			input: "github.event.*.*['email']",
			detected: []string{
				"github.event.head_commit.author.email",
			},
		},
		{
			input: "github.event.*.*['email'][0]",
			detected: []string{
				"github.event.head_commit.author.email",
			},
		},
		{
			input: "github.event.*.author.email",
			detected: []string{
				"github.event.head_commit.author.email",
			},
		},
		{
			input: "github.event.*['author']['email']",
			detected: []string{
				"github.event.head_commit.author.email",
			},
		},
		{
			input: "github.event.*.*.message",
			detected: []string{
				"github.event.commits.*.message",
			},
		},
		{
			input: "github.event.*.*.*",
			detected: []string{
				"github.event.commits.*.message",
				"github.event.head_commit.author.email",
				"github.event.head_commit.author.name",
			},
		},
		{
			input: "github.*",
			detected: []string{
				"github.head_ref",
			},
		},
		{
			input: "github.*[0]",
			detected: []string{
				"github.head_ref",
			},
		},
		{
			input: "github.*.*.body",
			detected: []string{
				"github.event.issue.body",
				"github.event.comment.body",
				"github.event.review.body",
				"github.event.review_comment.body",
				"github.event.pull_request.body",
			},
		},
		{
			input: "github.*.commits.*.message",
			detected: []string{
				"github.event.commits.*.message", // Second .* is for array
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
			testRunTrustedInputsCheckerForNode(t, c, tc.input)
			errs := c.Errs()
			if len(errs) != 1 {
				t.Fatalf("wanted 1 error but got %d error(s): %v", len(errs), errs)
			}
			msg := errs[0].Error()
			for _, want := range tc.detected {
				if !strings.Contains(msg, want) {
					t.Fatalf("error message did not include expected untrusted input %q: %q. input was %q", want, msg, tc.input)
				}
			}
		})
	}

	// Check all inputs with one checker
	c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
	for _, tc := range tests {
		c.Init()
		testRunTrustedInputsCheckerForNode(t, c, tc.input)
		errs := c.Errs()
		if len(errs) != 1 {
			t.Fatalf("%q: wanted 1 error but got %d error(s): %v", tc.input, len(errs), errs)
		}
		msg := errs[0].Error()
		for _, want := range tc.detected {
			if !strings.Contains(msg, want) {
				t.Fatalf("error message did not include expected untrusted input %q: %q. input was %q", want, msg, tc.input)
			}
		}
	}
}

func TestExprInsecureNoUntrustedObjectFiltering(t *testing.T) {
	inputs := []string{
		"github.*.foo",
		"github.*['foo']",
		"github.event.*.body.foo",
		//"github.event.*.body.*", // `['aaa', 'bbb'].*` is `[]`
		"github.*.*.foo",
		"github.*.commits.*.foo",
		"github.*['commits'].*.foo",
		"github.*.commits.*.message.foo",
		"a.*",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
			testRunTrustedInputsCheckerForNode(t, c, input)
			errs := c.Errs()
			if len(errs) != 0 {
				t.Fatalf("unexpected %d error(s): %v", len(errs), errs)
			}
		})
	}
}

func BenchmarkInsecureDetectUntrustedInputs(b *testing.B) {
	parseNodes := func(exprs []string) []ExprNode {
		ns := make([]ExprNode, 0, len(exprs))
		p := NewExprParser()
		for _, e := range exprs {
			n, err := p.Parse(NewExprLexer(e + "}}"))
			if err != nil {
				b.Fatal(err)
			}
			ns = append(ns, n)
		}
		return ns
	}

	untrustedExprs := []string{
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
		"github.event.issue.body || github.event.issue.title",
		"matrix.foo[github.event.issue.title].bar[github.event.issue.body]",
		"github.event.pages[github.event.issue.title].page_name",
		"github.event.pages[foo[github.event.issue.title]].page_name",
		"github.event.issue.body[github.event.issue.title][github.head_ref]",
		"github.event.pages[format('0')].page_name",
		"github.event.pages[matrix.page_num].page_name",
		"github.event.pages[github.event.commits[github.event.issue.title].author.name].page_name",
		"github.event.pages[format('{0}', github.event.issue.title)].page_name",
		"contains(github.event.pages.*.page_name, github.event.issue.title)",
		"github.event.*.body",
		"github.event.*.body[0]",
		"github.event.*.*.email",
		"github.event.*.*.email[0]",
		"github.event.*.author.email",
		"github.event.*.*",
		"github.*",
		"github.*.*.body",
		"github.*.commits.*.message",
	}
	untrustedNodes := parseNodes(untrustedExprs)

	b.Run("UntrustedInput", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
			for j, n := range untrustedNodes {
				c.Init()
				VisitExprNode(n, func(n, p ExprNode, entering bool) {
					if !entering {
						c.OnVisitNodeLeave(n)
					}
				})
				c.OnVisitEnd()
				errs := c.Errs()
				if len(errs) == 0 {
					b.Fatalf("no error detected: %q", untrustedExprs[j])
				}
			}
		}
	})

	trustedExprs := []string{
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
		"foo(github.event, bar().pull_request.body)",
		"github.event[pull_request.body]",
		"github[event.pull_request].body",
		"github[github.event.pull_request].body",
		"matrix.foo[github.event.pages].page_name",
		"github.event.issue.body.foo.bar",
		"github.event.issue.body[0]",
		"github.*.foo",
		"github.event.*.body.foo",
		"github.event.*.body.*",
		"github.*.*.foo",
		"github.*.commits.*.foo",
		"github.*.commits.*.message.foo",
		"a.*",
	}
	trustedNodes := parseNodes(trustedExprs)

	b.Run("NoUntrustedInput", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := NewUntrustedInputChecker(BuiltinUntrustedInputs)
			for j, n := range trustedNodes {
				c.Init()
				VisitExprNode(n, func(n, p ExprNode, entering bool) {
					if !entering {
						c.OnVisitNodeLeave(n)
					}
				})
				c.OnVisitEnd()
				errs := c.Errs()
				if len(errs) != 0 {
					b.Fatalf("error detected: %q: %v", trustedExprs[j], errs)
				}
			}
		}
	})
}
