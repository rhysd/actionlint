package actionlint

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sys/execabs"
)

func TestLinterLintOK(t *testing.T) {
	dir := filepath.Join("testdata", "ok")

	es, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	fs := make([]string, 0, len(es))
	for _, e := range es {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".yaml") || strings.HasSuffix(n, ".yml") {
			fs = append(fs, filepath.Join(dir, n))
		}
	}

	proj := &Project{root: dir}
	shellcheck, err := execabs.LookPath("shellcheck")
	if err != nil {
		t.Skip("skipped because \"shellcheck\" command does not exist in system")
	}

	pyflakes, err := execabs.LookPath("pyflakes")
	if err != nil {
		t.Skip("skipped because \"pyflakes\" command does not exist in system")
	}

	for _, f := range fs {
		t.Run(filepath.Base(f), func(t *testing.T) {
			opts := LinterOptions{
				Shellcheck: shellcheck,
				Pyflakes:   pyflakes,
			}

			linter, err := NewLinter(ioutil.Discard, &opts)
			if err != nil {
				t.Fatal(err)
			}

			config := Config{}
			linter.defaultConfig = &config

			errs, err := linter.LintFile(f, proj)
			if err != nil {
				t.Fatal(err)
			}
			if len(errs) > 0 {
				t.Fatal(errs)
			}
		})
	}
}

func testFindAllWorkflowsInDir(subdir string) (string, []string, error) {
	dir := filepath.Join("testdata", subdir)

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", nil, err
	}

	fs := make([]string, 0, len(entries))
	for _, info := range entries {
		if info.IsDir() {
			continue
		}
		n := info.Name()
		if strings.HasSuffix(n, ".yaml") || strings.HasSuffix(n, ".yml") {
			fs = append(fs, filepath.Join(dir, n))
		}
	}

	return dir, fs, nil
}

func TestLinterLintError(t *testing.T) {
	for _, subdir := range []string{"examples", "err"} {
		dir, infiles, err := testFindAllWorkflowsInDir(subdir)
		if err != nil {
			panic(err)
		}

		proj := &Project{root: dir}

		for _, infile := range infiles {
			base := strings.TrimSuffix(infile, filepath.Ext(infile))
			outfile := base + ".out"
			testName := filepath.Base(base)
			t.Run(subdir+"/"+testName, func(t *testing.T) {
				b, err := ioutil.ReadFile(infile)
				if err != nil {
					panic(err)
				}

				opts := LinterOptions{}

				if strings.Contains(testName, "shellcheck") {
					p, err := execabs.LookPath("shellcheck")
					if err != nil {
						t.Skip("skipped because \"shellcheck\" command does not exist in system")
					}
					opts.Shellcheck = p
				}

				if strings.Contains(testName, "pyflakes") {
					p, err := execabs.LookPath("pyflakes")
					if err != nil {
						t.Skip("skipped because \"pyflakes\" command does not exist in system")
					}
					opts.Pyflakes = p
				}

				linter, err := NewLinter(ioutil.Discard, &opts)
				if err != nil {
					t.Fatal(err)
				}

				config := Config{}
				linter.defaultConfig = &config

				expected := []string{}
				{
					f, err := os.Open(outfile)
					if err != nil {
						panic(err)
					}
					s := bufio.NewScanner(f)
					for s.Scan() {
						expected = append(expected, s.Text())
					}
					if err := s.Err(); err != nil {
						panic(err)
					}
				}

				errs, err := linter.Lint("test.yaml", b, proj)
				if err != nil {
					t.Fatal(err)
				}

				if len(errs) != len(expected) {
					t.Fatalf("%d errors are expected but actually got %d errors: %# v", len(expected), len(errs), errs)
				}

				sort.Sort(ByErrorPosition(errs))

				for i := 0; i < len(errs); i++ {
					want, have := expected[i], errs[i].Error()
					if strings.HasPrefix(want, "/") && strings.HasSuffix(want, "/") {
						want := regexp.MustCompile(want[1 : len(want)-1])
						if !want.MatchString(have) {
							t.Errorf("error message mismatch at %dth error does not match to regular expression\n  want: /%s/\n  have: %q", i+1, want, have)
						}
					} else {
						if want != have {
							t.Errorf("error message mismatch at %dth error does not match exactly\n  want: %q\n  have: %q", i+1, want, have)
						}
					}
				}
			})
		}
	}
}

