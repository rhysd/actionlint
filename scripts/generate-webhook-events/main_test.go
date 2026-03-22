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
		{"mismatched_aria_labelledby.html", `could not parse webhook types table for "check_run": expected table[aria-labelledby] to be "check_run", got "check_suite"`},
		{"no_thead.html", `could not parse webhook types table for "check_run": missing thead element`},
		{"no_header_row.html", `could not parse webhook types table for "check_run": missing header row in thead`},
		{"too_few_headers.html", `could not parse webhook types table for "check_run": expected at least 2 header columns, got 1`},
		{"non_webhook_table.html", `could not parse webhook types table for "check_run": expected first header to be "Webhook event payload", got "Operator"`},
		{"bad_second_header.html", `could not parse webhook types table for "check_run": expected second header to be "Activity types", got "Description"`},
		{"no_tbody.html", `could not parse webhook types table for "check_run": missing tbody element`},
		{"no_table_row.html", `could not parse webhook types table for "check_run": missing first data row in tbody`},
		{"too_few_columns.html", `could not parse webhook types table for "check_run": expected at least 2 data columns, got 1`},
		{"bad_activity_types.html", `could not parse webhook types table for "check_run": unexpected activity types cell text "Surprise"; expected code elements, "Not applicable", or "Custom"`},
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
