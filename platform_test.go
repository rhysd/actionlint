package actionlint

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsGitHubHostedPlatform(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"", true},
		{"https://github.com", true},
		{"https://github.com/", true},
		{"https://my-org.ghe.com", true},
		{"https://my-org.ghe.com/", true},
		{"https://github.mycompany.com", false},
		{"https://ghes.internal.corp", false},
	}

	for _, tc := range tests {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv("GITHUB_SERVER_URL", tc.env)
			if got := isGitHubHostedPlatform(); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestOutdatedActionsSkippedOnGHES(t *testing.T) {
	t.Setenv("GITHUB_SERVER_URL", "https://github.example.com")

	linter, err := NewLinter(io.Discard, &LinterOptions{})
	if err != nil {
		t.Fatal(err)
	}

	f := filepath.Join("testdata", "err", "outdated_actions.yaml")
	errs, err := linter.LintFile(f, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range errs {
		if strings.Contains(e.Message, "too old to run on GitHub Actions") {
			t.Fatalf("unexpected outdated action error on GHES: %s", e.Message)
		}
	}
}

func TestGHESArtifactV3LintOK(t *testing.T) {
	t.Setenv("GITHUB_SERVER_URL", "https://github.example.com")

	linter, err := NewLinter(io.Discard, &LinterOptions{})
	if err != nil {
		t.Fatal(err)
	}

	f := filepath.Join("testdata", "ok", "ghes", "artifact-v3.yaml")
	errs, err := linter.LintFile(f, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestOutdatedActionsReportedOnGitHubCom(t *testing.T) {
	t.Setenv("GITHUB_SERVER_URL", "https://github.com")

	linter, err := NewLinter(io.Discard, &LinterOptions{})
	if err != nil {
		t.Fatal(err)
	}

	f := filepath.Join("testdata", "err", "outdated_actions.yaml")
	errs, err := linter.LintFile(f, nil)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "too old to run on GitHub Actions") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected outdated action errors on github.com")
	}
}
