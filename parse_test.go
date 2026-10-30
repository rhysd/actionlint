package actionlint

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkParseWorkflow(b *testing.B) {
	type bench struct {
		name  string
		input []byte
	}

	loadBench := func(name string) bench {
		i, err := os.ReadFile(filepath.Join("testdata", "bench", name+".yaml"))
		if err != nil {
			b.Fatal(err)
		}
		return bench{name, i}
	}

	for _, bc := range []bench{
		loadBench("minimal"),
		loadBench("small"),
		loadBench("large"),
	} {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, errs := Parse(bc.input); len(errs) > 0 {
					b.Fatal(errs)
				}
			}
		})
	}
}

func BenchmarkParseTestData(b *testing.B) {
	type bench struct {
		name   string
		inputs [][]byte
	}

	loadBench := func(name string) bench {
		inputs := [][]byte{}
		_, fs, err := testFindAllWorkflowsInDir(name)
		if err != nil {
			b.Fatal(err)
		}
		for _, f := range fs {
			bs, err := os.ReadFile(f)
			if err != nil {
				b.Fatal(err)
			}
			inputs = append(inputs, bs)
		}
		return bench{name, inputs}
	}

	for _, bc := range []bench{
		loadBench("examples"),
		loadBench("ok"),
		loadBench("err"),
	} {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, in := range bc.inputs {
					// Note: Some workflows may cause parse error
					Parse(in)
				}
			}
		})
	}
}

func TestParseContainerVolumesAndPorts(t *testing.T) {
	input := []byte(`on: push
jobs:
  test:
    runs-on: ubuntu-latest
    container:
      image: node:18
      ports:
        - 8080
      volumes:
        - /foo:/bar
    services:
      redis:
        image: redis
        volumes:
          - /redis:/data
        ports:
          - 6379
    steps:
      - run: echo ok
`)

	w, errs := Parse(input)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	job := w.Jobs["test"]
	if job == nil {
		t.Fatal("job 'test' not found")
	}

	c := job.Container
	if c == nil {
		t.Fatal("job container is nil")
	}
	if len(c.Ports) != 1 || c.Ports[0].Value != "8080" {
		t.Fatalf("unexpected container ports: %v", c.Ports)
	}
	if len(c.Volumes) != 1 || c.Volumes[0].Value != "/foo:/bar" {
		t.Fatalf("unexpected container volumes: %v", c.Volumes)
	}

	if job.Services == nil || job.Services.Value["redis"] == nil {
		t.Fatal("service 'redis' not found")
	}
	s := job.Services.Value["redis"]
	if s.Container == nil {
		t.Fatal("service 'redis' container is nil")
	}
	sc := s.Container
	if len(sc.Ports) != 1 || sc.Ports[0].Value != "6379" {
		t.Fatalf("unexpected service ports: %v", sc.Ports)
	}
	if len(sc.Volumes) != 1 || sc.Volumes[0].Value != "/redis:/data" {
		t.Fatalf("unexpected service volumes: %v", sc.Volumes)
	}
}
