package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
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
	yamlExtYAML yamlExt = iota
	yamlExtYML
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
	tags []string
	ext  yamlExt
}

// Note: Actions used by top 1000 public repositories at GitHub sorted by number of occurrences:
// https://gist.github.com/rhysd/1db81fa80096b699b9c045f435d0cace

var popularActions = []action{
	{"8398a7/action-slack", []string{"v1", "v2", "v3"}, yamlExtYML},
	{"Azure/container-scan", []string{"v0"}, yamlExtYAML},
	{"EnricoMi/publish-unit-test-result-action", []string{"v1"}, yamlExtYML},
	{"JamesIves/github-pages-deploy-action", []string{"releases/v3", "releases/v4"}, yamlExtYML},
	{"ReactiveCircus/android-emulator-runner", []string{"v1", "v2"}, yamlExtYML},
	{"actions-cool/issues-helper", []string{"v1", "v2"}, yamlExtYML},
	{"actions-rs/audit-check", []string{"v1"}, yamlExtYML},
	{"actions-rs/cargo", []string{"v1"}, yamlExtYML},
	{"actions-rs/clippy-check", []string{"v1"}, yamlExtYML},
	{"actions-rs/toolchain", []string{"v1"}, yamlExtYML},
	{"actions/cache", []string{"v1", "v2"}, yamlExtYML},
	{"actions/checkout", []string{"v1", "v2"}, yamlExtYML},
	{"actions/delete-package-versions", []string{"v1"}, yamlExtYML},
	{"actions/download-artifact", []string{"v1", "v2"}, yamlExtYML},
	{"actions/first-interaction", []string{"v1"}, yamlExtYML},
	{"actions/github-script", []string{"v1", "v2", "v3", "v4"}, yamlExtYML},
	{"actions/labeler", []string{"v2", "v3"}, yamlExtYML}, // v1 does not exist
	{"actions/setup-dotnet", []string{"v1"}, yamlExtYML},
	{"actions/setup-go", []string{"v1", "v2"}, yamlExtYML},
	{"actions/setup-java", []string{"v1", "v2"}, yamlExtYML},
	{"actions/setup-node", []string{"v1", "v2"}, yamlExtYML},
	{"actions/setup-python", []string{"v1", "v2"}, yamlExtYML},
	{"actions/stale", []string{"v1", "v2", "v3", "v4"}, yamlExtYML},
	{"actions/upload-artifact", []string{"v1", "v2"}, yamlExtYML},
	{"aws-actions/configure-aws-credentials", []string{"v1"}, yamlExtYML},
	{"azure/aks-set-context", []string{"v1"}, yamlExtYML},
	{"azure/login", []string{"v1"}, yamlExtYML},
	{"bahmutov/npm-install", []string{"v1"}, yamlExtYML},
	{"codecov/codecov-action", []string{"v1", "v2"}, yamlExtYML},
	{"dawidd6/action-download-artifact", []string{"v2"}, yamlExtYML},
	{"dawidd6/action-send-mail", []string{"v1", "v2", "v3"}, yamlExtYML},
	{"dessant/lock-threads", []string{"v2"}, yamlExtYML}, // v1 does not exist
	{"docker/build-push-action", []string{"v1", "v2"}, yamlExtYML},
	{"docker/login-action", []string{"v1"}, yamlExtYML},
	{"docker/metadata-action", []string{"v1", "v2", "v3"}, yamlExtYML},
	{"docker/setup-buildx-action", []string{"v1"}, yamlExtYML},
	{"docker/setup-qemu-action", []string{"v1"}, yamlExtYML},
	{"dorny/paths-filter", []string{"v1", "v2"}, yamlExtYML},
	{"enriikke/gatsby-gh-pages-action", []string{"v2"}, yamlExtYML},
	{"erlef/setup-beam", []string{"v1"}, yamlExtYML},
	{"game-ci/unity-builder", []string{"v2"}, yamlExtYML},
	{"getsentry/paths-filter", []string{"v1", "v2"}, yamlExtYML},
	{"github/codeql-action/analyze", []string{"v1"}, yamlExtYML},
	{"github/codeql-action/autobuild", []string{"v1"}, yamlExtYML},
	{"github/codeql-action/init", []string{"v1"}, yamlExtYML},
	{"github/super-linter", []string{"v3", "v4"}, yamlExtYML},
	{"githubocto/flat", []string{"v1", "v2", "v3"}, yamlExtYML},
	{"golangci/golangci-lint-action", []string{"v1", "v2"}, yamlExtYML},
	{"goreleaser/goreleaser-action", []string{"v1", "v2"}, yamlExtYML},
	{"gradle/wrapper-validation-action", []string{"v1"}, yamlExtYML},
	{"haskell/actions/setup", []string{"v1"}, yamlExtYML},
	{"marvinpinto/action-automatic-releases", []string{"latest"}, yamlExtYML},
	{"microsoft/playwright-github-action", []string{"v1"}, yamlExtYML},
	{"mikepenz/release-changelog-builder-action", []string{"v1", "v2"}, yamlExtYML},
	{"msys2/setup-msys2", []string{"v1", "v2"}, yamlExtYML},
	{"ncipollo/release-action", []string{"v1"}, yamlExtYML},
	{"nwtgck/actions-netlify", []string{"v1"}, yamlExtYML},
	{"octokit/request-action", []string{"v1.x", "v2.x"}, yamlExtYML},
	{"peaceiris/actions-gh-pages", []string{"v2", "v3"}, yamlExtYML},
	{"peter-evans/create-pull-request", []string{"v1", "v2", "v3"}, yamlExtYML},
	{"preactjs/compressed-size-action", []string{"v1", "v2"}, yamlExtYML},
	{"reviewdog/action-actionlint", []string{"v1"}, yamlExtYML},
	{"reviewdog/action-eslint", []string{"v1"}, yamlExtYML},
	{"reviewdog/action-golangci-lint", []string{"v1"}, yamlExtYML},
	{"reviewdog/action-hadolint", []string{"v1"}, yamlExtYML},
	{"reviewdog/action-misspell", []string{"v1"}, yamlExtYML},
	{"reviewdog/action-rubocop", []string{"v1"}, yamlExtYML},
	{"reviewdog/action-shellcheck", []string{"v1"}, yamlExtYML},
	{"reviewdog/action-tflint", []string{"v1"}, yamlExtYML},
	{"rhysd/action-setup-vim", []string{"v1"}, yamlExtYML},
	{"ridedott/merge-me-action", []string{"v1", "v2"}, yamlExtYML},
	{"rtCamp/action-slack-notify", []string{"v2"}, yamlExtYML},
	{"ruby/setup-ruby", []string{"v1"}, yamlExtYML},
	{"shivammathur/setup-php", []string{"v1", "v2"}, yamlExtYML},
	{"softprops/action-gh-release", []string{"v1"}, yamlExtYML},
	{"subosito/flutter-action", []string{"v1"}, yamlExtYML},
	{"treosh/lighthouse-ci-action", []string{"v1", "v2", "v3", "v7", "v8"}, yamlExtYML},
	{"wearerequired/lint-action", []string{"v1"}, yamlExtYML},
}

