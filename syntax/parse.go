package syntax

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// START: TEMP
func kindString(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "Document"
	case yaml.SequenceNode:
		return "Sequence"
	case yaml.MappingNode:
		return "Mapping"
	case yaml.ScalarNode:
		return "Scalar"
	case yaml.AliasNode:
		return "Arias"
	default:
		panic("unreachable")
	}
}

func dump(n *yaml.Node, level int) {
	fmt.Printf("%s%s (%s, %d,%d): %q\n", strings.Repeat("  ", level), kindString(n.Kind), n.Tag, n.Line, n.Column, n.Value)
	for _, c := range n.Content {
		dump(c, level+1)
	}
}

// END: TEMP

func assert(b bool) {
	if !b {
		panic("assertion failed")
	}
}

func nodeKindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "arias"
	default:
		panic("unreachable")
	}
}

func posAt(n *yaml.Node) *Pos {
	return &Pos{n.Line, n.Column}
}

func isNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!null"
}

type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
}

func Parse(b []byte) (*Workflow, []*ParseError) {
	var n yaml.Node

	if err := yaml.Unmarshal(b, &n); err != nil {
		msg := fmt.Sprintf("Could not parse as YAML: %s", err.Error())
		return nil, []*ParseError{{msg, 0, 0}}
	}

	fmt.Println("DEBUG START")
	dump(&n, 0)
	fmt.Println("DEBUG END")

	p := &parser{}
	w := p.parse(&n)

	return w, p.errors
}

type keyVal struct {
	key *String
	val *yaml.Node
}

type parser struct {
	errors []*ParseError
}

func (p *parser) error(n *yaml.Node, m string) {
	p.errors = append(p.errors, &ParseError{m, n.Line, n.Column})
}

func (p *parser) errorAt(pos *Pos, m string) {
	p.errors = append(p.errors, &ParseError{m, pos.Line, pos.Col})
}

func (p *parser) errorfAt(pos *Pos, format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	p.errorAt(pos, m)
}

func (p *parser) errorf(n *yaml.Node, format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	p.error(n, m)
}

func (p *parser) unexpectedKey(s *String, sec string, expected []string) {
	l := len(expected)
	var m string
	if l == 1 {
		m = fmt.Sprintf("expected %q key for %q section but got %q", expected[0], sec, s.Value)
	} else if l > 1 {
		m = fmt.Sprintf("unexpected key %q for %q section. expected one of %v", s.Value, sec, expected)
	} else {
		m = fmt.Sprintf("unexpected key %q for %q section", s.Value, sec)
	}
	p.errors = append(p.errors, &ParseError{m, s.Pos.Line, s.Pos.Col})
}

func (p *parser) checkSequence(sec string, n *yaml.Node) bool {
	if n.Kind != yaml.SequenceNode {
		p.errorf(n, "%q section must be sequence node but got %s node", sec, nodeKindName(n.Kind))
		return false
	}
	return true
}

