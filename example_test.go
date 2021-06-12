package actionlint

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		t.Run(base, func(t *testing.T) {
			b, err := ioutil.ReadFile(infile)
			if err != nil {
				panic(err)
			}

			opts := LinterOptions{}
			linter, err := NewLinter(ioutil.Discard, &opts)
			if err != nil {
				t.Fatal(err)
			}

			config := Config{}
			linter.defaultConfig = &config

			errs, err := linter.Lint("test.yaml", b, nil)
			if err != nil {
				t.Fatal(err)
			}

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

			if len(errs) != len(expected) {
				t.Fatalf("%d errors are expected but actually got %d errors: %# v", len(expected), len(errs), errs)
			}

			for i := 0; i < len(errs); i++ {
				want, have := expected[i], errs[i].Error()
				if want != have {
					t.Errorf("error message mismatch at %dth error\n  want: %q\n  have: %q", i+1, want, have)
				}
			}
		})
	}
}
