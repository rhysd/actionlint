package actionlint

import (
	"path/filepath"
	"sort"
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
			c, err := parseConfig([]byte(tc.input), "/path/to/file.yml")
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(c.SelfHostedRunner.Labels, tc.labels) {
				t.Fatal(cmp.Diff(c.SelfHostedRunner.Labels, tc.labels))
			}
		})
	}
}

func TestConfigParseSelfHostedRunnerError(t *testing.T) {
	input := "self-hosted-runner: 42\n"
	_, err := parseConfig([]byte(input), "/path/to/file.yml")
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	if !strings.Contains(msg, "could not parse config file \"/path/to/file.yml\"") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestConfigParsePathConfigsOK(t *testing.T) {
	tests := []struct {
		what  string
		input string
		want  map[string][]string
	}{
		{
			what:  "empty",
			input: "",
			want:  map[string][]string{},
		},
		{
			what:  "no config",
			input: ".github/**/*.yaml:\n",
			want: map[string][]string{
				".github/**/*.yaml": {},
			},
		},
		{
			what: "empty ignore",
			input: `
.github/**/*.yaml:
  ignore: []
`,
			want: map[string][]string{
				".github/**/*.yaml": {},
			},
		},
		{
			what: "ignore",
			input: `
.github/**/*.yaml:
  ignore:
    - ^foo.+bar$
    - aaa\s+bbb
`,
			want: map[string][]string{
				".github/**/*.yaml": {"^foo.+bar$", `aaa\s+bbb`},
			},
		},
		{
			what: "multiple cofigs",
			input: `
.github/**/*.yaml:
  ignore: [aaa]
.github/**/foo.yaml:
  ignore:
    - bbb
    - ccc
.github/**/bar.yaml:
`,
			want: map[string][]string{
				".github/**/*.yaml":   {"aaa"},
				".github/**/foo.yaml": {"bbb", "ccc"},
				".github/**/bar.yaml": {},
			},
		},
		{
			what: "non-string regex",
			input: `
.github/**/*.yaml:
  ignore: [true]
`,
			want: map[string][]string{
				".github/**/*.yaml": {"true"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			var c PathConfigs
			if err := yaml.Unmarshal([]byte(tc.input), &c); err != nil {
				t.Fatal(err)
			}
			have := map[string][]string{}
			for glob, cfg := range c {
				ignore := []string{}
				for _, i := range cfg.Ignore {
					ignore = append(ignore, i.String())
				}
				have[glob] = ignore
			}
			if !cmp.Equal(tc.want, have) {
				t.Fatal(cmp.Diff(tc.want, have))
			}
		})
	}
}

func TestConfigParsePathConfigsError(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"42\n", "paths must be mapping node"},
		{"foo:\nfoo:\n", "key duplicates"},
		{"foo:\n  unknown: 42\n", `invalid key "unknown"`},
		{"foo:\n  ignore: 42\n", `"ignore" must be a sequence node`},
		{"foo:\n  ignore: ['(foo']\n", "invalid regular expression"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			var c PathConfigs
			err := yaml.Unmarshal([]byte(tc.input), &c)
			if err == nil {
				t.Fatal("error did not occur")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("error message %q doesn't contain expected message %q", msg, tc.want)
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
			have := c.Ignores(&Error{Message: tc.msg})
			if tc.want != have {
				t.Fatalf("wanted %v but got %v for message %q and input %q", tc.want, have, tc.msg, tc.input)
			}
		})
	}
}

func TestConfigGetPathConfigs(t *testing.T) {
	tests := []struct {
		what  string
		input string
		path  string
		want  []string
	}{
		{
			what:  "empty",
			input: "",
			path:  ".github/workflows/foo.yaml",
			want:  []string{},
		},
		{
			what:  "single",
			input: ".github/workflows/*.yaml:",
			path:  ".github/workflows/foo.yaml",
			want:  []string{".github/workflows/*.yaml"},
		},
		{
			what: "multiple",
			input: `
.github/workflows/**/*.yaml:
.github/workflows/foo/bar.yaml:
.github/workflows/foo/*.yaml:
.github/workflows/foo/piyo.yaml:
.github/workflows/*/bar.yaml:
.github/workflows/piyo/bar.yaml:
`,
			path: ".github/workflows/foo/bar.yaml",
			want: []string{
				".github/workflows/**/*.yaml",
				".github/workflows/foo/bar.yaml",
				".github/workflows/foo/*.yaml",
				".github/workflows/*/bar.yaml",
			},
		},
		{
			what: "not found",
			input: `
.github/workflows/foo/bar.yaml:
.github/workflows/foo/*.yaml:
.github/workflows/foo/piyo.yaml:
.github/workflows/*/bar.yaml:
.github/workflows/piyo/bar.yaml:
`,
			path: ".github/workflows/woo/boo.yaml",
			want: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			var pc PathConfigs
			if err := yaml.Unmarshal([]byte(tc.input), &pc); err != nil {
				t.Fatal(err)
			}

			cfg := &Config{Paths: pc}
			have := []string{}
			for _, c := range cfg.PathConfigsFor(tc.path) {
				for k, v := range pc {
					if c == v {
						have = append(have, k)
						break
					}
				}
			}
			sort.Strings(have)
			sort.Strings(tc.want)

			if !cmp.Equal(tc.want, have) {
				t.Fatal(cmp.Diff(tc.want, have))
			}
		})
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
