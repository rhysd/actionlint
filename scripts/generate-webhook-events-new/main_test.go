package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWriteStdoutOK(t *testing.T) {
	in := filepath.Join("testdata", "ok.html")

	b, err := os.ReadFile(in)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := parseWebhookActivityTypes(b)
	if err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	if err := write(parsed, out); err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(filepath.Join("testdata", "ok.go"))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(string(want), out.String()); diff != "" {
		t.Fatal(diff)
	}
}
