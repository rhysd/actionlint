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
	root untrustedInputMap
	cur  untrustedInputMap
	errs []*ExprError
}

// NewUntrustedInputChecker creates new UntrustedInputChecker instance.
func NewUntrustedInputChecker(m untrustedInputMap) *UntrustedInputChecker {
	return &UntrustedInputChecker{
		root: m,
		cur:  nil,
	}
}

func (u *UntrustedInputChecker) found(v *VariableNode, path string) {
	err := errorfAtExpr(v, "%q is possibly untrusted. please avoid using it directly in script by passing it through environment variable. see https://securitylab.github.com/research/github-actions-untrusted-input for more details.", path)
	u.errs = append(u.errs, err)
	u.done()
}

func (u *UntrustedInputChecker) done() {
	if u.cur == nil {
		return
	}
	u.cur = nil
}

func (u *UntrustedInputChecker) checkVar(name string) bool {
	m, ok := u.root[name]
	if !ok {
		u.done()
		return true
	}

	if m == nil {
		return false
	}

	u.cur = m
	return true
}

func (u *UntrustedInputChecker) checkProp(name string) bool {
	c, ok := u.cur.findChild(name)
	if !ok {
		u.done()
		return true
	}

	if c == nil {
		return false
	}

	u.cur = c
	return true
}

func buildPathOfObjectDeref(b *strings.Builder, n *ObjectDerefNode) *VariableNode {
	var v *VariableNode
	switch n := n.Receiver.(type) {
	case *VariableNode:
		b.WriteString(n.Name)
		v = n
	case *ObjectDerefNode:
		v = buildPathOfObjectDeref(b, n)
	default:
		panic("unreachable")
	}
	b.WriteByte('.')
	b.WriteString(n.Property)
	return v
}

// OnNode is a callback which should be called on visiting node. This method assumes to be called
// in depth-first, bottom-up order.
func (u *UntrustedInputChecker) OnNode(n ExprNode) {
	switch n := n.(type) {
	case *VariableNode:
		if !u.checkVar(n.Name) {
			u.found(n, n.Name)
		}
	case *ObjectDerefNode:
		if !u.checkProp(n.Property) {
			var b strings.Builder
			v := buildPathOfObjectDeref(&b, n)
			u.found(v, b.String())
		}
	default:
		u.done()
	}
}

// Errs returns errors detected by this checker. This method should be called after visiting all
// nodes in a syntax tree.
func (u *UntrustedInputChecker) Errs() []*ExprError {
	return u.errs
}
