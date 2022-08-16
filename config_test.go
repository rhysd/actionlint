package actionlint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfigParseOK(t *testing.T) {
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

func TestConfigParseError(t *testing.T) {
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

func TestConfigReadFileOK(t *testing.T) {
	p := filepath.Join("testdata", "config", "ok.yml")
	c, err := readConfigFile(p)
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
	_, err := readConfigFile(p)
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
	_, err := readConfigFile(p)
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	if !strings.Contains(msg, "could not parse config file") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestConfigGenerateDefaultConfigFileOK(t *testing.T) {
	dir, err := os.MkdirTemp(filepath.Join("testdata", "config"), "generate")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	f := filepath.Join(dir, "test.yml")
	if err := writeDefaultConfigFile(f); err != nil {
		t.Fatal(err)
	}
	if _, err := readConfigFile(f); err != nil {
		t.Fatal(err)
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
