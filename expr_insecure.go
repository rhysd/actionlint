package actionlint

import (
	"strings"
)

// UntrustedInputMap is a recursive map to match context object property dereferences.
// Root of this map represents each context names and their ancestors represent recursive properties.
type UntrustedInputMap map[string]UntrustedInputMap

func (m UntrustedInputMap) findPropChild(name string) (UntrustedInputMap, bool) {
	if m != nil {
		if c, ok := m[name]; ok {
			return c, true
		}
	}
	return nil, false
}

func (m UntrustedInputMap) findElemChild() (UntrustedInputMap, bool) {
	if m != nil {
		if c, ok := m["*"]; ok {
			return c, true
		}
	}
	return nil, false
}

// BuiltinUntrustedInputs is list of untrusted inputs. These inputs are detected as untrusted in
// `run:` scripts. See the URL for more details.
// - https://securitylab.github.com/research/github-actions-untrusted-input/
// - https://docs.github.com/en/actions/learn-github-actions/security-hardening-for-github-actions
var BuiltinUntrustedInputs = UntrustedInputMap{
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
			"review_comment": {
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
// Note: To avoid breaking the state of checking property accesses on nested property accesses like
// foo[aaa.bbb].bar, IndexAccessNode.Index must be visited before IndexAccessNode.Operand.
type UntrustedInputChecker struct {
	root UntrustedInputMap
	cur  UntrustedInputMap
	errs []*ExprError
}

// NewUntrustedInputChecker creates new UntrustedInputChecker instance.
func NewUntrustedInputChecker(m UntrustedInputMap) *UntrustedInputChecker {
	return &UntrustedInputChecker{
		root: m,
		errs: []*ExprError{},
	}
}

func (u *UntrustedInputChecker) done() {
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
	c, ok := u.cur.findPropChild(name)
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

func (u *UntrustedInputChecker) checkElem() bool {
	c, ok := u.cur.findElemChild()
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

func buildPathOfObjectDereference(b *strings.Builder, n ExprNode) *VariableNode {
	switch n := n.(type) {
	case *VariableNode:
		b.WriteString(n.Name)
		return n
	case *ObjectDerefNode:
		v := buildPathOfObjectDereference(b, n.Receiver)
		b.WriteByte('.')
		b.WriteString(n.Property)
		return v
	case *IndexAccessNode:
		v := buildPathOfObjectDereference(b, n.Operand)
		if lit, ok := n.Index.(*StringNode); ok {
			b.WriteByte('.')
			b.WriteString(lit.Value)
			return v
		}
		b.WriteString(".*")
		return v
	case *ArrayDerefNode:
		v := buildPathOfObjectDereference(b, n.Receiver)
		b.WriteString(".*")
		return v
	}
	panic("unreachable")
}

func (u *UntrustedInputChecker) error(n ExprNode) {
	var b strings.Builder
	b.WriteByte('"')
	v := buildPathOfObjectDereference(&b, n)
	b.WriteString(`" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/learn-github-actions/security-hardening-for-github-actions for more details`)
	err := errorAtExpr(v, b.String())
	u.errs = append(u.errs, err)
	u.done()
}

// OnNodeLeave is a callback which should be called on visiting node after visiting its children.
func (u *UntrustedInputChecker) OnNodeLeave(n ExprNode) {
	switch n := n.(type) {
	case *VariableNode:
		if !u.checkVar(n.Name) {
			u.error(n)
		}
	case *ObjectDerefNode:
		if !u.checkProp(n.Property) {
			u.error(n)
		}
	case *IndexAccessNode:
		if lit, ok := n.Index.(*StringNode); ok {
			// Special case like github['event']['issue']['title']
			if !u.checkProp(lit.Value) {
				u.error(n)
			}
			break
		}
		if !u.checkElem() {
			u.error(n)
		}
	case *ArrayDerefNode:
		if !u.checkElem() {
			u.error(n)
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

// Init initializes a state of checker.
func (u *UntrustedInputChecker) Init() {
	u.errs = []*ExprError{}
	u.done()
}
