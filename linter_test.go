package actionlint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func BenchmarkLintWorkflowFiles(b *testing.B) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	large := filepath.Join(dir, "testdata", "bench", "many_scripts.yaml")
	small := filepath.Join(dir, "testdata", "bench", "small.yaml")
	min := filepath.Join(dir, "testdata", "bench", "minimal.yaml")
	proj := &Project{root: dir}
	shellcheck, err := execabs.LookPath("shellcheck")
	if err != nil {
		b.Skipf("shellcheck is not found: %s", err)
	}

	bms := []struct {
		what       string
		files      []string
		shellcheck string
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
	}

	for _, bm := range bms {
		sc := ""
		if bm.shellcheck != "" {
			sc = "-shellcheck"
		}
		b.Run(fmt.Sprintf("%s%s-%d", bm.what, sc, len(bm.files)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				opts := LinterOptions{
					Shellcheck: bm.shellcheck,
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