func fetchRemote(actions []action) (map[string]*actionlint.ActionMetadata, error) {
	type request struct {
		slug string
		tag  string
		ext  yamlExt
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
					slug := req.slug
					path := ""
					if ss := strings.Split(slug, "/"); len(ss) > 2 {
						slug = ss[0] + "/" + ss[1]
						path = strings.Join(ss[2:], "/") + "/"
					}
					url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%saction.%s", slug, req.tag, path, req.ext.String())
					log.Println("Start fetching", url)
					res, err := c.Get(url)
					if err != nil {
						ret <- &fetched{err: fmt.Errorf("could not fetch %s: %w", url, err)}
						break
					}
					if res.StatusCode < 200 || 300 <= res.StatusCode {
						ret <- &fetched{err: fmt.Errorf("request was not successful %s: %s", url, res.Status)}
						break
					}
					body, err := ioutil.ReadAll(res.Body)
					res.Body.Close()
					if err != nil {
						ret <- &fetched{err: fmt.Errorf("could not read body for %s: %w", url, err)}
						break
					}
					spec := fmt.Sprintf("%s@%s", req.slug, req.tag)
					var meta actionlint.ActionMetadata
					if err := yaml.Unmarshal(body, &meta); err != nil {
						ret <- &fetched{err: fmt.Errorf("coult not parse metadata for %s: %w", url, err)}
						break
					}
					ret <- &fetched{spec: spec, meta: &meta}
				case <-done:
					return
				}
			}
		}(results, reqs, done)
	}

	n := 0
	for _, action := range actions {
		n += len(action.tags)
	}

	go func(reqs chan<- *request, done <-chan struct{}) {
		for _, action := range actions {
			for _, tag := range action.tags {
				select {
				case reqs <- &request{action.slug, tag, action.ext}:
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

func writeJSONL(out io.Writer, actions map[string]*actionlint.ActionMetadata) error {
	enc := json.NewEncoder(out)
	for spec, meta := range actions {
		j := actionJSON{spec, meta}
		if err := enc.Encode(&j); err != nil {
			return fmt.Errorf("could not encode action %q data into JSON: %w", spec, err)
		}
	}
	return nil
}

func writeGo(out io.Writer, actions map[string]*actionlint.ActionMetadata) error {
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

	defs := map[string]int{}
	defID := 0

	for _, spec := range specs {
		meta := actions[spec]
		fmt.Fprintf(b, "%q: {\n", spec)
		fmt.Fprintf(b, "Name: %q,\n", meta.Name)

		if len(meta.Inputs) > 0 {
			names := make([]string, 0, len(meta.Inputs))
			for n := range meta.Inputs {
				names = append(names, n)
			}
			sort.Strings(names)

			fmt.Fprintf(b, "Inputs: map[string]*ActionMetadataInput{\n")
			for _, name := range names {
				input := meta.Inputs[name]
				fmt.Fprintf(b, "%q: {\n", name)
				if input.Required {
					fmt.Fprintf(b, "Required: true,\n")
				}
				if input.Default != nil {
					var id int
					if i, ok := defs[*input.Default]; ok {
						id = i
					} else {
						defs[*input.Default] = defID
						id = defID
						defID++
					}
					fmt.Fprintf(b, "Default: &popularActionDefaultValue%d,\n", id)
				}
				fmt.Fprintf(b, "},\n")
			}
			fmt.Fprintf(b, "},\n")
		}

		if len(meta.Outputs) > 0 {
			names := make([]string, 0, len(meta.Outputs))
			for n := range meta.Outputs {
				names = append(names, n)
			}
			sort.Strings(names)

			fmt.Fprintf(b, "Outputs: map[string]struct{}{\n")
			for _, name := range names {
				fmt.Fprintf(b, "%q: {},\n", name)
			}
			fmt.Fprintf(b, "},\n")
		}

		fmt.Fprintf(b, "},\n")
	}

	fmt.Fprintln(b, "}")

	sortedDefs := make([]string, len(defs))
	for s, i := range defs {
		sortedDefs[i] = s
	}
	for i, s := range sortedDefs {
		fmt.Fprintf(b, "var popularActionDefaultValue%d = %q\n", i, s)
	}

	// Format the generated source with checking Go syntax
	gen := b.Bytes()
	src, err := format.Source(gen)
	if err != nil {
		return fmt.Errorf("could not format generated Go code: %w\n%s", err, gen)
	}

	if _, err := out.Write(src); err != nil {
		return fmt.Errorf("could not output generated Go source to stdout: %w", err)
	}
	return nil
}

func readJSONL(file string) (map[string]*actionlint.ActionMetadata, error) {
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

func run(args []string, stdout, stderr io.Writer, knownActions []action) int {
	var source string
	var format string
	var quiet bool

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.StringVar(&source, "s", "remote", "source of actions. \"remote\" or jsonl file path. \"remote\" fetches data from github.com")
	flags.StringVar(&format, "f", "go", "format of generated code output to stdout. \"go\" or \"jsonl\"")
	flags.BoolVar(&quiet, "q", false, "disable log output to stderr")
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, `Usage: go run generate-popular-actions [FLAGS] [FILE]

  This tool fetches action.yml files of popular actions and generates code to
  given file. When no file path is given in arguments, this tool outputs
  generated code to stdout. It can fetch data from remote GitHub repositories
  and from local JSONL file (-s option). And it can output Go code or JSONL
  serialized data (-f option).

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
		fmt.Fprintf(stderr, "this command takes one or zero argument but given: %s\n", flags.Args())
		return 1
	}

	if quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(stderr)
	}

	if format != "go" && format != "jsonl" {
		fmt.Fprintf(stderr, "invalid value for -f option: %s\n", format)
		return 1
	}

	where := "stdout"
	out := stdout
	if flags.NArg() == 1 {
		where = flags.Arg(0)
		f, err := os.Create(where)
		if err != nil {
			fmt.Fprintf(stderr, "could not open file to output: %s\n", err)
			return 1
		}
		defer f.Close()
		out = f
	}

	var actions map[string]*actionlint.ActionMetadata
	if source == "remote" {
		log.Println("Fetching data from https://github.com")
		m, err := fetchRemote(knownActions)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		actions = m
	} else {
		log.Println("Fetching data from", source)
		m, err := readJSONL(source)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		actions = m
	}

	switch format {
	case "go":
		log.Println("Generating Go source code to", where)
		if err := writeGo(out, actions); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	case "jsonl":
		log.Println("Generating JSONL source to", where)
		if err := writeJSONL(out, actions); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}

	log.Println("Done")
	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr, popularActions))
}
