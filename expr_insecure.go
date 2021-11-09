package actionlint

import (
	"fmt"
	"strings"
)

// UntrustedInputMap2 is a recursive map to match context object property dereferences.
// Root of this map represents each context names and their ancestors represent recursive properties.
type UntrustedInputMap2 struct {
	Name     string
	Parent   *UntrustedInputMap2
	Children map[string]*UntrustedInputMap2
}

func (m *UntrustedInputMap2) String() string {
	if m.Children == nil {
		return fmt.Sprintf("{%q}", m.Name)
	}
	return fmt.Sprintf("{%q: %v}", m.Name, m.Children)
}

// Find child object property in this map
func (m *UntrustedInputMap2) findObjectProp(name string) (*UntrustedInputMap2, bool) {
	if m != nil && m.Children != nil {
		if c, ok := m.Children[name]; ok {
			return c, true
		}
	}
	return nil, false
}

// Find child array element in this map. This is special case with object filter where its receiver is an array
func (m *UntrustedInputMap2) findArrayElem() (*UntrustedInputMap2, bool) {
	return m.findObjectProp("*")
}

func (m *UntrustedInputMap2) buildPath(b *strings.Builder) {
	if m.Parent != nil && m.Parent.Name != "" {
		m.Parent.buildPath(b)
		b.WriteRune('.')
	}
	b.WriteString(m.Name)
}

func NewUntrustedInputMap2(name string, children ...*UntrustedInputMap2) *UntrustedInputMap2 {
	m := &UntrustedInputMap2{
		Name:     name,
		Parent:   nil,
		Children: nil, // Leaf of the tree is nil
	}
	if len(children) > 0 {
		m.Children = make(map[string]*UntrustedInputMap2, len(children))
		for _, c := range children {
			c.Parent = m
			m.Children[c.Name] = c
		}
	}
	return m
}

type UntrustedInputSearchRoots map[string]*UntrustedInputMap2

func (ms UntrustedInputSearchRoots) AddRoot(m *UntrustedInputMap2) {
	ms[m.Name] = m
}

// TODO: Automatically generate BuitinUntrustedInputs from https://github.com/github/codeql/blob/main/javascript/ql/src/experimental/Security/CWE-094/ExpressionInjection.ql

// BuiltinUntrustedInputs2 is list of untrusted inputs. These inputs are detected as untrusted in
// `run:` scripts. See the URL for more details.
// - https://securitylab.github.com/research/github-actions-untrusted-input/
// - https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions
// - https://github.com/github/codeql/blob/main/javascript/ql/src/experimental/Security/CWE-094/ExpressionInjection.ql
var BuiltinUntrustedInputs2 = UntrustedInputSearchRoots{
	"github": NewUntrustedInputMap2("github",
		NewUntrustedInputMap2("event",
			NewUntrustedInputMap2("issue",
				NewUntrustedInputMap2("title"),
				NewUntrustedInputMap2("body"),
			),
			NewUntrustedInputMap2("pull_request",
				NewUntrustedInputMap2("title"),
				NewUntrustedInputMap2("body"),
				NewUntrustedInputMap2("head",
					NewUntrustedInputMap2("ref"),
					NewUntrustedInputMap2("label"),
					NewUntrustedInputMap2("repo",
						NewUntrustedInputMap2("default_branch"),
					),
				),
			),
			NewUntrustedInputMap2("comment",
				NewUntrustedInputMap2("body"),
			),
			NewUntrustedInputMap2("review",
				NewUntrustedInputMap2("body"),
			),
			NewUntrustedInputMap2("review_comment",
				NewUntrustedInputMap2("body"),
			),
			NewUntrustedInputMap2("pages",
				NewUntrustedInputMap2("*",
					NewUntrustedInputMap2("page_name"),
				),
			),
			NewUntrustedInputMap2("commits",
				NewUntrustedInputMap2("*",
					NewUntrustedInputMap2("message"),
					NewUntrustedInputMap2("author",
						NewUntrustedInputMap2("email"),
						NewUntrustedInputMap2("name"),
					),
				),
			),
			NewUntrustedInputMap2("head_commit",
				NewUntrustedInputMap2("message"),
				NewUntrustedInputMap2("author",
					NewUntrustedInputMap2("email"),
					NewUntrustedInputMap2("name"),
				),
			),
			NewUntrustedInputMap2("discussion",
				NewUntrustedInputMap2("title"),
				NewUntrustedInputMap2("body"),
			),
		),
		NewUntrustedInputMap2("head_ref"),
	),
}

