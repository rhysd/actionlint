package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
	Spec string                 `json:"spec"`
	Meta *actionlint.ActionSpec `json:"metadata"`
}

type action struct {
	slug string
	tags []string
	ext  yamlExt
}

var popularActions = []action{
	{"actions/checkout", []string{"v1", "v2"}, yamlExtYML},
}

func fetchRemote(actions []action) (map[string]*actionlint.ActionSpec, error) {
	type request struct {
		slug string
		tag  string
		ext  yamlExt
	}

	type fetched struct {
		spec string
		meta *actionlint.ActionSpec
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
					url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/action.%s", req.slug, req.tag, req.ext.String())
					res, err := c.Get(url)
					if err != nil {
						ret <- &fetched{err: fmt.Errorf("could not fetch %s: %w", url, err)}
						break
					}
					body, err := ioutil.ReadAll(res.Body)
					res.Body.Close()
					if err != nil {
						ret <- &fetched{err: fmt.Errorf("could not read body for %s: %w", url, err)}
						break
					}
					spec := fmt.Sprintf("%s@%s", req.slug, req.tag)
					var meta actionlint.ActionSpec
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
		for _, tag := range action.tags {
			reqs <- &request{action.slug, tag, action.ext}
			n++
		}
	}

	ret := make(map[string]*actionlint.ActionSpec, n)
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

func writeJSONL(out io.Writer, actions map[string]*actionlint.ActionSpec) error {
	enc := json.NewEncoder(out)
	for spec, meta := range actions {
		j := actionJSON{spec, meta}
		if err := enc.Encode(&j); err != nil {
			return err
		}
	}
	return nil
}

func writeGo(out io.Writer, actions map[string]*actionlint.ActionSpec) {
	fmt.Fprint(out, `package actionlint

import (
	"github.com/rhysd/actionlint"
)

func addressOfString(s string) *string {
	return &s // Go does not allow get pointer of string literal
}

// PopularActions is data set of known popular actions. Keys are specs (owner/repo@ref) of actions
// and values are their metadata.
var PopularActions = map[string]*actionlint.ActionSpec{
`)

	specs := make([]string, 0, len(actions))
	for s := range actions {
		specs = append(specs, s)
	}
	sort.Strings(specs)

	for _, spec := range specs {
		meta := actions[spec]
		fmt.Fprintf(out, "\t%q: {\n", spec)
		fmt.Fprintf(out, "\t\tName: %q,\n", meta.Name)

		if len(meta.Inputs) > 0 {
			names := make([]string, 0, len(meta.Inputs))
			for n := range meta.Inputs {
				names = append(names, n)
			}
			sort.Strings(names)

			fmt.Fprintf(out, "\t\tInputs: map[string]*actionlint.ActionInput{\n")
			for _, name := range names {
				input := meta.Inputs[name]
				fmt.Fprintf(out, "\t\t\t%q: {\n", name)
				if input.Required {
					fmt.Fprintf(out, "\t\t\t\tRequired: true,\n")
				}
				if input.Default != nil {
					fmt.Fprintf(out, "\t\t\t\tDefault: addressOfString(%q),\n", *input.Default)
				}
				fmt.Fprintf(out, "\t\t\t\tDescription: %q,\n", input.Description)
				fmt.Fprintf(out, "\t\t\t},\n")
			}
			fmt.Fprintf(out, "\t\t},\n")
		}

		if len(meta.Outputs) > 0 {
			names := make([]string, 0, len(meta.Outputs))
			for n := range meta.Outputs {
				names = append(names, n)
			}
			sort.Strings(names)

			fmt.Fprintf(out, "\t\tOutputs: map[string]string{\n")
			for _, name := range names {
				desc := meta.Outputs
				fmt.Fprintf(out, "\t\t\t%q: %q,\n", name, desc)
			}
			fmt.Fprintf(out, "\t\t},\n")
		}

		fmt.Fprintf(out, "\t},\n")
	}

	fmt.Fprintln(out, "}")
}

func readJSONL(file string) (map[string]*actionlint.ActionSpec, error) {
	if !strings.HasSuffix(file, ".jsonl") {
		return nil, fmt.Errorf("JSONL file name must end with \".jsonl\": %s", file)
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", file, err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	ret := map[string]*actionlint.ActionSpec{}
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

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.StringVar(&source, "s", "remote", "source of actions. \"remote\" or jsonl file")
	flags.StringVar(&format, "f", "go", "format of generated code output to stdout. \"go\" or \"jsonl\"")
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, `Usage: go run generate-popular-actions [FLAGS]

  This tool fetches action.yml files of popular actions and generates code to
  stdout. It can fetch data from remote GitHub repositories and from local
  JSONL file (-s flag). And it can output Go code or JSONL serialized data
  (-f option).

Flags:`)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return 0 // When -h or -help
		}
		return 1
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "this command takes no argument but given: %s\n", flags.Args())
		return 1
	}

	var actions map[string]*actionlint.ActionSpec
	if source == "remote" {
		m, err := fetchRemote(knownActions)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		actions = m
	} else {
		m, err := readJSONL(source)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		actions = m
	}

	switch format {
	case "go":
		writeGo(stdout, actions)
	case "jsonl":
		if err := writeJSONL(stdout, actions); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	default:
		fmt.Fprintf(stderr, "invalid -format value: %s\n", format)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr, popularActions))
}
