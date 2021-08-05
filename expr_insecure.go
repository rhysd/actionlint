package actionlint

import (
	"strings"
)

type untrustedInputMap map[string]untrustedInputMap

func (m untrustedInputMap) findChild(name string) (untrustedInputMap, bool) {
	if m == nil {
		return nil, false
	}
	if c, ok := m[name]; ok {
		return c, true
	}
	if c, ok := m["*"]; ok {
		return c, true
	}
	return nil, false
}

// BuiltinUntrustedInputs is list of untrusted inputs. These inputs are detected as untrusted in
// `run:` scripts. See the URL for more details.
// https://securitylab.github.com/research/github-actions-untrusted-input/
var BuiltinUntrustedInputs = untrustedInputMap{
	"github": {
		"event": {
			"issue": {
				"title": nil,
				"body":  nil,
			},
			"pull_request": {
				"title": nil,
				"body":  nil,
				"head": {
					"ref":   nil,
					"label": nil,
					"repo": {
						"default_branch": nil,
					},
				},
			},
			"comment": {
				"body": nil,
			},
			"review": {
				"body": nil,
			},
			"pages": {
				"*": {
					"page_name": nil,
				},
			},
			"commits": {
				"*": {
					"message": nil,
					"author": {
						"email": nil,
						"name":  nil,
					},
				},
			},
			"head_commit": {
				"message": nil,
				"author": {
					"email": nil,
					"name":  nil,
				},
			},
		},
		"head_ref": nil,
	},
}

// UntrustedInputChecker is a checker to detect untrusted inputs in an expression syntax tree.
type UntrustedInputChecker struct {
	root    untrustedInputMap
	cur     untrustedInputMap
	varNode *VariableNode
	path    []string
	errs    []*ExprError
}

// NewUntrustedInputChecker creates new UntrustedInputChecker instance.
func NewUntrustedInputChecker(m untrustedInputMap) *UntrustedInputChecker {
	return &UntrustedInputChecker{
		root:    m,
		cur:     nil,
		varNode: nil,
		path:    nil,
	}
}

func (u *UntrustedInputChecker) found() {
	err := errorfAtExpr(u.varNode, "%q is possibly untrusted. please avoid using it directly in script by passing it through environment variable. see https://securitylab.github.com/research/github-actions-untrusted-input for more details.", strings.Join(u.path, "."))
	u.errs = append(u.errs, err)
	u.done()
}

func (u *UntrustedInputChecker) done() {
	if u.cur == nil {
		return
	}
	u.cur = nil
	u.varNode = nil
	u.path = u.path[:0]
}

func (u *UntrustedInputChecker) checkVar(name string) {
	m, ok := u.root[name]
	if !ok {
		u.done()
		return
	}

	u.path = append(u.path, name)

	if m == nil {
		u.found()
		return
	}

	u.cur = m
}

func (u *UntrustedInputChecker) checkProp(name string) {
	c, ok := u.cur.findChild(name)
	if !ok {
		u.done()
		return
	}

	u.path = append(u.path, name)

	if c == nil {
		u.found()
		return
	}

	u.cur = c
}

// OnNode is a callback which should be called on visiting node. This method assumes to be called
// in depth-first, bottom-up order.
func (u *UntrustedInputChecker) OnNode(n ExprNode) {
	switch n := n.(type) {
	case *VariableNode:
		u.varNode = n
		u.checkVar(n.Name)
	case *ObjectDerefNode:
		u.checkProp(n.Property)
	default:
		u.done()
	}
}

// Errs returns errors detected by this checker. This method should be called after visiting all
// nodes in a syntax tree.
func (u *UntrustedInputChecker) Errs() []*ExprError {
	return u.errs
}