type UntrustedInputChecker2 struct {
	roots           UntrustedInputSearchRoots
	filteringObject bool
	cur             map[*UntrustedInputMap2]struct{}
	start           ExprNode
	errs            []*ExprError
}

func NewUntrustedInputChecker2(roots UntrustedInputSearchRoots) *UntrustedInputChecker2 {
	return &UntrustedInputChecker2{
		roots:           roots,
		filteringObject: false,
		cur:             make(map[*UntrustedInputMap2]struct{}),
		start:           nil,
		errs:            []*ExprError{},
	}
}

func (u *UntrustedInputChecker2) reset() {
	u.start = nil
	u.filteringObject = false
	for k := range u.cur {
		delete(u.cur, k)
	}
}

func (u *UntrustedInputChecker2) onVar(v *VariableNode) {
	c, ok := u.roots[v.Name] // Find root context (currently only "github" exists)
	if !ok {
		return
	}
	u.start = v
	u.cur[c] = struct{}{}
}

func (u *UntrustedInputChecker2) onPropAccess(name string) {
	for cur := range u.cur {
		delete(u.cur, cur)

		c, ok := cur.findObjectProp(name)
		if !ok {
			continue
		}

		u.cur[c] = struct{}{} // depth + 1
	}
}

func (u *UntrustedInputChecker2) onIndexAccess() {
	if u.filteringObject {
		u.filteringObject = false
		return // For example, match `github.event.*.body[0]` as `github.event.commits[0].body`
	}
	for cur := range u.cur {
		delete(u.cur, cur)
		if c, ok := cur.findArrayElem(); ok {
			u.cur[c] = struct{}{}
		}
	}
}

func (u *UntrustedInputChecker2) onObjectFilter() {
	u.filteringObject = true

	// Do not iterate elements which are added in the loop.
	// Order of map element iterations is random, but an element newly created while loop iteration
	// is always after iterating all elements existing before the loop (if it is not skipped).
	count := len(u.cur)

	for cur := range u.cur {
		if count == 0 {
			return
		}
		delete(u.cur, cur)

		// Object filter for arrays
		if c, ok := cur.findArrayElem(); ok {
			u.cur[c] = struct{}{}
			continue
		}

		// Object filter for objects
		for _, c := range cur.Children {
			u.cur[c] = struct{}{}
		}

		count--
	}
}

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

func (u *UntrustedInputChecker2) end() {
	for cur := range u.cur {
		if cur.Children != nil {
			continue // When `Children` is nil, the node is a leaf
		}
		var b strings.Builder
		b.WriteRune('"')
		cur.buildPath(&b)
		b.WriteString(`" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions for more details`)
		err := errorAtExpr(u.start, b.String())
		u.errs = append(u.errs, err)
		break // Get only first path
	}
	u.reset()
}

// OnVisitNodeLeave is a callback which should be called on visiting node after visiting its children.
func (u *UntrustedInputChecker2) OnVisitNodeLeave(n ExprNode) {
	switch n := n.(type) {
	case *VariableNode:
		u.end()
		u.onVar(n)
	case *ObjectDerefNode:
		u.onPropAccess(n.Property)
	case *IndexAccessNode:
		if lit, ok := n.Index.(*StringNode); ok {
			// Special case like github['event']['issue']['title']
			u.onPropAccess(lit.Value)
			break
		}
		u.onIndexAccess()
	case *ArrayDerefNode:
		u.onObjectFilter()
	default:
		u.end()
	}
}

// OnVisitEnd is a callback which should be called after visiting whole syntax tree. This callback
// is necessary to handle the case where an untrusted input is at root of expression
func (u *UntrustedInputChecker2) OnVisitEnd() {
	u.end()
}

// Errs returns errors detected by this checker. This method should be called after visiting all
// nodes in a syntax tree.
func (u *UntrustedInputChecker2) Errs() []*ExprError {
	return u.errs
}

// Init initializes a state of checker.
func (u *UntrustedInputChecker2) Init() {
	u.errs = []*ExprError{}
	u.reset()
}

// TODO: Automatically generate BuitinUntrustedInputs from https://github.com/github/codeql/blob/main/javascript/ql/src/experimental/Security/CWE-094/ExpressionInjection.ql

// BuiltinUntrustedInputs is list of untrusted inputs. These inputs are detected as untrusted in
// `run:` scripts. See the URL for more details.
// - https://securitylab.github.com/research/github-actions-untrusted-input/
// - https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions
// - https://github.com/github/codeql/blob/main/javascript/ql/src/experimental/Security/CWE-094/ExpressionInjection.ql
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
			"discussion": {
				"title": nil,
				"body":  nil,
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
	b.WriteString(`" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions for more details`)
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
