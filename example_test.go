package actionlint

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"golang.org/x/sys/execabs"
)

func testExampleFilePaths() (string, []string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	dir := filepath.Join(wd, "testdata", "examples")

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

func TestExamples(t *testing.T) {
	dir, infiles, err := testExampleFilePaths()
	if err != nil {
		panic(err)
	}

	proj := &Project{root: dir}

	for _, infile := range infiles {
		base := strings.TrimSuffix(infile, filepath.Ext(infile))
		outfile := base + ".out"
		testName := filepath.Base(base)
		t.Run(testName, func(t *testing.T) {
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

func BenchmarkExamplesLintFiles(b *testing.B) {
	dir, files, err := testExampleFilePaths()
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
