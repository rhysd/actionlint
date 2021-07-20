package actionlint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/execabs"
)

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
