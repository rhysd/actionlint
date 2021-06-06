package actionlint

import (
	"fmt"
	"io"
	"time"
)

// Pass is an interface to traverse a workflow syntax tree
type Pass interface {
	// VisitStep is callback when visiting Step node
	VisitStep(node *Step)
	// VisitJobPre is callback when visiting Job node before visiting its children
	VisitJobPre(node *Job)
	// VisitJobPost is callback when visiting Job node after visiting its children
	VisitJobPost(node *Job)
	// VisitWorkflowPre is callback when visiting Workflow node before visiting its children
	VisitWorkflowPre(node *Workflow)
	// VisitWorkflowPost is callback when visiting Workflow node after visiting its children
	VisitWorkflowPost(node *Workflow)
}

// Visitor visits syntax tree from root in depth-first order
type Visitor struct {
	passes []Pass
	dbg    io.Writer
}

// NewVisitor creates Visitor instance
func NewVisitor() *Visitor {
	return &Visitor{}
}

// AddPass adds new pass which is called on traversing a syntax tree
func (v *Visitor) AddPass(p Pass) {
	v.passes = append(v.passes, p)
}

// EnableDebug enables debug output when non-nil io.Writer value is given. All debug outputs from
// visitor will be written to the writer.
func (v *Visitor) EnableDebug(w io.Writer) {
	v.dbg = w
}

func (v *Visitor) reportElapsedTime(what string, start time.Time) {
	fmt.Fprintf(v.dbg, "[Visitor] %s took %vms\n", what, time.Since(start).Milliseconds())
}

// Visit visits given syntax tree in depth-first order
func (v *Visitor) Visit(n *Workflow) {
	var t time.Time
	if v.dbg != nil {
		t = time.Now()
	}

	for _, p := range v.passes {
		p.VisitWorkflowPre(n)
	}

	if v.dbg != nil {
		v.reportElapsedTime("VisitWorkflowPre", t)
		t = time.Now()
	}

	for _, j := range n.Jobs {
		v.visitJob(j)
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("Visiting %d jobs", len(n.Jobs)), t)
		t = time.Now()
	}

	for _, p := range v.passes {
		p.VisitWorkflowPost(n)
	}

	if v.dbg != nil {
		v.reportElapsedTime("VisitWorkflowPost", t)
	}
}

func (v *Visitor) visitJob(n *Job) {
	var t time.Time
	if v.dbg != nil {
		t = time.Now()
	}

	for _, p := range v.passes {
		p.VisitJobPre(n)
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("VisitWorkflowJobPre at job %q", n.ID.Value), t)
		t = time.Now()
	}

	for _, s := range n.Steps {
		v.visitStep(s)
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("Visiting %d steps at job %q", len(n.Steps), n.ID.Value), t)
		t = time.Now()
	}

	for _, p := range v.passes {
		p.VisitJobPost(n)
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("VisitWorkflowJobPost at job %q", n.ID.Value), t)
	}
}

func (v *Visitor) visitStep(n *Step) {
	var t time.Time
	if v.dbg != nil {
		t = time.Now()
	}

	for _, p := range v.passes {
		p.VisitStep(n)
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("VisitStep at %s", n.Pos), t)
	}
}
