package actionlint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/execabs"
)

func BenchmarkLintWorkflows(b *testing.B) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	large := filepath.Join(dir, "testdata", "bench", "many_scripts.yaml")
	small := filepath.Join(dir, "testdata", "bench", "small.yaml")
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