func TestLinterFormatErrorMessageOK(t *testing.T) {
	tests := []struct {
		file   string
		format string
	}{
		{
			file:   "test.json",
			format: "{{json .}}",
		},
		{
			file:   "test.jsonl",
			format: "{{range $err := .}}{{json $err}}{{end}}",
		},
		{
			file:   "test.jsonl",
			format: "{{range $err := .}}{{json $err}}{{end}}",
		},
		{
			file:   "test.md",
			format: "{{range $ := .}}### Error at line {{$.Line}}, col {{$.Column}} of `{{$.Filepath}}`\\n\\n{{$.Message}}\\n\\n```\\n{{$.Snippet}}\\n```\\n\\n{{end}}",
		},
	}

	dir := filepath.Join("testdata", "format")
	proj := &Project{root: dir}
	infile := filepath.Join(dir, "test.yaml")
	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			opts := LinterOptions{Format: tc.format}

			var b strings.Builder
			l, err := NewLinter(&b, &opts)
			if err != nil {
				t.Fatal(err)
			}

			config := Config{}
			l.defaultConfig = &config
			errs, err := l.LintFile(infile, proj)
			if err != nil {
				t.Fatal(err)
			}
			if len(errs) == 0 {
				t.Fatal("no error")
			}

			buf, err := ioutil.ReadFile(filepath.Join(dir, tc.file))
			if err != nil {
				panic(err)
			}
			want := string(buf)

			have := b.String()
			// Fix path separators on Windows
			if runtime.GOOS == "windows" {
				slash := filepath.ToSlash(infile)
				have = strings.ReplaceAll(have, infile, slash)
				escaped := strings.ReplaceAll(slash, "/", `\\`)
				have = strings.ReplaceAll(have, escaped, slash)
			}

			if !cmp.Equal(want, have) {
				t.Fatal(cmp.Diff(want, have))
			}
		})
	}
}

func BenchmarkLintWorkflowFiles(b *testing.B) {
	large := filepath.Join("testdata", "bench", "many_scripts.yaml")
	small := filepath.Join("testdata", "bench", "small.yaml")
	min := filepath.Join("testdata", "bench", "minimal.yaml")
	proj := &Project{root: "."}
	shellcheck, err := execabs.LookPath("shellcheck")
	if err != nil {
		b.Skipf("shellcheck is not found: %s", err)
	}
	format := "{{range $ := .}}### Error at line {{$.Line}}, col {{$.Column}} of `{{$.Filepath}}`\\n\\n{{$.Message}}\\n\\n```\\n{{$.Snippet}}\\n```\\n\\n{{end}}"

	// Workflow files for this repository
	workflows := []string{}
	{
		d := filepath.Join(".github", "workflows")
		es, err := ioutil.ReadDir(d)
		if err != nil {
			panic(err)
		}
		for _, e := range es {
			workflows = append(workflows, filepath.Join(d, e.Name()))
		}
	}

	bms := []struct {
		what       string
		files      []string
		shellcheck string
		format     string
	}{
		{
			what:  "minimal",
			files: []string{min},
		},
		{
			what:  "minimal",
			files: []string{min, min, min, min, min},
		},
		{
			what:  "minimal",
			files: []string{min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min},
		},
		{
			what: "minimal",
			files: []string{
				min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min,
				min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min, min,
			},
		},
		{
			what:  "small",
			files: []string{small},
		},
		{
			what:  "small",
			files: []string{small, small, small},
		},
		{
			what:  "small",
			files: []string{small, small, small, small, small, small, small, small, small, small},
		},
		{
			what:       "small",
			files:      []string{small},
			shellcheck: shellcheck,
		},
		{
			what:       "small",
			files:      []string{small, small, small},
			shellcheck: shellcheck,
		},
		{
			what:       "small",
			files:      []string{small, small, small, small, small, small, small, small, small, small},
			shellcheck: shellcheck,
		},
		{
			what:       "large",
			files:      []string{large},
			shellcheck: shellcheck,
		},
		{
			what:       "large",
			files:      []string{large, large, large},
			shellcheck: shellcheck,
		},
		{
			what:       "large",
			files:      []string{large, large, large, large, large, large, large, large, large, large},
			shellcheck: shellcheck,
		},
		{
			what:   "small",
			files:  []string{small, small, small, small, small, small, small, small, small, small},
			format: format,
		},
		{
			what:  "our workflows",
			files: workflows,
		},
	}

	for _, bm := range bms {
		sc := ""
		if bm.shellcheck != "" {
			sc = "-shellcheck"
		}
		fm := ""
		if bm.format != "" {
			fm = "-format"
		}
		b.Run(fmt.Sprintf("%s%s%s-%d", bm.what, sc, fm, len(bm.files)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				opts := LinterOptions{
					Shellcheck: bm.shellcheck,
					Format:     bm.format,
				}

				l, err := NewLinter(ioutil.Discard, &opts)
				if err != nil {
					b.Fatal(err)
				}
				l.defaultConfig = &Config{}

				errs, err := l.LintFiles(bm.files, proj)
				if err != nil {
					b.Fatal(err)
				}

				if len(errs) > 0 {
					b.Fatal("some error occurred:", errs)
				}
			}
		})
	}
}

