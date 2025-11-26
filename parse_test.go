package actionlint

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	small, err := os.ReadFile(filepath.Join("testdata", "bench", "small.yaml"))
	if err != nil {
		b.Fatal(err)
	}
	large, err := os.ReadFile(filepath.Join("testdata", "bench", "large.yaml"))
	if err != nil {
		b.Fatal(err)
	}
	minimal, err := os.ReadFile(filepath.Join("testdata", "bench", "minimal.yaml"))
	if err != nil {
		b.Fatal(err)
	}

	benches := []struct {
		name  string
		input []byte
	}{
		{"minimal", minimal},
		{"small", small},
		{"large", large},
	}

	for _, bench := range benches {
		in := bench.input
		b.Run(bench.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, errs := Parse(in); len(errs) > 0 {
					b.Fatal(errs)
				}
			}
		})
	}
}