func (p *parser) parseString(n *yaml.Node) *String {
	// Do not check n.Tag is !!str because we don't need to check the node is string strictly.
	// In almost all cases, other nodes (like 42) are handled as string with its string representation.
	if n.Kind != yaml.ScalarNode {
		p.errorf(n, "expected scalar node for string value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return nil
	}
	return &String{n.Value, posAt(n)}
}

func (p *parser) parseStringSequence(sec string, n *yaml.Node) []*String {
	if ok := p.checkSequence(sec, n); !ok {
		return nil
	}
	ss := make([]*String, 0, len(n.Content))
	for _, c := range n.Content {
		s := p.parseString(c)
		if s != nil {
			ss = append(ss, s)
		}
	}
	return ss
}

func (p *parser) parseBool(n *yaml.Node) *Bool {
	if n.Kind != yaml.ScalarNode || n.Tag != "!!bool" {
		p.errorf(n, "expected bool value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return nil
	}
	return &Bool{n.Value == "true", posAt(n)}
}

func (p *parser) parseInt(n *yaml.Node) *Int {
	if n.Kind != yaml.ScalarNode {
		p.errorf(n, "expected scalar node for integer value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return nil
	}
	i, err := strconv.Atoi(n.Value)
	if err != nil {
		p.errorf(n, "invalid integer value: %q: %s", n.Value, err.Error())
		return nil
	}
	return &Int{i, posAt(n)}
}

func (p *parser) parseFloat(n *yaml.Node) *Float {
	if n.Kind != yaml.ScalarNode {
		p.errorf(n, "expected scalar node for float value but found %s node with %q tag", nodeKindName(n.Kind), n.Tag)
		return nil
	}
	f, err := strconv.ParseFloat(n.Value, 64)
	if err != nil {
		p.errorf(n, "invalid float value: %q: %s", n.Value, err.Error())
		return nil
	}
	return &Float{f, posAt(n)}
}

func (p *parser) parseMapping(what string, n *yaml.Node, allowEmpty bool) []keyVal {
	if !isNull(n) && n.Kind != yaml.MappingNode {
		p.errorf(n, "%s is %s node but mapping node is expected", what, nodeKindName(n.Kind))
		return nil
	}

	l := len(n.Content) / 2
	keys := make(map[string]struct{}, l)
	m := make([]keyVal, 0, l)
	for i := 0; i < l; i++ {
		k := p.parseString(n.Content[i])
		if k == nil {
			continue
		}
		if _, ok := keys[k.Value]; ok {
			p.errorf(n, "key %q is duplicate in %s", k.Value, what)
			continue
		}
		m = append(m, keyVal{k, n.Content[i+1]})
		keys[k.Value] = struct{}{}
	}

	if !allowEmpty && len(m) == 0 {
		p.errorf(n, "%s should not be empty. please remove this section if it's unnecessary", what)
	}

	return m
}

func (p *parser) parseSectionMapping(sec string, n *yaml.Node, allowEmpty bool) []keyVal {
	return p.parseMapping(fmt.Sprintf("%q section", sec), n, allowEmpty)
}

func (p *parser) parseScheduleEvent(pos *Pos, n *yaml.Node) *ScheduledEvent {
	if ok := p.checkSequence("schedule", n); !ok {
		return nil
	}

	cron := make([]*String, 0, len(n.Content))
	for _, c := range n.Content {
		m := p.parseMapping("element of \"schedule\" section", c, false)
		if len(m) != 1 || m[0].key.Value != "cron" {
			p.error(c, "element of \"schedule\" section must be mapping and must contain one key \"cron\"")
			continue
		}
		s := p.parseString(c)
		if s != nil {
			cron = append(cron, s)
		}
	}

	return &ScheduledEvent{cron, pos}
}

// https://docs.github.com/en/actions/reference/events-that-trigger-workflows#workflow_dispatch
func (p *parser) parseWorkflowDispatchEvent(pos *Pos, n *yaml.Node) *WorkflowDispatchEvent {
	ret := &WorkflowDispatchEvent{Pos: pos}

	for _, kv := range p.parseSectionMapping("workflow_dispatch", n, true) {
		if kv.key.Value == "inputs" {
			inputs := p.parseSectionMapping("inputs", kv.val, true)
			for _, input := range inputs {
				name, spec := input.key, input.val

				var desc *String
				var req *Bool
				var def *String

				for _, attr := range p.parseMapping("input settings of workflow_dispatch event", spec, true) {
					switch attr.key.Value {
					case "description":
						desc = p.parseString(attr.val)
					case "required":
						req = p.parseBool(attr.val)
					case "default":
						def = p.parseString(attr.val)
					default:
						p.unexpectedKey(attr.key, "inputs", []string{"description", "required", "default"})
					}
				}

				ret.Inputs[name.Value] = &DispatchInput{
					Name:        name,
					Description: desc,
					Required:    req,
					Default:     def,
				}
			}
		} else {
			p.unexpectedKey(kv.key, "workflow_dispatch", []string{"inputs"})
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/events-that-trigger-workflows#repository_dispatch
func (p *parser) parseRepositoryDispatchEvent(pos *Pos, n *yaml.Node) *RepositoryDispatchEvent {
	ret := &RepositoryDispatchEvent{Pos: pos}

	for _, kv := range p.parseSectionMapping("repository_dispatch", n, false) {
		if kv.key.Value == "types" {
			ret.Types = p.parseStringSequence("types", kv.val)
		} else {
			p.unexpectedKey(kv.key, "repository_dispatch", []string{"types"})
		}
	}

	return ret
}

func (p *parser) parseWebhookEvent(name *String, n *yaml.Node) *WebhookEvent {
	ret := &WebhookEvent{Hook: name, Pos: name.Pos}

	for _, kv := range p.parseSectionMapping(name.Value, n, true) {
		switch kv.key.Value {
		case "types":
			ret.Types = p.parseStringSequence(kv.key.Value, kv.val)
		case "branches":
			ret.Branches = p.parseStringSequence(kv.key.Value, kv.val)
		case "branches-ignore":
			ret.BranchesIgnore = p.parseStringSequence(kv.key.Value, kv.val)
		case "tags":
			ret.Tags = p.parseStringSequence(kv.key.Value, kv.val)
		case "tags-ignore":
			ret.TagsIgnore = p.parseStringSequence(kv.key.Value, kv.val)
		case "paths":
			ret.Paths = p.parseStringSequence(kv.key.Value, kv.val)
		case "paths-ignore":
			ret.PathsIgnore = p.parseStringSequence(kv.key.Value, kv.val)
		case "workflows":
			ret.Workflows = p.parseStringSequence(kv.key.Value, kv.val)
		default:
			p.unexpectedKey(kv.key, name.Value, []string{
				"types",
				"branches",
				"branches-ignore",
				"tags",
				"tags-ignore",
				"paths",
				"paths-ignore",
				"workflows",
			})
		}
	}

	return ret
}

func (p *parser) parseEvents(n *yaml.Node) []Event {
	switch n.Kind {
	case yaml.MappingNode:
		kvs := p.parseSectionMapping("on", n, false)
		ret := make([]Event, 0, len(kvs))

		for _, kv := range kvs {
			switch kv.key.Value {
			case "schedule":
				if e := p.parseScheduleEvent(kv.key.Pos, kv.val); e != nil {
					ret = append(ret, e)
				}
			case "workflow_dispatch":
				ret = append(ret, p.parseWorkflowDispatchEvent(kv.key.Pos, kv.val))
			case "repository_dispatch":
				ret = append(ret, p.parseRepositoryDispatchEvent(kv.key.Pos, kv.val))
			default:
				ret = append(ret, p.parseWebhookEvent(kv.key, kv.val))
			}
		}

		return ret
	case yaml.SequenceNode:
		ret := make([]Event, 0, len(n.Content))

		for _, c := range n.Content {
			if s := p.parseString(c); s != nil {
				switch s.Value {
				case "schedule", "workflow_dispatch", "repository_dispatch":
					p.errorf(c, "%q event should not be listed in sequence. Use mapping for \"on\" section and configure the event as value of the mapping", s.Value)
				default:
					ret = append(ret, &WebhookEvent{Hook: s, Pos: posAt(c)})
				}
			}
		}

		return ret
	default:
		p.errorf(n, "\"on\" section value is expected to be mapping or sequence but found %s node", nodeKindName(n.Kind))
		return nil
	}
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#permissions
func (p *parser) parsePermissions(pos *Pos, n *yaml.Node) *Permissions {
	ret := &Permissions{Pos: pos}

	if n.Kind == yaml.ScalarNode {
		var kind PermKind
		switch n.Value {
		case "read-all":
			kind = PermKindRead
		case "write-all":
			kind = PermKindWrite
		default:
			p.errorf(n, "permission must be one of \"read-all\", \"write-all\" but got %q", n.Value)
		}
		ret.All = &Permission{nil, kind, posAt(n)}
	} else {
		m := p.parseSectionMapping("permissions", n, false)
		scopes := make(map[string]*Permission, len(m))

		for _, kv := range m {
			perm := p.parseString(kv.val).Value
			kind := PermKindNone
			switch perm {
			case "read":
				kind = PermKindRead
			case "write":
				kind = PermKindWrite
			case "none":
				kind = PermKindNone
			default:
				p.errorf(kv.val, "permission must be one of \"none\", \"read\", \"write\" but got %q", perm)
				continue
			}
			scopes[kv.key.Value] = &Permission{
				Name: kv.key,
				Kind: kind,
				Pos:  kv.key.Pos,
			}
		}

		ret.Scopes = scopes
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#env
func (p *parser) parseEnv(n *yaml.Node) map[string]*EnvVar {
	m := p.parseMapping("env", n, false)
	ret := make(map[string]*EnvVar, len(m))

	for _, kv := range m {
		ret[kv.key.Value] = &EnvVar{
			Name:  kv.key,
			Value: p.parseString(kv.val),
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#defaults
func (p *parser) parseDefaults(pos *Pos, n *yaml.Node) *Defaults {
	ret := &Defaults{Pos: pos}

	for _, kv := range p.parseSectionMapping("defaults", n, false) {
		if kv.key.Value != "run" {
			p.unexpectedKey(kv.key, "defaults", []string{"run"})
			continue
		}
		ret.Run = &DefaultsRun{Pos: kv.key.Pos}

		for _, attr := range p.parseSectionMapping("run", kv.val, false) {
			switch attr.key.Value {
			case "shell":
				ret.Run.Shell = p.parseString(attr.val)
			case "working-directory":
				ret.Run.WorkingDirectory = p.parseString(attr.val)
			default:
				p.unexpectedKey(attr.key, "run", []string{"shell", "working-directory"})
			}
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idconcurrency
func (p *parser) parseConcurrency(pos *Pos, n *yaml.Node) *Concurrency {
	ret := &Concurrency{Pos: pos}

	if n.Kind == yaml.ScalarNode {
		ret.Group = p.parseString(n)
	} else {
		for _, kv := range p.parseSectionMapping("concurrency", n, false) {
			switch kv.key.Value {
			case "group":
				ret.Group = p.parseString(kv.val)
			case "cancel-in-progress":
				ret.CancelInProgress = p.parseBool(kv.val)
			default:
				p.unexpectedKey(kv.key, "concurrency", []string{"group", "cancel-in-progress"})
			}
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idenvironment
func (p *parser) parseEnvironment(pos *Pos, n *yaml.Node) *Environment {
	ret := &Environment{Pos: pos}

	if n.Kind == yaml.ScalarNode {
		ret.Name = p.parseString(n)
	} else {
		for _, kv := range p.parseSectionMapping("environment", n, false) {
			switch kv.key.Value {
			case "name":
				ret.Name = p.parseString(kv.val)
			case "url":
				ret.URL = p.parseString(kv.val)
			default:
				p.unexpectedKey(kv.key, "environment", []string{"name", "url"})
			}
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idoutputs
func (p *parser) parseOutputs(n *yaml.Node) map[string]*Output {
	outputs := p.parseSectionMapping("outputs", n, false)
	ret := make(map[string]*Output, len(outputs))
	for _, output := range outputs {
		ret[output.key.Value] = &Output{
			Name:  output.key,
			Value: p.parseString(output.val),
		}
	}
	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-including-additional-values-into-combinations
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-excluding-configurations-from-a-matrix
func (p *parser) parseMatrixCombinations(sec string, n *yaml.Node) []map[string]*MatrixCombination {
	if ok := p.checkSequence(sec, n); !ok {
		return nil
	}

	ret := make([]map[string]*MatrixCombination, 0, len(n.Content))
	for _, c := range n.Content {
		kvs := p.parseMapping(fmt.Sprintf("element in %q section", sec), c, false)
		elem := make(map[string]*MatrixCombination, len(kvs))
		for _, kv := range kvs {
			s := p.parseString(kv.val)
			if s != nil {
				elem[kv.key.Value] = &MatrixCombination{kv.key, s}
			}
		}
		ret = append(ret, elem)
	}
	return ret
}

func (p *parser) parseMatrixRows(key *String, n *yaml.Node) *MatrixRow {
	return &MatrixRow{
		Name:   key,
		Values: p.parseStringSequence("matrix", n),
	}
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
func (p *parser) parseMatrix(pos *Pos, n *yaml.Node) *Matrix {
	ret := &Matrix{Pos: pos}

	for _, kv := range p.parseSectionMapping("matrix", n, false) {
		switch kv.key.Value {
		case "include":
			ret.Include = p.parseMatrixCombinations("include", kv.val)
		case "exclude":
			ret.Include = p.parseMatrixCombinations("exclude", kv.val)
		default:
			ret.Rows[kv.key.Value] = &MatrixRow{
				Name:   kv.key,
				Values: p.parseStringSequence("matrix", kv.val),
			}
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategy
func (p *parser) parseStrategy(pos *Pos, n *yaml.Node) *Strategy {
	ret := &Strategy{Pos: pos}

	for _, kv := range p.parseSectionMapping("strategy", n, false) {
		switch kv.key.Value {
		case "matrix":
			ret.Matrix = p.parseMatrix(kv.key.Pos, kv.val)
		case "fail-fast":
			ret.FailFast = p.parseBool(kv.val)
		case "max-parallel":
			ret.MaxParallel = p.parseInt(kv.val)
		default:
			p.unexpectedKey(kv.key, "strategy", []string{"matrix", "fail-fast", "max-parallel"})
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainer
func (p *parser) parseContainer(sec string, pos *Pos, n *yaml.Node) *Container {
	ret := &Container{Pos: pos}

	for _, kv := range p.parseSectionMapping(sec, n, false) {
		switch kv.key.Value {
		case "image":
			ret.Image = p.parseString(kv.val)
		case "credentials":
			cred := &Credentials{Pos: kv.key.Pos}
			for _, c := range p.parseSectionMapping("credentials", kv.val, false) {
				switch c.key.Value {
				case "username":
					cred.Username = p.parseString(c.val)
				case "password":
					cred.Password = p.parseString(c.val)
				default:
					p.unexpectedKey(c.key, "credentials", []string{"username", "password"})
				}
			}
			if cred.Username == nil || cred.Password == nil {
				p.errorAt(kv.key.Pos, "both \"username\" and \"password\" must be specified in \"credentials\" section")
				continue
			}
			ret.Credentials = cred
		case "env":
			ret.Env = p.parseEnv(kv.val)
		case "ports":
			ret.Ports = p.parseStringSequence("ports", kv.val)
		case "volumes":
			ret.Ports = p.parseStringSequence("volumes", kv.val)
		case "options":
			ret.Options = p.parseString(kv.val)
		default:
			p.unexpectedKey(kv.key, sec, []string{
				"image",
				"credentials",
				"env",
				"ports",
				"volumes",
				"options",
			})
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idsteps
func (p *parser) parseStep(n *yaml.Node) *Step {
	ret := &Step{Pos: posAt(n)}

	for _, kv := range p.parseMapping("element of \"steps\" section", n, false) {
		switch kv.key.Value {
		case "id":
			ret.ID = p.parseString(kv.val)
		case "if":
			ret.If = p.parseString(kv.val)
		case "name":
			ret.Name = p.parseString(kv.val)
		case "env":
			ret.Env = p.parseEnv(kv.val)
		case "continue-on-error":
			ret.ContinueOnError = p.parseBool(kv.val)
		case "timeout-minutes":
			ret.TimeoutMinutes = p.parseFloat(kv.val)
		case "uses", "with":
			var exec *ExecAction
			if ret.Exec == nil {
				exec = &ExecAction{}
			} else if e, ok := ret.Exec.(*ExecAction); ok {
				exec = e
			} else {
				p.errorfAt(kv.key.Pos, "this step is for running shell command, but contains %q key which is used for running action", kv.key.Value)
				continue
			}
			if kv.key.Value == "uses" {
				exec.Uses = p.parseString(kv.val)
			} else {
				// kv.key.Value == "with"
				for _, with := range p.parseSectionMapping("with", kv.val, false) {
					switch with.key.Value {
					case "entrypoint":
						// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepswithentrypoint
						exec.Entrypoint = p.parseString(with.val)
					case "args":
						// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepswithargs
						exec.Args = p.parseString(with.val)
					default:
						exec.Inputs[with.key.Value] = &Input{with.key, p.parseString(with.val)}
					}
				}
			}
			ret.Exec = exec
		case "run", "working-directory", "shell":
			var exec *ExecRun
			if ret.Exec == nil {
				exec = &ExecRun{}
			} else if e, ok := ret.Exec.(*ExecRun); ok {
				exec = e
			} else {
				p.errorfAt(kv.key.Pos, "this step is for running action, but contains %q key which is used for running shell command", kv.key.Value)
				continue
			}
			switch kv.key.Value {
			case "run":
				exec.Run = p.parseString(kv.val)
			case "working-directory":
				exec.WorkingDirectory = p.parseString(kv.val)
			case "shell":
				exec.Shell = p.parseString(kv.val)
			}
			ret.Exec = exec
		default:
			p.unexpectedKey(kv.key, "step", []string{
				"id",
				"if",
				"name",
				"env",
				"continue-on-error",
				"timeout-minutes",
				"uses",
				"with",
				"run",
				"working-directory",
				"shell",
			})
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idsteps
func (p *parser) parseSteps(n *yaml.Node) []*Step {
	if ok := p.checkSequence("steps", n); !ok {
		return nil
	}

	if len(n.Content) == 0 {
		p.error(n, "sequence in \"steps\" section must not be empty")
	}

	ret := make([]*Step, 0, len(n.Content))

	for _, c := range n.Content {
		if s := p.parseStep(c); s != nil {
			ret = append(ret, s)
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_id
func (p *parser) parseJob(id *String, n *yaml.Node) *Job {
	ret := &Job{ID: id, Pos: id.Pos}

	for _, kv := range p.parseMapping(fmt.Sprintf("%q job", id.Value), n, false) {
		k, v := kv.key, kv.val
		switch k.Value {
		case "name":
			ret.Name = p.parseString(v)
		case "needs":
			if v.Kind == yaml.ScalarNode {
				// needs: job1
				ret.Needs = []*String{p.parseString(v)}
			} else {
				// needs: [job1, job2]
				ret.Needs = p.parseStringSequence("needs", v)
			}
		case "runs-on":
			if v.Kind == yaml.ScalarNode {
				label := p.parseString(v)
				if label.Value == "self-hosted" {
					ret.RunsOn = &SelfHostedRunner{Pos: k.Pos}
				} else {
					ret.RunsOn = &GitHubHostedRunner{label, k.Pos}
				}
			} else {
				s := p.parseStringSequence("runs-on", v)
				if len(s) == 0 || s[0].Value != "self-hosted" {
					p.error(v, "sequence at \"runs-on\" section cannot be empty and must start with \"self-hosted\" string element")
					continue
				}
				ret.RunsOn = &SelfHostedRunner{
					Labels: s[1:], // Omit first "self-hosted" element
					Pos:    k.Pos,
				}
			}
		case "permissions":
			ret.Permissions = p.parsePermissions(k.Pos, v)
		case "environment":
			ret.Environment = p.parseEnvironment(k.Pos, v)
		case "concurrency":
			ret.Concurrency = p.parseConcurrency(k.Pos, v)
		case "outputs":
			ret.Outputs = p.parseOutputs(v)
		case "env":
			ret.Env = p.parseEnv(v)
		case "defaults":
			ret.Defaults = p.parseDefaults(k.Pos, v)
		case "if":
			ret.If = p.parseString(v)
		case "steps":
			ret.Steps = p.parseSteps(v)
		case "timeout-minutes":
			ret.TimeoutMinutes = p.parseFloat(v)
		case "strategy":
			ret.Strategy = p.parseStrategy(k.Pos, v)
		case "continue-on-error":
			ret.ContinueOnError = p.parseBool(v)
		case "container":
			ret.Container = p.parseContainer("container", k.Pos, v)
		case "services":
			services := p.parseSectionMapping("services", v, false)
			ret.Services = make(map[string]*Service, len(services))
			for _, s := range services {
				ret.Services[s.key.Value] = &Service{
					Name:      s.key,
					Contaienr: p.parseContainer("services", s.key.Pos, s.val),
				}
			}
		default:
			p.unexpectedKey(kv.key, "job", []string{
				"name",
				"needs",
				"runs-on",
				"permissions",
				"environment",
				"concurrency",
				"outputs",
				"env",
				"defaults",
				"if",
				"steps",
				"timeout-minutes",
				"strategy",
				"continue-on-error",
				"container",
				"services",
			})
		}
	}

	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobs
func (p *parser) parseJobs(n *yaml.Node) map[string]*Job {
	jobs := p.parseSectionMapping("jobs", n, false)
	ret := make(map[string]*Job, len(jobs))
	for _, kv := range jobs {
		id, job := kv.key, kv.val
		ret[id.Value] = p.parseJob(id, job)
	}
	return ret
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
func (p *parser) parse(n *yaml.Node) *Workflow {
	w := &Workflow{}

	for _, kv := range p.parseMapping("workflow", n, false) {
		k, v := kv.key, kv.val
		switch k.Value {
		case "name":
			w.Name = p.parseString(v)
		case "on":
			w.On = p.parseEvents(v)
		case "permissions":
			w.Permissions = p.parsePermissions(k.Pos, v)
		case "env":
			w.Env = p.parseEnv(v)
		case "defaults":
			w.Defaults = p.parseDefaults(k.Pos, v)
		case "concurrency":
			w.Concurrency = p.parseConcurrency(k.Pos, v)
		case "jobs":
			w.Jobs = p.parseJobs(v)
		default:
			p.unexpectedKey(k, "workflow", []string{"name", "on", "permissions", "env", "defaults", "concurrency", "jobs"})
		}
	}

	return w
}
