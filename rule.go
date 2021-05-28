package actionlint

// RuleBase is a struct to be a base of rule structs. Embed this struct to define default methods
// automatically
type RuleBase struct {
	errs []*Error
}

// VisitStep is callback when visiting Step node
func (r *RuleBase) VisitStep(node *Step) {}

// VisitJobPre is callback when visiting Job node before visiting its children
func (r *RuleBase) VisitJobPre(node *Job) {}

// VisitJobPost is callback when visiting Job node after visiting its children
func (r *RuleBase) VisitJobPost(node *Job) {}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children
func (r *RuleBase) VisitWorkflowPre(node *Workflow) {}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children
func (r *RuleBase) VisitWorkflowPost(node *Workflow) {}

func (r *RuleBase) error(pos *Pos, msg string) {
	err := errorAt(pos, msg)
	r.errs = append(r.errs, err)
}

func (r *RuleBase) errorf(pos *Pos, format string, args ...interface{}) {
	err := errorfAt(pos, format, args...)
	r.errs = append(r.errs, err)
}

// Errs returns errors found by the rule.
func (r *RuleBase) Errs() []*Error {
	return r.errs
}

// Rule is an interface which all rule structs must meet
type Rule interface {
	Pass
	Errs() []*Error
}
