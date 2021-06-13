package actionlint

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type byErrorPosition []*Error

func (by byErrorPosition) Len() int {
	return len(by)
}

func (by byErrorPosition) Less(i, j int) bool {
	if by[i].Line == by[j].Line {
		return by[i].Column < by[j].Column
	}
	return by[i].Line < by[j].Line
}

func (by byErrorPosition) Swap(i, j int) {
	by[i], by[j] = by[j], by[i]
}

func TestExamples(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(wd, "testdata", "examples")

	infiles := []string{}

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			infiles = append(infiles, path)
		}
		return nil
	}); err != nil {
		panic(err)
	}

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
				p, err := exec.LookPath("shellcheck")
				if err != nil {
					t.Skip("skipped because \"shellcheck\" command does not exist in system")
				}
				opts.Shellcheck = p
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

			errs, err := linter.Lint("test.yaml", b, nil)
			if err != nil {
				t.Fatal(err)
			}

			if len(errs) != len(expected) {
				t.Fatalf("%d errors are expected but actually got %d errors: %# v", len(expected), len(errs), errs)
			}

			sort.Sort(byErrorPosition(errs))

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
