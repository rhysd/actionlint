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
		cur := BuiltinUntrustedInputs2
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
	var rec func(m map[string]*UntrustedInputMap2, path []string)
	rec = func(m map[string]*UntrustedInputMap2, path []string) {
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

	rec(BuiltinUntrustedInputs2, []string{})
}

func testRunTrustedInputsCheckerForNode(t *testing.T, c *UntrustedInputChecker2, input string) {
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
	)

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			c := NewUntrustedInputChecker2(BuiltinUntrustedInputs2)
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

	c := NewUntrustedInputChecker2(BuiltinUntrustedInputs2)
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
	c := NewUntrustedInputChecker2(BuiltinUntrustedInputs2)
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
		"foo(github.event, bar().pull_request.body)",
		"github.event[pull_request.body]",
		"github[event.pull_request].body",
		"github[github.event.pull_request].body",
		"matrix.foo[github.event.pages].page_name",
		"github.event.issue.body.foo.bar",
		"github.event.issue.body[0]",
		// Object filter
		"github.event.*.foo",
		"github.*.foo",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := NewUntrustedInputChecker2(BuiltinUntrustedInputs2)
			testRunTrustedInputsCheckerForNode(t, c, input)
			if errs := c.Errs(); len(errs) > 0 {
				t.Fatalf("%d error(s) occurred: %v", len(errs), errs)
			}
		})
	}
}

func TestExprInsecureCustomizedUntrustedInputMapping(t *testing.T) {
	testCases := []struct {
		mapping *UntrustedInputMap2
		input   string
		want    string
	}{
		{
			mapping: NewUntrustedInputMap2("foo"),
			input:   "foo",
			want:    `"foo"`,
		},
		{
			mapping: NewUntrustedInputMap2("foo",
				NewUntrustedInputMap2("bar",
					NewUntrustedInputMap2("piyo"),
				),
			),
			input: "foo.bar.piyo",
			want:  `"foo.bar.piyo"`,
		},
		{
			mapping: NewUntrustedInputMap2("github",
				NewUntrustedInputMap2("foo",
					NewUntrustedInputMap2("*"),
				),
			),
			input: "github.foo[0]",
			want:  `"github.foo.*"`,
		},
		{
			mapping: NewUntrustedInputMap2("github",
				NewUntrustedInputMap2("foo",
					NewUntrustedInputMap2("*"),
				),
			),
			input: "github.foo.*",
			want:  `"github.foo.*"`,
		},
		{
			mapping: NewUntrustedInputMap2("foo",
				NewUntrustedInputMap2("bar",
					NewUntrustedInputMap2("piyo"),
				),
			),
			input: "foo.*.piyo",
			want:  `"foo.bar.piyo"`,
		},
		{
			mapping: NewUntrustedInputMap2("foo",
				NewUntrustedInputMap2("bar",
					NewUntrustedInputMap2("piyo"),
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
			c := NewUntrustedInputChecker2(roots)
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
	inputs := []string{
		"github.event.*.body",
		"github.event.*.body[0]",
		"github.event.*.*", // github.event.commits.*.message
		"github.*",
		"github.*[0]",
		"github.*.*.body",            // github.event.issue.body
		"github.*.commits.*.message", // github.event.commits.*.message: Second .* is for array
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := NewUntrustedInputChecker2(BuiltinUntrustedInputs2)
			testRunTrustedInputsCheckerForNode(t, c, input)
			errs := c.Errs()
			if len(errs) != 1 {
				t.Fatalf("wanted 1 error but got %d error(s): %v", len(errs), errs)
			}
			msg := errs[0].Error()

			pat := strings.ReplaceAll(input, ".", `\.`)
			pat = strings.ReplaceAll(pat, "*", `([0-9a-z_]+|\*)`)
			pat = strings.TrimSuffix(pat, "[0]") // Array index at the end means choosing one from object filtering results

			re := regexp.MustCompile(pat)
			if !re.MatchString(msg) {
				t.Fatalf("error message did not match to regex %q: %q. input was %q", pat, msg, input)
			}
		})
	}
}

func TestExprInsecureNoUntrustedObjectFiltering(t *testing.T) {
	inputs := []string{
		"github.*.foo",
		"github.event.*.body.foo",
		"github.event.*.body.*", // `['aaa', 'bbb'].*` is `[]`
		"github.*.*.foo",
		"github.*.commits.*.foo",
		"github.*.commits.*.message.foo",
		"a.*",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := NewUntrustedInputChecker2(BuiltinUntrustedInputs2)
			testRunTrustedInputsCheckerForNode(t, c, input)
			errs := c.Errs()
			if len(errs) != 0 {
				t.Fatalf("unexpected %d error(s): %v", len(errs), errs)
			}
		})
	}
}
