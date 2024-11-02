package actionlint

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestConfigParseSelfHostedRunnerOK(t *testing.T) {
	testCases := []struct {
		what   string
		input  string
		labels []string
	}{
		{
			what:   "empty config",
			input:  "",
			labels: nil,
		},
		{
			what:   "empty self-hosted-runner",
			input:  "self-hosted-runner:\n",
			labels: nil,
		},
		{
			what:   "null self-hosted-runner labels",
			input:  "self-hosted-runner:\n  labels:",
			labels: nil,
		},
		{
			what:   "empty self-hosted-runner labels",
			input:  "self-hosted-runner:\n  labels: []",
			labels: []string{},
		},
		{
			what:   "self-hosted-runner labels",
			input:  "self-hosted-runner:\n  labels: [foo, bar]",
			labels: []string{"foo", "bar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			c, err := ParseConfig([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(c.SelfHostedRunner.Labels, tc.labels) {
				t.Fatal(cmp.Diff(c.SelfHostedRunner.Labels, tc.labels))
			}
		})
	}
}

func TestConfigParseError(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{
			in:   `self-hosted-runner: 42`,
			want: `cannot unmarshal`,
		},
		{
			in: `
paths:
  foo:
    ignore: foo+
`,
			want: `"ignore" must be a sequence node`,
		},
		{
			in: `
paths:
  foo:
    ignore: ['(foo']
`,
			want: `invalid regular expression "(foo" in "ignore"`,
		},
		{
			in: `
paths:
  foo.{txt,xml:
`,
			want: `invalid glob pattern`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			_, err := ParseConfig([]byte(tc.in))
			if err == nil {
				t.Fatal("no error occurred")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("wanted error message %q to contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestConfigPathConfigIgnores(t *testing.T) {
	tests := []struct {
		input string
		msg   string
		want  bool
	}{
		{
			input: ``,
			msg:   "this is test",
			want:  false,
		},
		{
			input: `ignore: []`,
			msg:   "this is test",
			want:  false,
		},
		{
			input: `ignore: ['(is )+']`,
			msg:   "this is test",
			want:  true,
		},
		{
			input: `ignore: ['does not match', '(is )+']`,
			msg:   "this is test",
			want:  true,
		},
		{
			input: `ignore: ['does not match', 'does not match 2']`,
			msg:   "this is test",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input+"_"+tc.msg, func(t *testing.T) {
			var c PathConfig
			if err := yaml.Unmarshal([]byte(tc.input), &c); err != nil {
				t.Fatal(err)
			}
			have := c.Ignore.Match(&Error{Message: tc.msg})
			if tc.want != have {
				t.Fatalf("wanted %v but got %v for message %q and input %q", tc.want, have, tc.msg, tc.input)
			}
		})
	}
}

func TestConfigIgnoreErrors(t *testing.T) {
	src := `
paths:
  .github/workflows/**/*.yaml:
    ignore: [xxx]
  .github/workflows/*.yaml:
    ignore: [yyy]
  .github/workflows/a/*.yaml:
    ignore: [zzz]
  .github/workflows/*/b.yaml:
    ignore: [uuu]
  .github/workflows/a/b.yaml:
    ignore: [vvv]
  .github/workflows/**/x.yaml:
    ignore: [www]
`

	var cfg Config
	if err := yaml.Unmarshal([]byte(src), &cfg); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path string
		msg  string
		want bool
	}{
		{"foo.yaml", "xxx", false},
		{".github/workflows/a.yaml", "xxx", true},
		{".github/workflows/a/b.yaml", "xxx", true},
		{".github/workflows/a/b/c/d/e/f/g/h.yaml", "xxx", true},
		{".github/workflows/a.yaml", "yyy", true},
		{".github/workflows/a/b.yaml", "yyy", false},
		{".github/workflows/a/b.yaml", "zzz", true},
		{".github/workflows/a/a.yaml", "zzz", true},
		{".github/workflows/b/b.yaml", "zzz", false},
		{".github/workflows/a/b.yaml", "uuu", true},
		{".github/workflows/b/b.yaml", "uuu", true},
		{".github/workflows/a/a.yaml", "uuu", false},
		{".github/workflows/a/b.yaml", "vvv", true},
		{".github/workflows/b/b.yaml", "vvv", false},
		{".github/workflows/a/a.yaml", "vvv", false},
		{".github/workflows/x.yaml", "www", true},
		{".github/workflows/a/x.yaml", "www", true},
		{".github/workflows/a/b/x.yaml", "www", true},
		{".github/workflows/a/b/c/x.yaml", "www", true},
		{".github/workflows/a/b.yaml", "this is not ignored", false},
	}

	for _, tc := range tests {
		var ignored bool
		for _, c := range cfg.PathConfigs(tc.path) {
			if c.Ignore.Match(&Error{Message: tc.msg}) {
				ignored = true
				break
			}
		}
		if ignored != tc.want {
			want, have := "not be ignored", "was ignored"
			if tc.want {
				want, have = "be ignored", "was not ignored"
			}
			t.Fatalf("error message %q with path %q should %s but actually %s", tc.msg, tc.path, want, have)
		}
	}
}

func TestConfigReadFileOK(t *testing.T) {
	p := filepath.Join("testdata", "config", "ok.yml")
	c, err := ReadConfigFile(p)
	if err != nil {
		t.Fatal(err)
	}
	labels := []string{"foo", "bar"}
	if !cmp.Equal(c.SelfHostedRunner.Labels, labels) {
		t.Fatal(cmp.Diff(c.SelfHostedRunner.Labels, labels))
	}
}

func TestConfigReadFileReadError(t *testing.T) {
	p := filepath.Join("testdata", "config", "does-not-exist.yml")
	_, err := ReadConfigFile(p)
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	if !strings.Contains(msg, "could not read config file") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestConfigReadFileParseError(t *testing.T) {
	p := filepath.Join("testdata", "config", "broken.yml")
	_, err := ReadConfigFile(p)
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	if !strings.Contains(msg, "could not parse config file") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestConfigGenerateDefaultConfigFileOK(t *testing.T) {
	f := filepath.Join(t.TempDir(), "default-config-for-test.yml")
	if err := writeDefaultConfigFile(f); err != nil {
		t.Fatal(err)
	}
	c, err := ReadConfigFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.SelfHostedRunner.Labels) != 0 {
		t.Fatal(c.SelfHostedRunner.Labels)
	}
	if c.ConfigVariables != nil {
		t.Fatal(c.SelfHostedRunner.Labels)
	}
	if len(c.Paths) != 0 {
		t.Fatal(c.Paths)
	}
}

func TestConfigGenerateDefaultConfigFileError(t *testing.T) {
	p := filepath.Join("testdata", "config", "dir-does-not-exist", "test.yml")
	err := writeDefaultConfigFile(p)
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	if !strings.Contains(msg, "could not write default configuration file") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}
