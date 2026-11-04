package actionlint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuleActionCheckLocalCompositeActionPins(t *testing.T) {
	const pinned = "0123456789abcdef0123456789abcdef01234567"

	tests := []struct {
		name    string
		file    string
		uses    string
		wantErr bool
	}{
		{name: "tag is rejected", file: "action.yml", uses: "actions/checkout@v4", wantErr: true},
		{name: "short SHA is rejected", file: "action.yaml", uses: "actions/checkout@1234", wantErr: true},
		{name: "full SHA is accepted", file: "action.yml", uses: "actions/checkout@" + pinned},
		{name: "local action is skipped", file: "action.yml", uses: "./nested"},
		{name: "docker action is skipped", file: "action.yml", uses: "docker://alpine:3.20"},
		{name: "expression is skipped", file: "action.yml", uses: "${{ inputs.action }}"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			metadata := "name: Test\ndescription: Test composite action\nruns:\n  using: composite\n  steps:\n    - uses: " + tc.uses + "\n"
			if err := os.WriteFile(filepath.Join(dir, tc.file), []byte(metadata), 0600); err != nil {
				t.Fatal(err)
			}

			rule := NewRuleAction(nil)
			rule.SetConfig(&Config{RequireCommitHashInComposites: true})
			rule.checkLocalCompositeActionPins(dir, &Pos{Line: 99, Col: 1})

			errs := rule.Errs()
			if tc.wantErr {
				if len(errs) != 1 {
					t.Fatalf("wanted one pinning error, got %d: %v", len(errs), errs)
				}
				if !strings.Contains(errs[0].Message, "must be pinned to a full-length commit SHA") {
					t.Fatalf("unexpected error: %q", errs[0].Message)
				}
				if errs[0].Line != 6 {
					t.Errorf("error should point to the nested uses line 6, got line %d", errs[0].Line)
				}
			} else if len(errs) != 0 {
				t.Fatalf("did not expect errors, got %v", errs)
			}
		})
	}
}

func TestRuleActionCheckLocalCompositeActionPinsDisabledByDefault(t *testing.T) {
	dir := t.TempDir()
	metadata := "name: Test\ndescription: Test composite action\nruns:\n  using: composite\n  steps:\n    - uses: actions/checkout@v4\n"
	if err := os.WriteFile(filepath.Join(dir, "action.yml"), []byte(metadata), 0600); err != nil {
		t.Fatal(err)
	}

	rule := NewRuleAction(nil)
	rule.SetConfig(&Config{})
	rule.checkLocalCompositeActionPins(dir, &Pos{Line: 99, Col: 1})
	if errs := rule.Errs(); len(errs) != 0 {
		t.Fatalf("pinning check should be disabled by default, got %v", errs)
	}
}
