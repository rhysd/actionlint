package main

import (
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

	want, err := os.ReadFile(filepath.Join("testdata", "ok.go"))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(string(want), stdout.String()); diff != "" {
		t.Fatal(diff)
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
		{"no_thead.html", "no webhook table was found"},
		{"no_header_row.html", "no webhook table was found"},
		{"too_few_headers.html", "no webhook table was found"},
		{"non_webhook_table.html", "no webhook table was found"},
		{"bad_second_header.html", "no webhook table was found"},
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
