package actionlint

import (
	"fmt"
	"io"
)

// RuleBase is a struct to be a base of rule structs. Embed this struct to define default methods
// automatically
type RuleBase struct {
	name   string
	desc   string
	errs   []*Error
	dbg    io.Writer
	config *Config
}

// NewRuleBase creates a new RuleBase instance. It should be embedded to your own
// rule instance.
func NewRuleBase(name string, desc string) RuleBase {
	return RuleBase{
		name: name,
		desc: desc,
	}
}

// VisitStep is callback when visiting Step node
func (r *RuleBase) VisitStep(node *Step) error { return nil }

// VisitJobPre is callback when visiting Job node before visiting its children.
func (r *RuleBase) VisitJobPre(node *Job) error { return nil }

// VisitJobPost is callback when visiting Job node after visiting its children.
func (r *RuleBase) VisitJobPost(node *Job) error { return nil }

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (r *RuleBase) VisitWorkflowPre(node *Workflow) error { return nil }

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (r *RuleBase) VisitWorkflowPost(node *Workflow) error { return nil }

// VisitActionPre is callback when visiting Workflow node before visiting its children.
func (r *RuleBase) VisitActionPre(node *Action) error { return nil }

// VisitActionPost is callback when visiting Workflow node after visiting its children.
func (r *RuleBase) VisitActionPost(node *Action) error { return nil }

// Error creates a new error from the source position and the error message and stores it in the
// rule instance. The errors can be accessed by Errs method.
func (r *RuleBase) Error(pos *Pos, msg string) {
	err := errorAt(pos, r.name, msg)
	r.errs = append(r.errs, err)
}

// Errorf reports a new error with the source position and the formatted error message and stores it
// in the rule instance. The errors can be accessed by Errs method.
func (r *RuleBase) Errorf(pos *Pos, format string, args ...interface{}) {
	err := errorfAt(pos, r.name, format, args...)
	r.errs = append(r.errs, err)
}

// Debug prints debug log to the output. The output is specified by the argument of EnableDebug method.
// By default, no output is set so debug log is not printed.
func (r *RuleBase) Debug(format string, args ...interface{}) {
	if r.dbg == nil {
		return
	}
	format = fmt.Sprintf("[%s] %s\n", r.name, format)
	fmt.Fprintf(r.dbg, format, args...)
}

// Errs returns errors found by the rule.
func (r *RuleBase) Errs() []*Error {
	return r.errs
}

// Name returns the name of the rule.
func (r *RuleBase) Name() string {
	return r.name
}

// Description returns the description of the rule.
func (r *RuleBase) Description() string {
	return r.desc
}

// EnableDebug enables debug output from the rule. Given io.Writer instance is used to print debug
// information to console. Setting nil means disabling debug output.
func (r *RuleBase) EnableDebug(out io.Writer) {
	r.dbg = out
}

// SetConfig populates user configuration of actionlint to the rule. When no config is set, rules
// should behave as if the default configuration is set.
func (r *RuleBase) SetConfig(cfg *Config) {
	r.config = cfg
}

// Rule is an interface which all rule structs must meet.
type Rule interface {
	Pass
	Errs() []*Error
	Name() string
	Description() string
	EnableDebug(out io.Writer)
	SetConfig(cfg *Config)
}
