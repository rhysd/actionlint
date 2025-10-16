package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/rhysd/actionlint"
	"go.yaml.in/yaml/v4"
)

// List of known outdated actions which cannot be detected from 'runs' in action.yml
var outdatedActions = []string{
	"actions/labeler@v1",
	"actions/checkout@v1",
	"actions/upload-artifact@v1",
	"actions/download-artifact@v1",
}

type actionOutput struct {
	Spec     string                     `json:"spec"`
	Meta     *actionlint.ActionMetadata `json:"metadata"`
	Outdated bool                       `json:"outdated"`
}

type registry struct {
	Slug    string   `json:"slug"`
	Path    string   `json:"path"`
	Tags    []string `json:"tags"`
	Next    string   `json:"next"`
	FileExt string   `json:"file_ext"`
	// slugs not to check inputs. Some actions allow to specify inputs which are not defined in action.yml.
	// In such cases, actionlint no longer can check the inputs, but it can still check outputs. (#16)
	SkipInputs bool `json:"skip_inputs"`
	// slugs which allows any outputs to be set. Some actions sets outputs 'dynamically'. Those outputs
	// may or may not exist. And they are not listed in action.yml metadata. actionlint cannot check
	// such outputs and fallback into allowing to set any outputs. (#18)
	SkipOutputs bool `json:"skip_outputs"`
}

