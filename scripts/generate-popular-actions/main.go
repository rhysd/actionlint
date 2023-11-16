package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/rhysd/actionlint"
	"gopkg.in/yaml.v3"
)

type yamlExt int

const (
	yamlExtYML yamlExt = iota
	yamlExtYAML
)

func (ext yamlExt) String() string {
	if ext == yamlExtYML {
		return "yml"
	}
	return "yaml"
}

type actionJSON struct {
	Spec string                     `json:"spec"`
	Meta *actionlint.ActionMetadata `json:"metadata"`
}

type action struct {
	slug string
	path string
	tags []string
	next string
	ext  yamlExt
}

func (a *action) rawURL(tag string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s%s/action.%s", a.slug, tag, a.path, a.ext.String())
}

func (a *action) githubURL(tag string) string {
	return fmt.Sprintf("https://github.com/%s/tree/%s%s", a.slug, tag, a.path)
}

func (a *action) spec(tag string) string {
	return fmt.Sprintf("%s%s@%s", a.slug, a.path, tag)
}

type slugSet = map[string]struct{}

// Note: Actions used by top 1000 public repositories at GitHub sorted by number of occurrences:
// https://gist.github.com/rhysd/1db81fa80096b699b9c045f435d0cace

var popularActions = []*action{
	{
		slug: "8398a7/action-slack",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "Azure/container-scan",
		tags: []string{"v0"},
		next: "v1",
		ext:  yamlExtYAML,
	},
	{
		slug: "Azure/functions-action",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "EnricoMi/publish-unit-test-result-action",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "JamesIves/github-pages-deploy-action",
		tags: []string{"releases/v3", "releases/v4"},
		next: "release/v5",
	},
	{
		slug: "ReactiveCircus/android-emulator-runner",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "Swatinem/rust-cache",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "actions-cool/issues-helper",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "actions-rs/audit-check",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "actions-rs/cargo",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "dtolnay/rust-toolchain",
		tags: []string{"stable", "beta", "nightly"},
		next: "",
	},
	{
		slug: "actions-rs/clippy-check",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "actions-rs/toolchain",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "actions/cache",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "actions/checkout",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "actions/configure-pages",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "actions/deploy-pages",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "actions/delete-package-versions",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "actions/download-artifact",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "actions/first-interaction",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "actions/github-script",
		tags: []string{"v1", "v2", "v3", "v4", "v5", "v6"},
		next: "v7",
	},
	{
		slug: "actions/labeler",
		tags: []string{"v2", "v3", "v4"},
		next: "v5"}, // v1 does not exist
	{
		slug: "actions/setup-dotnet",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "actions/setup-go",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "actions/setup-java",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "actions/setup-node",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "actions/setup-python",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "actions/stale",
		tags: []string{"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8"},
		next: "v9",
	},
	{
		slug: "actions/upload-artifact",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "actions/upload-pages-artifact",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "actions/dependency-review-action",
		tags: []string{"v3"},
		next: "v4",
	},
	{
		slug: "aws-actions/configure-aws-credentials",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "azure/aks-set-context",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "azure/login",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "bahmutov/npm-install",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "codecov/codecov-action",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "dawidd6/action-download-artifact",
		tags: []string{"v2"},
		next: "v3",
	},
	{
		slug: "dawidd6/action-send-mail",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "dessant/lock-threads",
		tags: []string{"v2", "v3", "v4"},
		next: "v5"}, // v1 does not exist
	{
		slug: "docker/build-push-action",
		tags: []string{"v1", "v2", "v3", "v4", "v5"},
		next: "v6",
	},
	{
		slug: "docker/login-action",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "docker/metadata-action",
		tags: []string{"v1", "v2", "v3", "v4", "v5"},
		next: "v6",
	},
	{
		slug: "docker/setup-buildx-action",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "docker/setup-qemu-action",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "dorny/paths-filter",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "enriikke/gatsby-gh-pages-action",
		tags: []string{"v2"},
		next: "v3",
	},
	{
		slug: "erlef/setup-beam",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "game-ci/unity-builder",
		tags: []string{"v2", "v3"},
		next: "v4",
	},
	{
		slug: "getsentry/paths-filter",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "github/codeql-action",
		path: "/analyze",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "github/codeql-action",
		path: "/autobuild",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "github/codeql-action",
		path: "/init",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "github/super-linter",
		tags: []string{"v3", "v4", "v5"},
		next: "v6",
	},
	{
		slug: "githubocto/flat",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "golangci/golangci-lint-action",
		tags: []string{"v1", "v2", "v3"},
		next: "v4",
	},
	{
		slug: "google-github-actions/auth",
		tags: []string{"v0", "v1"},
		next: "v2",
	},
	{
		slug: "google-github-actions/get-secretmanager-secrets",
		tags: []string{"v0", "v1"},
		next: "v2",
	},
	{
		slug: "google-github-actions/setup-gcloud",
		tags: []string{"v0", "v1"},
		next: "v2",
	},
	{
		slug: "google-github-actions/upload-cloud-storage",
		tags: []string{"v0", "v1"},
		next: "v2",
	},
	{
		slug: "goreleaser/goreleaser-action",
		tags: []string{"v1", "v2", "v3", "v4", "v5"},
		next: "v6",
	},
	{
		slug: "gradle/wrapper-validation-action",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "haskell/actions",
		path: "/setup",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "marvinpinto/action-automatic-releases",
		tags: []string{"latest"},
		next: "",
	},
	{
		slug: "microsoft/playwright-github-action",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "mikepenz/release-changelog-builder-action",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "msys2/setup-msys2",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "ncipollo/release-action",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "nwtgck/actions-netlify",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "octokit/request-action",
		tags: []string{"v1.x", "v2.x"},
		next: "v3.x",
	},
	{
		slug: "peaceiris/actions-gh-pages",
		tags: []string{"v2", "v3"},
		next: "v4",
	},
	{
		slug: "peter-evans/create-pull-request",
		tags: []string{"v1", "v2", "v3", "v4", "v5"},
		next: "v6",
	},
	{
		slug: "preactjs/compressed-size-action",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "pulumi/actions",
		tags: []string{"v1", "v2", "v3", "v4"},
		next: "v5",
	},
	{
		slug: "pypa/gh-action-pypi-publish",
		tags: []string{"release/v1"},
		next: "release/v2",
	},
	{
		slug: "reviewdog/action-actionlint",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "reviewdog/action-eslint",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "reviewdog/action-golangci-lint",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "reviewdog/action-hadolint",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "reviewdog/action-misspell",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "reviewdog/action-rubocop",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "reviewdog/action-shellcheck",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "reviewdog/action-tflint",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "rhysd/action-setup-vim",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "ridedott/merge-me-action",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "rtCamp/action-slack-notify",
		tags: []string{"v2"},
		next: "v3",
	},
	{
		slug: "ruby/setup-ruby",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "shivammathur/setup-php",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "softprops/action-gh-release",
		tags: []string{"v1"},
		next: "v2",
	},
	{
		slug: "subosito/flutter-action",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
	{
		slug: "treosh/lighthouse-ci-action",
		tags: []string{"v1", "v2", "v3", "v7", "v8", "v9", "v10"},
		next: "v11",
	},
	{
		slug: "wearerequired/lint-action",
		tags: []string{"v1", "v2"},
		next: "v3",
	},
}

// slugs not to check inputs. Some actions allow to specify inputs which are not defined in action.yml.
// In such cases, actionlint no longer can check the inputs, but it can still check outputs. (#16)
var doNotCheckInputs = slugSet{
	"octokit/request-action": {},
}

// slugs which allows any outputs to be set. Some actions sets outputs 'dynamically'. Those outputs
// may or may not exist. And they are not listed in action.yml metadata. actionlint cannot check
// such outputs and fallback into allowing to set any outputs. (#18)
var doNotCheckOutputs = slugSet{
	"dorny/paths-filter":     {},
	"getsentry/paths-filter": {},
}

type app struct {
	stdout      io.Writer
	stderr      io.Writer
	log         *log.Logger
	actions     []*action
	skipInputs  slugSet
	skipOutputs slugSet
}

func newApp(stdout, stderr, dbgout io.Writer, actions []*action, skipInputs, skipOutputs slugSet) *app {
	l := log.New(dbgout, "", log.LstdFlags)
	return &app{stdout, stderr, l, actions, skipInputs, skipOutputs}
}

func (a *app) fetchRemote() (map[string]*actionlint.ActionMetadata, error) {
	type request struct {
		action *action
		tag    string
	}

	type fetched struct {
		spec string
		meta *actionlint.ActionMetadata
		err  error
	}

	results := make(chan *fetched)
	reqs := make(chan *request)
	done := make(chan struct{})

	for i := 0; i <= 4; i++ {
		go func(ret chan<- *fetched, reqs <-chan *request, done <-chan struct{}) {
			var c http.Client
			for {
				select {
				case req := <-reqs:
					url := req.action.rawURL(req.tag)
					a.log.Println("Start fetching", url)
					res, err := c.Get(url)
					if err != nil {
						ret <- &fetched{err: fmt.Errorf("could not fetch %s: %w", url, err)}
						break
					}
					if res.StatusCode < 200 || 300 <= res.StatusCode {
						ret <- &fetched{err: fmt.Errorf("request was not successful %s: %s", url, res.Status)}
						break
					}
					body, err := io.ReadAll(res.Body)
					res.Body.Close()
					if err != nil {
						ret <- &fetched{err: fmt.Errorf("could not read body for %s: %w", url, err)}
						break
					}
					spec := req.action.spec(req.tag)
					var meta actionlint.ActionMetadata
					if err := yaml.Unmarshal(body, &meta); err != nil {
						ret <- &fetched{err: fmt.Errorf("could not parse metadata for %s: %w", url, err)}
						break
					}
					if _, ok := a.skipInputs[req.action.slug]; ok {
						meta.SkipInputs = true
					}
					if _, ok := a.skipOutputs[req.action.slug]; ok {
						meta.SkipOutputs = true
					}
					ret <- &fetched{spec: spec, meta: &meta}
				case <-done:
					return
				}
			}
		}(results, reqs, done)
	}

	n := 0
	for _, action := range a.actions {
		n += len(action.tags)
	}

	go func(reqs chan<- *request, done <-chan struct{}) {
		for _, action := range a.actions {
			for _, tag := range action.tags {
				select {
				case reqs <- &request{action, tag}:
				case <-done:
					return
				}
			}
		}
	}(reqs, done)

	ret := make(map[string]*actionlint.ActionMetadata, n)
	for i := 0; i < n; i++ {
		f := <-results
		if f.err != nil {
			close(done)
			return nil, f.err
		}
		ret[f.spec] = f.meta
	}

	close(done)
	return ret, nil
}

func (a *app) writeJSONL(out io.Writer, actions map[string]*actionlint.ActionMetadata) error {
	enc := json.NewEncoder(out)
	for spec, meta := range actions {
		j := actionJSON{spec, meta}
		if err := enc.Encode(&j); err != nil {
			return fmt.Errorf("could not encode action %q data into JSON: %w", spec, err)
		}
	}
	a.log.Printf("Wrote %d action metadata as JSONL", len(actions))
	return nil
}

func (a *app) writeGo(out io.Writer, actions map[string]*actionlint.ActionMetadata) error {
	b := &bytes.Buffer{}
	fmt.Fprint(b, `// Code generated by actionlint/scripts/generate-popular-actions. DO NOT EDIT.

package actionlint

// PopularActions is data set of known popular actions. Keys are specs (owner/repo@ref) of actions
// and values are their metadata.
var PopularActions = map[string]*ActionMetadata{
`)

	specs := make([]string, 0, len(actions))
	for s := range actions {
		specs = append(specs, s)
	}
	sort.Strings(specs)

	for _, spec := range specs {
		meta := actions[spec]
		fmt.Fprintf(b, "%q: {\n", spec)
		fmt.Fprintf(b, "Name: %q,\n", meta.Name)

		slug := spec[:strings.IndexRune(spec, '@')]
		_, skipInputs := a.skipInputs[slug]
		if skipInputs {
			fmt.Fprintf(b, "SkipInputs: true,\n")
		}

		if len(meta.Inputs) > 0 && !skipInputs {
			ids := make([]string, 0, len(meta.Inputs))
			for n := range meta.Inputs {
				ids = append(ids, n)
			}
			sort.Strings(ids)

			fmt.Fprintf(b, "Inputs: ActionMetadataInputs{\n")
			for _, id := range ids {
				i := meta.Inputs[id]
				fmt.Fprintf(b, "%q: {%q, %v},\n", id, i.Name, i.Required)
			}
			fmt.Fprintf(b, "},\n")
		}

		_, skipOutputs := a.skipOutputs[slug]
		if skipOutputs {
			fmt.Fprintf(b, "SkipOutputs: true,\n")
		}

		if len(meta.Outputs) > 0 && !skipOutputs {
			ids := make([]string, 0, len(meta.Outputs))
			for n := range meta.Outputs {
				ids = append(ids, n)
			}
			sort.Strings(ids)

			fmt.Fprintf(b, "Outputs: ActionMetadataOutputs{\n")
			for _, id := range ids {
				o := meta.Outputs[id]
				fmt.Fprintf(b, "%q: {%q},\n", id, o.Name)
			}
			fmt.Fprintf(b, "},\n")
		}

		fmt.Fprintf(b, "},\n")
	}

	fmt.Fprintln(b, "}")

	// Format the generated source with checking Go syntax
	gen := b.Bytes()
	src, err := format.Source(gen)
	if err != nil {
		return fmt.Errorf("could not format generated Go code: %w\n%s", err, gen)
	}

	if _, err := out.Write(src); err != nil {
		return fmt.Errorf("could not output generated Go source to stdout: %w", err)
	}

	a.log.Printf("Wrote %d action metadata as Go", len(actions))
	return nil
}

func (a *app) readJSONL(file string) (map[string]*actionlint.ActionMetadata, error) {
	if !strings.HasSuffix(file, ".jsonl") {
		return nil, fmt.Errorf("JSONL file name must end with \".jsonl\": %s", file)
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", file, err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	ret := map[string]*actionlint.ActionMetadata{}
	for {
		l, err := r.ReadBytes('\n')
		if err == io.EOF {
			a.log.Printf("Read %d action metadata from %s", len(ret), file)
			return ret, nil
		} else if err != nil {
			return nil, fmt.Errorf("could not read line in file %s: %w", file, err)
		}
		var j actionJSON
		if err := json.Unmarshal(l, &j); err != nil {
			return nil, fmt.Errorf("could not parse line as JSON for action metadata in file %s: %s", file, err)
		}
		ret[j.Spec] = j.Meta
	}
}

func (a *app) detectNewReleaseURLs() ([]string, error) {
	urls := make(chan string)
	done := make(chan struct{})
	errs := make(chan error)
	reqs := make(chan *action)

	for i := 0; i < 4; i++ {
		go func(ret chan<- string, errs chan<- error, reqs <-chan *action, done <-chan struct{}) {
			var c http.Client
			for {
				select {
				case r := <-reqs:
					if r.next == "" {
						ret <- ""
						break
					}
					url := r.rawURL(r.next)
					a.log.Println("Checking", url)
					res, err := c.Head(url)
					if err != nil {
						errs <- fmt.Errorf("could not send head request to %s: %w", url, err)
						break
					}
					if res.StatusCode == 404 {
						a.log.Println("Not found:", url)
						ret <- ""
						break
					}
					if res.StatusCode < 200 || 300 <= res.StatusCode {
						errs <- fmt.Errorf("head request for %s was not successful: %s", url, res.Status)
						break
					}
					a.log.Println("Found:", url)
					ret <- r.githubURL(r.next)
				case <-done:
					return
				}
			}
		}(urls, errs, reqs, done)
	}

	go func(done <-chan struct{}) {
		for _, a := range a.actions {
			select {
			case reqs <- a:
			case <-done:
				return
			}
		}
	}(done)

	us := []string{}
	for i := 0; i < len(a.actions); i++ {
		select {
		case u := <-urls:
			if u != "" {
				us = append(us, u)
			}
		case err := <-errs:
			close(done)
			return nil, err
		}
	}
	close(done)

	sort.Strings(us)
	return us, nil
}

func (a *app) run(args []string) int {
	var source string
	var format string
	var quiet bool
	var detect bool

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.StringVar(&source, "s", "remote", "source of actions. \"remote\" or jsonl file path. \"remote\" fetches data from github.com")
	flags.StringVar(&format, "f", "go", "format of generated code output to stdout. \"go\" or \"jsonl\"")
	flags.BoolVar(&detect, "d", false, "detect new version of actions are released")
	flags.BoolVar(&quiet, "q", false, "disable log output to stderr")
	flags.SetOutput(a.stderr)
	flags.Usage = func() {
		fmt.Fprintln(a.stderr, `Usage: go run generate-popular-actions [FLAGS] [FILE]

  This tool fetches action.yml files of popular actions and generates code to
  given file. When no file path is given in arguments, this tool outputs
  generated code to stdout.

  It can fetch data from remote GitHub repositories and from local JSONL file
  (-s option). And it can output Go code or JSONL serialized data (-f option).

  When -d flag is given, it tries to detect new release for popular actions.
  When detecting some new releases, it shows their URLs to stdout and returns
  non-zero exit status.

Flags:`)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return 0 // When -h or -help
		}
		return 1
	}
	if flags.NArg() > 1 {
		fmt.Fprintf(a.stderr, "this command takes one or zero argument but given: %s\n", flags.Args())
		return 1
	}

	if quiet {
		w := log.Writer()
		defer func() { a.log.SetOutput(w) }()
		a.log.SetOutput(io.Discard)
	}

	a.log.Println("Start generate-popular-actions script")

	if detect {
		urls, err := a.detectNewReleaseURLs()
		if err != nil {
			fmt.Fprintln(a.stderr, err)
			return 1
		}
		if len(urls) == 0 {
			return 0
		}
		fmt.Fprintln(a.stdout, "Detected some new releases")
		for _, u := range urls {
			fmt.Fprintln(a.stdout, u)
		}
		return 2
	}

	if format != "go" && format != "jsonl" {
		fmt.Fprintf(a.stderr, "invalid value for -f option: %s\n", format)
		return 1
	}

	var actions map[string]*actionlint.ActionMetadata
	if source == "remote" {
		a.log.Println("Fetching data from https://github.com")
		m, err := a.fetchRemote()
		if err != nil {
			fmt.Fprintln(a.stderr, err)
			return 1
		}
		actions = m
	} else {
		a.log.Println("Fetching data from", source)
		m, err := a.readJSONL(source)
		if err != nil {
			fmt.Fprintln(a.stderr, err)
			return 1
		}
		actions = m
	}

	where := "stdout"
	out := a.stdout
	if flags.NArg() == 1 {
		where = flags.Arg(0)
		f, err := os.Create(where)
		if err != nil {
			fmt.Fprintf(a.stderr, "could not open file to output: %s\n", err)
			return 1
		}
		defer f.Close()
		out = f
	}

	switch format {
	case "go":
		a.log.Println("Generating Go source code to", where)
		if err := a.writeGo(out, actions); err != nil {
			fmt.Fprintln(a.stderr, err)
			return 1
		}
	case "jsonl":
		a.log.Println("Generating JSONL source to", where)
		if err := a.writeJSONL(out, actions); err != nil {
			fmt.Fprintln(a.stderr, err)
			return 1
		}
	}

	a.log.Println("Done generate-popular-actions script successfully")
	return 0
}

func main() {
	os.Exit(newApp(os.Stdout, os.Stderr, os.Stderr, popularActions, doNotCheckInputs, doNotCheckOutputs).run(os.Args))
}
