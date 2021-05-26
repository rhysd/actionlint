package syntax

import (
	"fmt"
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

func pos(n *yaml.Node) *Pos {
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
	return &String{n.Value, pos(n)}
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
	return &Bool{n.Value == "true", pos(n)}
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
		p.errorf(n, "%s should not be empty", what)
	}

	return m
}

func (p *parser) parseSectionMapping(sec string, n *yaml.Node, allowEmpty bool) []keyVal {
	return p.parseMapping(fmt.Sprintf("%q section", sec), n, allowEmpty)
}

func (p *parser) parseScheduleEvent(n *yaml.Node) *ScheduledEvent {
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

	return &ScheduledEvent{cron, pos(n)}
}

// https://docs.github.com/en/actions/reference/events-that-trigger-workflows#workflow_dispatch
func (p *parser) parseWorkflowDispatchEvent(n *yaml.Node) *WorkflowDispatchEvent {
	ret := &WorkflowDispatchEvent{Pos: pos(n)}

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
func (p *parser) parseRepositoryDispatchEvent(n *yaml.Node) *RepositoryDispatchEvent {
	ret := &RepositoryDispatchEvent{Pos: pos(n)}

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
	ret := &WebhookEvent{Hook: name, Pos: pos(n)}

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
				if e := p.parseScheduleEvent(kv.val); e != nil {
					ret = append(ret, e)
				}
			case "workflow_dispatch":
				ret = append(ret, p.parseWorkflowDispatchEvent(kv.val))
			case "repository_dispatch":
				ret = append(ret, p.parseRepositoryDispatchEvent(kv.val))
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
					ret = append(ret, &WebhookEvent{Hook: s, Pos: pos(c)})
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
func (p *parser) parsePermissions(n *yaml.Node) *Permissions {
	ret := &Permissions{Pos: pos(n)}

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
		ret.All = &Permission{nil, kind, pos(n)}
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
func (p *parser) parseDefaults(n *yaml.Node) *Defaults {
	ret := &Defaults{Pos: pos(n)}

	for _, kv := range p.parseSectionMapping("defaults", n, false) {
		if kv.key.Value != "run" {
			p.unexpectedKey(kv.key, "defaults", []string{"run"})
			continue
		}
		ret.Run = &DefaultsRun{Pos: pos(kv.val)}

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

func (p *parser) parseConcurrency(n *yaml.Node) *Concurrency {
	ret := &Concurrency{Pos: pos(n)}

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
			w.Permissions = p.parsePermissions(v)
		case "env":
			w.Env = p.parseEnv(v)
		case "defaults":
			w.Defaults = p.parseDefaults(v)
		case "concurrency":
			w.Concurrency = p.parseConcurrency(v)
		case "jobs":
			panic("TODO")
		default:
			p.unexpectedKey(k, "workflow", []string{"name", "on", "permissions", "env", "defaults", "concurrency", "jobs"})
		}
	}

	return w
}
