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

func BenchmarkLintFilesExamples(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(wd, "testdata", "examples")

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	files := []string{}
	for _, info := range entries {
		if info.IsDir() {
			continue
		}
		n := info.Name()
		if strings.HasSuffix(n, ".yaml") || strings.HasSuffix(n, ".yml") {
			files = append(files, filepath.Join(dir, n))
		}
	}

	proj := &Project{root: dir}
	shellcheck, err := execabs.LookPath("shellcheck")
	if err != nil {
		b.Fatalf("shellcheck is not found: %s", err)
	}
	pyflakes, err := execabs.LookPath("pyflakes")
	if err != nil {
		b.Fatalf("pyflakes is not found: %s", err)
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

	file := filepath.Join(dir, "testdata", "bench", "many_scripts.yaml")
	proj := &Project{root: dir}
	shellcheck, err := execabs.LookPath("shellcheck")
	if err != nil {
		b.Fatalf("shellcheck is not found: %s", err)
	}

	bms := [][]string{
		{file},
		{file, file, file},
		{file, file, file, file, file, file, file, file, file, file},
	}

	for _, bm := range bms {
		b.Run(fmt.Sprintf("%d", len(bm)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				opts := LinterOptions{
					Shellcheck: shellcheck,
				}

				l, err := NewLinter(ioutil.Discard, &opts)
				if err != nil {
					b.Fatal(err)
				}
				l.defaultConfig = &Config{}

				errs, err := l.LintFiles(bm, proj)
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