func BenchmarkLintWorkflowContent(b *testing.B) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	proj := &Project{root: dir}

	// Measure performance of traversing with checks except for external process rules (shellcheck, pyflakes)
	// Reading file content is not included in benchmark measurement.

	for _, name := range []string{"minimal", "small", "large"} {
		var f string
		switch name {
		case "minimal":
			f = filepath.Join(dir, "testdata", "bench", "minimal.yaml")
		case "small":
			f = filepath.Join(dir, "testdata", "bench", "small.yaml")
		case "large":
			f = filepath.Join(dir, "testdata", "bench", "many_scripts.yaml")
		}
		content, err := ioutil.ReadFile(f)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				opts := LinterOptions{}
				l, err := NewLinter(ioutil.Discard, &opts)
				if err != nil {
					b.Fatal(err)
				}
				l.defaultConfig = &Config{}
				errs, err := l.Lint(f, content, proj)
				if err != nil {
					b.Fatal(err)
				}
				if len(errs) > 0 {
					b.Fatal("some error occurred:", errs)
				}
			}
		})
	}
}

func BenchmarkExamplesLintFiles(b *testing.B) {
	dir, files, err := testFindAllWorkflowsInDir("examples")
	if err != nil {
		panic(err)
	}

	proj := &Project{root: dir}
	shellcheck, err := execabs.LookPath("shellcheck")
	if err != nil {
		b.Skipf("shellcheck is not found: %s", err)
	}
	pyflakes, err := execabs.LookPath("pyflakes")
	if err != nil {
		b.Skipf("pyflakes is not found: %s", err)
	}

	for i := 0; i < b.N; i++ {
		opts := LinterOptions{
			Shellcheck: shellcheck,
			Pyflakes:   pyflakes,
		}

		l, err := NewLinter(ioutil.Discard, &opts)
		if err != nil {
			b.Fatal(err)
		}
		l.defaultConfig = &Config{}

		errs, err := l.LintFiles(files, proj)
		if err != nil {
			b.Fatal(err)
		}

		if len(errs) == 0 {
			b.Fatal("no error found")
		}
	}
}

func BenchmarkLintRepository(b *testing.B) {
	for i := 0; i < b.N; i++ {
		opts := LinterOptions{}
		l, err := NewLinter(ioutil.Discard, &opts)
		if err != nil {
			b.Fatal(err)
		}
		errs, err := l.LintRepository(".")
		if err != nil {
			b.Fatal(err)
		}
		if len(errs) > 0 {
			b.Fatal(errs)
		}
	}
}
