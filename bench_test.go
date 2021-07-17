package actionlint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/execabs"
)

func BenchmarkLintFilesExamples(b *testing.B) {
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

func BenchmarkLintManyScripts(b *testing.B) {
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
		what  string
		files []string
	}{
		{
			"small",
			[]string{small},
		},
		{
			"small",
			[]string{small, small, small},
		},
		{
			"small",
			[]string{small, small, small, small, small, small, small, small, small, small},
		},
		{
			"large",
			[]string{large},
		},
		{
			"large",
			[]string{large, large, large},
		},
		{
			"large",
			[]string{large, large, large, large, large, large, large, large, large, large},
		},
	}

	for _, bm := range bms {
		b.Run(fmt.Sprintf("%s-%d", bm.what, len(bm.files)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				opts := LinterOptions{
					Shellcheck: shellcheck,
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
