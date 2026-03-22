package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWriteStdoutOK(t *testing.T) {
	in := filepath.Join("testdata", "ok.html")
	stdout := &strings.Builder{}
	if err := run([]string{in, "-"}, stdout, io.Discard, ""); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "ok.go"))
	if err != nil {
		t.Fatal(err)
	}

	want := strings.ReplaceAll(string(b), "\r\n", "\n")
	have := strings.ReplaceAll(stdout.String(), "\r\n", "\n")
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("dummy write error")
}

func TestWriteError(t *testing.T) {
	in := filepath.Join("testdata", "ok.html")
	err := run([]string{in, "-"}, errWriter{}, io.Discard, "")
	if err == nil {
		t.Fatal("error did not occur")
	}
	if !strings.Contains(err.Error(), "could not write output") {
		t.Fatalf("unexpected error: %q", err)
	}
	if !strings.Contains(err.Error(), "dummy write error") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestInvalidCommandArgs(t *testing.T) {
	err := run([]string{"a", "b", "c"}, io.Discard, io.Discard, "")
	if err == nil {
		t.Fatal("error did not occur")
	}
	if !strings.Contains(err.Error(), "usage: generate-webhook-events [[srcfile] dstfile]") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"no_article_body.html", "article body was not found"},
		{"no_markdown_body.html", "markdown body was not found"},
		{"no_heading.html", `"About events that trigger workflows" heading was missing`},
		{"no_tables.html", "no webhook table was found"},
		{"mismatched_aria_labelledby.html", `table aria-labelledby "check_suite" did not match the expected hook id "check_run"`},
		{"no_thead.html", "thead element was missing"},
		{"no_header_row.html", "header row was missing"},
		{"too_few_headers.html", "table header had too few columns"},
		{"non_webhook_table.html", `unexpected first table header "Operator", want "Webhook event payload"`},
		{"bad_second_header.html", `unexpected second table header "Description", want "Activity types"`},
		{"no_tbody.html", "tbody element was missing"},
		{"no_table_row.html", "table row was missing"},
		{"too_few_columns.html", "table did not have at least two columns"},
		{"bad_activity_types.html", `activity types cell did not contain code elements nor 'Not applicable'`},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			stdout := &strings.Builder{}
			err := run([]string{filepath.Join("testdata", tc.file), "-"}, stdout, io.Discard, "")
			if err == nil {
				t.Fatal("error did not occur")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("wanted %q in error %q", tc.want, err)
			}
		})
	}
}
