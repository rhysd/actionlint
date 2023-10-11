package actionlint

import (
	"fmt"
	"io"
	"time"
)

// Pass is an interface to traverse a workflow syntax tree
type Pass interface {
	// VisitStep is callback when visiting Step node. It returns internal error when it cannot continue the process
	VisitStep(node *Step) error
	// VisitJobPre is callback when visiting Job node before visiting its children. It returns internal error when it cannot continue the process
	VisitJobPre(node *Job) error
	// VisitJobPost is callback when visiting Job node after visiting its children. It returns internal error when it cannot continue the process
	VisitJobPost(node *Job) error
	// VisitWorkflowPre is callback when visiting Workflow node before visiting its children. It returns internal error when it cannot continue the process
	VisitWorkflowPre(node *Workflow) error
	// VisitWorkflowPost is callback when visiting Workflow node after visiting its children. It returns internal error when it cannot continue the process
	VisitWorkflowPost(node *Workflow) error

	// VisitActionPre is callback when visiting Action node before visiting its children. It returns internal error when it cannot continue the process
	VisitActionPre(node *Action) error
	// VisitActionPost is callback when visiting Action node after visiting its children. It returns internal error when it cannot continue the process
	VisitActionPost(node *Action) error
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
func (v *Visitor) Visit(n *Workflow) error {
	var t time.Time
	if v.dbg != nil {
		t = time.Now()
	}

	for _, p := range v.passes {
		if err := p.VisitWorkflowPre(n); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime("VisitWorkflowPre", t)
		t = time.Now()
	}

	for _, j := range n.Jobs {
		if err := v.visitJob(j); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("Visiting %d jobs", len(n.Jobs)), t)
		t = time.Now()
	}

	for _, p := range v.passes {
		if err := p.VisitWorkflowPost(n); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime("VisitWorkflowPost", t)
	}

	return nil
}

// VisitAction visits given syntax tree in depth-first order
func (v *Visitor) VisitAction(n *Action) error {
	var t time.Time
	if v.dbg != nil {
		t = time.Now()
	}

	// dw := &Workflow{}
	// dw.Name = &String{Value: "dummy-workflow", Quoted: false, Pos: &Pos{Line: 1, Col: 3}}
	// de := WorkflowDispatchEvent{Inputs: make(map[string]*DispatchInput)}
	// for k, v := range n.Common().Inputs {
	// 	de.Inputs[k] = &DispatchInput{Name: &String{Value: k}, Description: v.Description, Required: v.Required, Default: v.Default}
	// }
	// dw.On = append(dw.On, &de)
	// dj := &Job{}
	// dj.ID = &String{Value: "action", Quoted: false, Pos: &Pos{Line: 1, Col: 3}}
	// if n.Kind() == ActionKindComposite {
	// 	dj.Outputs = make(map[string]*Output)
	// 	for k, v := range n.Common().Outputs {
	// 		dj.Outputs[k] = &Output{Name: &String{Value: k}, Value: v.Value}
	// 	}
	// }
	// dw.Jobs = make(map[string]*Job)
	// dw.Jobs[dj.ID.Value] = dj

	// for _, p := range v.passes {
	// 	if err := p.VisitWorkflowPre(dw); err != nil {
	// 		return err
	// 	}
	// }

	// if v.dbg != nil {
	// 	v.reportElapsedTime("VisitWorkflowPre", t)
	// 	t = time.Now()
	// }

	// for _, p := range v.passes {
	// 	if err := p.VisitJobPre(dj); err != nil {
	// 		return err
	// 	}
	// }

	for _, p := range v.passes {
		if err := p.VisitActionPre(n); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime("VisitActionPre", t)
		t = time.Now()
	}

	if c, ok := n.Runs.(*CompositeAction); ok {
		for _, s := range c.Steps {
			if err := v.visitStep(s); err != nil {
				return err
			}
		}

		if v.dbg != nil {
			// v.reportElapsedTime(fmt.Sprintf("Visiting %d steps at job %q", len(n.Steps), n.ID.Value), t)
			v.reportElapsedTime(fmt.Sprintf("Visiting %d steps of action %q", len(c.Steps), n.Name.Value), t)
		}
	}

	// for _, p := range v.passes {
	// 	if err := p.VisitJobPost(dj); err != nil {
	// 		return err
	// 	}
	// }

	for _, p := range v.passes {
		//	if err := p.VisitWorkflowPost(dw); err != nil {
		if err := p.VisitActionPost(n); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime("VisitActionPost", t)
	}

	return nil
}

func (v *Visitor) visitJob(n *Job) error {
	var t time.Time
	if v.dbg != nil {
		t = time.Now()
	}

	for _, p := range v.passes {
		if err := p.VisitJobPre(n); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("VisitWorkflowJobPre at job %q", n.ID.Value), t)
		t = time.Now()
	}

	for _, s := range n.Steps {
		if err := v.visitStep(s); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("Visiting %d steps at job %q", len(n.Steps), n.ID.Value), t)
		t = time.Now()
	}

	for _, p := range v.passes {
		if err := p.VisitJobPost(n); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("VisitWorkflowJobPost at job %q", n.ID.Value), t)
	}

	return nil
}

func (v *Visitor) visitStep(n *Step) error {
	var t time.Time
	if v.dbg != nil {
		t = time.Now()
	}

	for _, p := range v.passes {
		if err := p.VisitStep(n); err != nil {
			return err
		}
	}

	if v.dbg != nil {
		v.reportElapsedTime(fmt.Sprintf("VisitStep at %s", n.Pos), t)
	}

	return nil
}
