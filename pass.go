package actionlint

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
}

// NewVisitor creates Visitor instance
func NewVisitor() *Visitor {
	return &Visitor{}
}

// AddPass adds new pass which is called on traversing a syntax tree
func (v *Visitor) AddPass(p Pass) {
	v.passes = append(v.passes, p)
}

// Visit visits given syntax tree in depth-first order
func (v *Visitor) Visit(n *Workflow) {
	for _, p := range v.passes {
		p.VisitWorkflowPre(n)
	}

	for _, j := range n.Jobs {
		v.visitJob(j)
	}

	for _, p := range v.passes {
		p.VisitWorkflowPost(n)
	}
}

func (v *Visitor) visitJob(n *Job) {
	for _, p := range v.passes {
		p.VisitJobPre(n)
	}

	for _, s := range n.Steps {
		v.visitStep(s)
	}

	for _, p := range v.passes {
		p.VisitJobPost(n)
	}
}

func (v *Visitor) visitStep(n *Step) {
	for _, p := range v.passes {
		p.VisitStep(n)
	}
}
