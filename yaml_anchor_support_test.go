package actionlint

import (
	"io"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLAnchorResolution(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErrs []string
	}{
		{
			name: "Basic anchor and alias for env",
			input: `
name: Test YAML Anchors
on: push
env: &default_env
  NODE_VERSION: "18"
  PYTHON_VERSION: "3.9"
jobs:
  test:
    runs-on: ubuntu-latest
    env: *default_env
    steps:
      - uses: actions/checkout@v4
      - run: echo "Node version ${{ env.NODE_VERSION }}"
`,
			wantErrs: nil,
		},
		{
			name: "Anchor and alias for steps",
			input: `
name: Test Steps Anchor
on: push
jobs:
  test_steps:
    runs-on: ubuntu-latest
    steps: &common_steps
      - uses: actions/checkout@v4
      - run: npm install
      - run: npm test

  another_job:
    runs-on: ubuntu-latest
    steps: *common_steps
`,
			wantErrs: nil,
		},
		{
			name: "Anchor and alias for entire job",
			input: `
name: Test Job Anchor
on: push
jobs:
  build: &build_job
    runs-on: ubuntu-latest
    env:
      NODE_VERSION: "18"
    steps:
      - uses: actions/checkout@v4
      - run: npm install

  test: *build_job
`,
			wantErrs: nil,
		},
		{
			name: "Anchor for individual step",
			input: `
name: Test Step Anchor
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - &common_step
        name: "Common step"
        run: echo "Hello"
      - *common_step
      - name: "Different step"
        run: echo "World"
`,
			wantErrs: nil,
		},
		{
			name: "Anchors with linting errors should still be caught",
			input: `
name: Test Anchors with Errors
on: push
env: &default_env
  NODE_VERSION: "18"
jobs:
  test:
    runs-on: ubuntu-latest
    env: *default_env
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node_version: ${{ env.NODE_VERSION }}
`,
			wantErrs: []string{
				"input \"node_version\" is not defined in action \"actions/setup-node@v4\"",
			},
		},
		{
			name: "Undefined anchor should error",
			input: `
name: Test Undefined Anchor
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    env: *undefined_anchor
    steps:
      - uses: actions/checkout@v4
`,
			wantErrs: []string{
				"could not parse as YAML: yaml: unknown anchor 'undefined_anchor' referenced",
			},
		},
		{
			name: "Complex nested anchors",
			input: `
name: Complex Anchors
on: push
env: &env_vars
  NODE_VERSION: "18"
  PYTHON_VERSION: "3.9"
defaults: &default_settings
  run:
    shell: bash
jobs:
  test: &test_job
    runs-on: ubuntu-latest
    env: *env_vars
    defaults: *default_settings
    steps:
      - uses: actions/checkout@v4
      - run: echo "Testing with Node ${{ env.NODE_VERSION }}"

  build: *test_job

  deploy:
    runs-on: ubuntu-latest
    env: *env_vars
    defaults: *default_settings
    steps:
      - uses: actions/checkout@v4
      - run: echo "Deploying"
`,
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For tests that expect linting errors (not just parse errors), use the full linter
			if tt.wantErrs != nil && !isParseError(tt.wantErrs) {
				// Use linter for full validation
				l, err := NewLinter(io.Discard, &LinterOptions{})
				if err != nil {
					t.Fatalf("Failed to create linter: %v", err)
				}

				errs, err := l.Lint("test.yaml", []byte(tt.input), nil)
				if err != nil {
					t.Fatalf("Linter failed: %v", err)
				}

				if len(errs) == 0 {
					t.Errorf("Expected errors %v, but got none", tt.wantErrs)
					return
				}

				var gotErrMsgs []string
				for _, err := range errs {
					gotErrMsgs = append(gotErrMsgs, err.Message)
				}

				for _, wantErr := range tt.wantErrs {
					found := false
					for _, gotErr := range gotErrMsgs {
						if strings.Contains(gotErr, wantErr) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing %q, but got errors: %v", wantErr, gotErrMsgs)
					}
				}
			} else {
				// Use parser for syntax/parse errors
				workflow, errs := Parse([]byte(tt.input))

				if tt.wantErrs == nil {
					if len(errs) > 0 {
						t.Errorf("Expected no errors, but got: %v", errs)
					}
					if workflow == nil {
						t.Error("Expected workflow to be parsed, but got nil")
					}
				} else {
					if len(errs) == 0 {
						t.Errorf("Expected errors %v, but got none", tt.wantErrs)
						return
					}

					var gotErrMsgs []string
					for _, err := range errs {
						gotErrMsgs = append(gotErrMsgs, err.Message)
					}

					for _, wantErr := range tt.wantErrs {
						found := false
						for _, gotErr := range gotErrMsgs {
							if strings.Contains(gotErr, wantErr) {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Expected error containing %q, but got errors: %v", wantErr, gotErrMsgs)
						}
					}
				}
			}
		})
	}
}

