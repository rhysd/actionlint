package actionlint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewRuleRequiredActions(t *testing.T) {
	tests := []struct {
		name     string
		required []RequiredActionRule
		want     *RuleRequiredActions
	}{
		{
			name:     "nil when no rules",
			required: []RequiredActionRule{},
			want:     nil,
		},
		{
			name: "creates rule with requirements",
			required: []RequiredActionRule{
				{Action: "actions/checkout", Version: "v3"},
			},
			want: &RuleRequiredActions{
				RuleBase: RuleBase{
					name: "required-actions",
					desc: "Checks that required GitHub Actions are used in workflows",
				},
				required: []RequiredActionRule{
					{Action: "actions/checkout", Version: "v3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRuleRequiredActions(tt.required)
			if diff := cmp.Diff(got, tt.want,
				cmpopts.IgnoreUnexported(RuleBase{}, RuleRequiredActions{})); diff != "" {
				t.Errorf("NewRuleRequiredActions mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestParseActionRef(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantVersion string
	}{
		{
			name:        "valid action reference",
			input:       "actions/checkout@v3",
			wantName:    "actions/checkout",
			wantVersion: "v3",
		},
		{
			name:        "empty string",
			input:       "",
			wantName:    "",
			wantVersion: "",
		},
		{
			name:        "docker reference",
			input:       "docker://alpine:latest",
			wantName:    "",
			wantVersion: "",
		},
		{
			name:        "no version",
			input:       "actions/checkout",
			wantName:    "",
			wantVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotVersion := parseActionRef(tt.input)
			if gotName != tt.wantName || gotVersion != tt.wantVersion {
				t.Errorf("parseActionRef(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotName, gotVersion, tt.wantName, tt.wantVersion)
			}
		})
	}
}

func TestRuleRequiredActions(t *testing.T) {
	tests := []struct {
		name        string
		required    []RequiredActionRule
		workflow    *Workflow
		wantNilRule bool
		wantErrs    int
		wantMsg     string
	}{
		{
			name:        "nil workflow",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}},
			workflow:    nil,
			wantNilRule: false,
			wantErrs:    0,
		},
		{
			name:        "empty workflow",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}},
			workflow:    &Workflow{},
			wantNilRule: false,
			wantErrs:    1,
			wantMsg:     `:1:1: required action "actions/checkout" (version "v3") is not used in this workflow [required-actions]`,
		},
		{
			name:        "NoRequiredActions",
			required:    []RequiredActionRule{},
			workflow:    &Workflow{},
			wantNilRule: true,
			wantErrs:    0,
		},
		{
			name:        "SingleRequiredAction_Present",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}},
			workflow:    &Workflow{Jobs: map[string]*Job{"build": {Steps: []*Step{{Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v3"}}}}}}},
			wantNilRule: false,
			wantErrs:    0,
		},
		{
			name:        "SingleRequiredAction_Missing_With_Version",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}},
			workflow:    &Workflow{Jobs: map[string]*Job{"build": {Steps: []*Step{{Exec: &ExecAction{Uses: &String{Value: "actions/setup-node@v2"}}}}}}},
			wantNilRule: false,
			wantErrs:    1,
			wantMsg:     `:1:1: required action "actions/checkout" (version "v3") is not used in this workflow [required-actions]`,
		},
		{
			name:        "SingleRequiredAction_Missing_Without_Version",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: ""}},
			workflow:    &Workflow{Jobs: map[string]*Job{"build": {Steps: []*Step{{Exec: &ExecAction{Uses: &String{Value: "actions/setup-node@v2"}}}}}}},
			wantNilRule: false,
			wantErrs:    1,
			wantMsg:     `:1:1: required action "actions/checkout" (version "") is not used in this workflow [required-actions]`,
		},
		{
			name:        "SingleRequiredAction_WrongVersion",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}},
			workflow:    &Workflow{Jobs: map[string]*Job{"build": {Steps: []*Step{{Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v2"}}}}}}},
			wantNilRule: false,
			wantErrs:    1,
			wantMsg:     `:1:1: action "actions/checkout" must use version "v3" but found version "v2" [required-actions]`,
		},
		{
			name:        "MultipleRequiredActions_Present",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}, {Action: "actions/setup-node", Version: "v2"}},
			workflow:    &Workflow{Jobs: map[string]*Job{"build": {Steps: []*Step{{Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v3"}}}, {Exec: &ExecAction{Uses: &String{Value: "actions/setup-node@v2"}}}}}}},
			wantNilRule: false,
			wantErrs:    0,
		},
		{
			name:        "MultipleRequiredActions_MissingOne",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}, {Action: "actions/setup-node", Version: "v2"}},
			workflow:    &Workflow{Jobs: map[string]*Job{"build": {Steps: []*Step{{Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v3"}}}}}}},
			wantNilRule: false,
			wantErrs:    1,
			wantMsg:     `:1:1: required action "actions/setup-node" (version "v2") is not used in this workflow [required-actions]`,
		},
		{
			name:        "MultipleRequiredActions_WrongVersion",
			required:    []RequiredActionRule{{Action: "actions/checkout", Version: "v3"}, {Action: "actions/setup-node", Version: "v2"}},
			workflow:    &Workflow{Jobs: map[string]*Job{"build": {Steps: []*Step{{Exec: &ExecAction{Uses: &String{Value: "actions/checkout@v2"}}}, {Exec: &ExecAction{Uses: &String{Value: "actions/setup-node@v2"}}}}}}},
			wantNilRule: false,
			wantErrs:    1,
			wantMsg:     `:1:1: action "actions/checkout" must use version "v3" but found version "v2" [required-actions]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewRuleRequiredActions(tt.required)

			if tt.wantNilRule {
				if rule != nil {
					t.Fatal("Expected nil rule")
				}
				return
			}

			if rule == nil {
				t.Fatal("Expected non-nil rule")
			}

			rule.VisitWorkflowPre(tt.workflow)
			errs := rule.Errs()
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d", len(errs), tt.wantErrs)
			}
			if tt.wantMsg != "" && len(errs) > 0 && errs[0].Error() != tt.wantMsg {
				t.Errorf("error message mismatch\ngot:  %q\nwant: %q", errs[0].Error(), tt.wantMsg)
			}
		})
	}
}