func (r *registry) rawURL(tag string) string {
	ext := "yml"
	if r.FileExt != "" {
		ext = r.FileExt
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s%s/action.%s", r.Slug, tag, r.Path, ext)
}

func (r *registry) githubURL(tag string) string {
	return fmt.Sprintf("https://github.com/%s/tree/%s%s", r.Slug, tag, r.Path)
}

func (r *registry) spec(tag string) string {
	return fmt.Sprintf("%s%s@%s", r.Slug, r.Path, tag)
}

// Note: Actions used by top 1000 public repositories at GitHub sorted by number of occurrences:
// https://gist.github.com/rhysd/1db81fa80096b699b9c045f435d0cace

//go:embed popular_actions.json
var defaultPopularActionsJSON []byte

const minNodeRunnerVersion = 20

func isOutdated(spec, runs string) bool {
	for _, s := range outdatedActions {
		if s == spec {
			return true
		}
	}
	if !strings.HasPrefix(runs, "node") {
		return false
	}
	v, err := strconv.ParseUint(runs[len("node"):], 10, 8)
	return err == nil && v < minNodeRunnerVersion
}

type gen struct {
	stdout      io.Writer
	stderr      io.Writer
	log         *log.Logger
	rawRegistry []byte
}

func newGen(stdout, stderr, dbgout io.Writer) *gen {
	l := log.New(dbgout, "", log.LstdFlags)
	return &gen{stdout, stderr, l, defaultPopularActionsJSON}
}

func (g *gen) registry() ([]*registry, error) {
	var a []*registry
	if err := json.Unmarshal(g.rawRegistry, &a); err != nil {
		return nil, fmt.Errorf("could not parse the local action registry file as JSON: %w", err)
	}
	return a, nil
}

func (g *gen) fetchRemote() (map[string]*actionlint.ActionMetadata, error) {
	type request struct {
		action *registry
		tag    string
	}

	type fetched struct {
		spec string
		meta *actionlint.ActionMetadata
		err  error
	}

	actions, err := g.registry()
	if err != nil {
		return nil, err
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
					g.log.Println("Start fetching", url)
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
					if req.action.SkipInputs {
						meta.SkipInputs = true
					}
					if req.action.SkipOutputs {
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
	for _, action := range actions {
		n += len(action.Tags)
	}

	go func(reqs chan<- *request, done <-chan struct{}) {
		for _, action := range actions {
			for _, tag := range action.Tags {
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

		// Workaround for #416.
		// Once this PR is merged, remove this `if` statement and regenerate popular_actions.go.
		// https://github.com/dorny/paths-filter/pull/236
		if f.spec == "dorny/paths-filter@v3" {
			f.meta.Inputs["predicate-quantifier"] = &actionlint.ActionMetadataInput{
				Name:     "predicate-quantifier",
				Required: false,
			}
		}

		// Workaround for #442.
		// https://github.com/actions/download-artifact/issues/355
		if f.spec == "actions/download-artifact@v3-node20" {
			if f.meta.Outputs == nil {
				f.meta.Outputs = actionlint.ActionMetadataOutputs{}
			}
			f.meta.Outputs["download-path"] = &actionlint.ActionMetadataOutput{
				Name: "download-path",
			}
		}

		ret[f.spec] = f.meta
	}

	close(done)
	return ret, nil
}

func (g *gen) writeJSONL(out io.Writer, actions map[string]*actionlint.ActionMetadata) error {
	enc := json.NewEncoder(out)
	for spec, meta := range actions {
		j := actionOutput{spec, meta, isOutdated(spec, meta.Runs.Using)}
		if err := enc.Encode(&j); err != nil {
			return fmt.Errorf("could not encode action %q data into JSON: %w", spec, err)
		}
	}
	g.log.Printf("Wrote %d action metadata as JSONL", len(actions))
	return nil
}

func (g *gen) writeGo(out io.Writer, actions map[string]*actionlint.ActionMetadata) error {
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

	outdated := []string{}
	for _, spec := range specs {
		meta := actions[spec]
		if isOutdated(spec, meta.Runs.Using) {
			outdated = append(outdated, spec)
			continue
		}

		fmt.Fprintf(b, "%q: {\n", spec)
		fmt.Fprintf(b, "Name: %q,\n", meta.Name)

		if meta.SkipInputs {
			fmt.Fprintf(b, "SkipInputs: true,\n")
		}

		if len(meta.Inputs) > 0 && !meta.SkipInputs {
			ids := make([]string, 0, len(meta.Inputs))
			for n := range meta.Inputs {
				ids = append(ids, n)
			}
			sort.Strings(ids)

			fmt.Fprintf(b, "Inputs: ActionMetadataInputs{\n")
			for _, id := range ids {
				i := meta.Inputs[id]
				fmt.Fprintf(b, "%q: {%q, %v, %v},\n", id, i.Name, i.Required, i.Deprecated)
			}
			fmt.Fprintf(b, "},\n")
		}

		if meta.SkipOutputs {
			fmt.Fprintf(b, "SkipOutputs: true,\n")
		}

		if len(meta.Outputs) > 0 && !meta.SkipOutputs {
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

	fmt.Fprintln(b, `// OutdatedPopularActionSpecs is a spec set of known outdated popular actions. The word 'outdated'
// means that the runner used by the action is no longer available such as "node12", "node16".
var OutdatedPopularActionSpecs = map[string]struct{}{`)
	for _, s := range outdated {
		fmt.Fprintf(b, "%q: {},\n", s)
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

	g.log.Printf("Wrote %d action metadata and %d outdated action specs as Go", len(actions)-len(outdated), len(outdated))
	return nil
}

func (g *gen) readJSONL(file string) (map[string]*actionlint.ActionMetadata, error) {
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
			g.log.Printf("Read %d action metadata from %s", len(ret), file)
			return ret, nil
		} else if err != nil {
			return nil, fmt.Errorf("could not read line in file %s: %w", file, err)
		}
		var j actionOutput
		if err := json.Unmarshal(l, &j); err != nil {
			return nil, fmt.Errorf("could not parse line as JSON for action metadata in file %s: %w", file, err)
		}
		ret[j.Spec] = j.Meta
	}
}

func (g *gen) detectNewReleaseURLs() ([]string, error) {
	all, err := g.registry()
	if err != nil {
		return nil, err
	}

	// Filter actions which have no next versions
	actions := []*registry{}
	for _, a := range all {
		if a.Next != "" {
			actions = append(actions, a)
		}
	}

	g.log.Println("Start detecting new versions in", len(actions), "repositories")

	urls := make(chan string)
	done := make(chan struct{})
	errs := make(chan error)
	reqs := make(chan *registry)

	for i := 0; i < 4; i++ {
		go func(ret chan<- string, errs chan<- error, reqs <-chan *registry, done <-chan struct{}) {
			var c http.Client
			for {
				select {
				case r := <-reqs:
					url := r.rawURL(r.Next)
					g.log.Println("Checking", url)
					res, err := c.Head(url)
					if err != nil {
						errs <- fmt.Errorf("could not send head request to %s: %w", url, err)
						break
					}
					if res.StatusCode == 404 {
						g.log.Println("Not found:", url)
						ret <- ""
						break
					}
					if res.StatusCode < 200 || 300 <= res.StatusCode {
						errs <- fmt.Errorf("head request for %s was not successful: %s", url, res.Status)
						break
					}
					g.log.Println("Found:", url)
					ret <- r.githubURL(r.Next)
				case <-done:
					return
				}
			}
		}(urls, errs, reqs, done)
	}

	go func(done <-chan struct{}) {
		for _, a := range actions {
			select {
			case reqs <- a:
			case <-done:
				return
			}
		}
	}(done)

	us := []string{}
	for i := 0; i < len(actions); i++ {
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

	g.log.Println("Done detecting new versions in", len(actions), "repositories")
	return us, nil
}

func (g *gen) run(args []string) int {
	var source string
	var format string
	var quiet bool
	var detect bool
	var registry string

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.StringVar(&source, "s", "", "source of actions as local jsonl file path instead of fetching actions metadata from github.com")
	flags.StringVar(&format, "f", "go", `format of generated code output to stdout. "go" or "jsonl"`)
	flags.StringVar(&registry, "r", "", "registry of actions as local JSON file path. when this flag is not given, the default popular actions registry will be used")
	flags.BoolVar(&detect, "d", false, "detect new version of actions are released")
	flags.BoolVar(&quiet, "q", false, "disable log output to stderr")
	flags.SetOutput(g.stderr)
	flags.Usage = func() {
		fmt.Fprintln(g.stderr, `Usage: go run generate-popular-actions [FLAGS] [FILE]

  This tool fetches action.yml files of popular actions and generates code to
  given file. When no file path is given in arguments, this tool outputs
  generated code to stdout.

  It can fetch data from remote GitHub repositories and from local JSONL file
  (-s option). And it can output Go code or JSONL serialized data (-f option).

  What actions to be included is defined in the popular actions registry embedded
  in the executable. To use your own registry JSON file, use -r option.

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
		fmt.Fprintf(g.stderr, "this command takes one or zero argument but given: %s\n", flags.Args())
		return 1
	}

	if quiet {
		w := log.Writer()
		defer func() { g.log.SetOutput(w) }()
		g.log.SetOutput(io.Discard)
	}
	if registry != "" {
		b, err := os.ReadFile(registry)
		if err != nil {
			fmt.Fprintf(g.stderr, "could not read the file for actions registry: %s\n", err)
			return 1
		}
		g.rawRegistry = b
	}

	g.log.Println("Start generate-popular-actions script")

	if detect {
		urls, err := g.detectNewReleaseURLs()
		if err != nil {
			fmt.Fprintln(g.stderr, err)
			return 1
		}
		if len(urls) == 0 {
			fmt.Fprintln(g.stdout, "No new release was found")
			return 0
		}
		fmt.Fprintln(g.stdout, "Detected some new releases")
		for _, u := range urls {
			fmt.Fprintln(g.stdout, u)
		}
		return 2
	}

	if format != "go" && format != "jsonl" {
		fmt.Fprintf(g.stderr, "invalid value for -f option: %s\n", format)
		return 1
	}

	var actions map[string]*actionlint.ActionMetadata
	if source == "" {
		g.log.Println("Fetching data from https://github.com")
		m, err := g.fetchRemote()
		if err != nil {
			fmt.Fprintln(g.stderr, err)
			return 1
		}
		actions = m
	} else {
		g.log.Println("Fetching data from", source)
		m, err := g.readJSONL(source)
		if err != nil {
			fmt.Fprintln(g.stderr, err)
			return 1
		}
		actions = m
	}

	where := "stdout"
	out := g.stdout
	if flags.NArg() == 1 {
		where = flags.Arg(0)
		f, err := os.Create(where)
		if err != nil {
			fmt.Fprintf(g.stderr, "could not open file to output: %s\n", err)
			return 1
		}
		defer f.Close()
		out = f
	}

	switch format {
	case "go":
		g.log.Println("Generating Go source code to", where)
		if err := g.writeGo(out, actions); err != nil {
			fmt.Fprintln(g.stderr, err)
			return 1
		}
	case "jsonl":
		g.log.Println("Generating JSONL source to", where)
		if err := g.writeJSONL(out, actions); err != nil {
			fmt.Fprintln(g.stderr, err)
			return 1
		}
	}

	g.log.Println("Done generate-popular-actions script successfully")
	return 0
}

func main() {
	os.Exit(newGen(os.Stdout, os.Stderr, os.Stderr).run(os.Args))
}