func TestYAMLAnchorResolutionPreservesContent(t *testing.T) {
	input := `
name: Test Content Preservation
on: push
env: &default_env
  NODE_VERSION: "18"
  PYTHON_VERSION: "3.9"
jobs:
  test:
    runs-on: ubuntu-latest
    env: *default_env
    steps:
      - uses: actions/checkout@v4
      - name: Check env vars
        run: |
          echo "Node: ${{ env.NODE_VERSION }}"
          echo "Python: ${{ env.PYTHON_VERSION }}"
`

	workflow, errs := Parse([]byte(input))
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if workflow == nil {
		t.Fatal("Expected workflow to be parsed")
	}

	// Verify that the content was properly resolved
	if workflow.Name == nil || workflow.Name.Value != "Test Content Preservation" {
		t.Errorf("Expected workflow name to be preserved")
	}

	if workflow.Jobs == nil || len(workflow.Jobs) != 1 {
		t.Errorf("Expected exactly one job")
	}

	var testJob *Job
	for _, job := range workflow.Jobs {
		testJob = job
		break
	}
	if testJob.ID == nil || testJob.ID.Value != "test" {
		t.Errorf("Expected job ID to be 'test'")
	}

	// Check that env vars were properly resolved
	if testJob.Env == nil || len(testJob.Env.Vars) != 2 {
		t.Errorf("Expected 2 environment variables, got %d", len(testJob.Env.Vars))
	}

	envVars := make(map[string]string)
	for _, env := range testJob.Env.Vars {
		envVars[env.Name.Value] = env.Value.Value
	}

	if envVars["NODE_VERSION"] != "18" {
		t.Errorf("Expected NODE_VERSION to be '18', got %q", envVars["NODE_VERSION"])
	}

	if envVars["PYTHON_VERSION"] != "3.9" {
		t.Errorf("Expected PYTHON_VERSION to be '3.9', got %q", envVars["PYTHON_VERSION"])
	}
}

func TestYAMLWithoutAnchors(t *testing.T) {
	input := `
name: No Anchors
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "Hello World"
`

	workflow, errs := Parse([]byte(input))
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if workflow == nil {
		t.Fatal("Expected workflow to be parsed")
	}

	// Should work exactly as before for workflows without anchors
	if workflow.Name == nil || workflow.Name.Value != "No Anchors" {
		t.Errorf("Expected workflow name to be 'No Anchors'")
	}
}

// Helper function to determine if errors are parse errors vs linting errors
func isParseError(errs []string) bool {
	for _, err := range errs {
		if strings.Contains(err, "could not parse as YAML") || strings.Contains(err, "unknown anchor") {
			return true
		}
	}
	return false
}

func TestContainsAnchorsOrAliases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name: "No anchors or aliases",
			input: `
name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			want: false,
		},
		{
			name: "Has anchor",
			input: `
name: Test
env: &default_env
  VAR: value
`,
			want: true,
		},
		{
			name: "Has alias",
			input: `
name: Test
env: &default_env
  VAR: value
other: *default_env
`,
			want: true,
		},
		{
			name: "Has both anchor and alias",
			input: `
name: Test
env: &default_env
  VAR: value
jobs:
  test:
    env: *default_env
`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.input), &node); err != nil {
				t.Fatalf("Failed to unmarshal YAML: %v", err)
			}

			got := containsAnchorsOrAliases(&node)
			if got != tt.want {
				t.Errorf("containsAnchorsOrAliases() = %v, want %v", got, tt.want)
			}
		})
	}
}
