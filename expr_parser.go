package actionlint

import (
	"fmt"
	"strconv"
	"strings"
)

// ExprNode is a node of expression syntax tree. To know the syntax, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type ExprNode interface {
	// Token returns the first token of the node. This method is useful to get position of this node.
	Token() *Token
}

// Variable

// VariableNode is node for variable access.
type VariableNode struct {
	// Name is name of the variable
	Name string
	tok  *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *VariableNode) Token() *Token {
	return n.tok
}

// Literals

// NullNode is node for null literal.
type NullNode struct {
	tok *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *NullNode) Token() *Token {
	return n.tok
}

// BoolNode is node for boolean literal, true or false.
type BoolNode struct {
	// Value is value of the boolean literal.
	Value bool
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *BoolNode) Token() *Token {
	return n.tok
}

// IntNode is node for integer literal.
type IntNode struct {
	// Value is value of the integer literal.
	Value int
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *IntNode) Token() *Token {
	return n.tok
}

// FloatNode is node for float literal.
type FloatNode struct {
	// Value is value of the float literal.
	Value float64
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *FloatNode) Token() *Token {
	return n.tok
}

// StringNode is node for string literal.
type StringNode struct {
	// Value is value of the string literal. Escapes are resolved and quotes at both edges are
	// removed.
	Value string
	tok   *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *StringNode) Token() *Token {
	return n.tok
}

// Operators

// ObjectDerefNode represents property dereference of object like 'foo.bar'.
type ObjectDerefNode struct {
	// Receiver is an expression at receiver of property dereference.
	Receiver ExprNode
	// Property is a name of property to access.
	Property string
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n ObjectDerefNode) Token() *Token {
	return n.Receiver.Token()
}

// ArrayDerefNode represents elements dereference of arrays like '*' in 'foo.bar.*.piyo'.
type ArrayDerefNode struct {
	// Receiver is an expression at receiver of array element dereference.
	Receiver ExprNode
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n ArrayDerefNode) Token() *Token {
	return n.Receiver.Token()
}

// IndexAccessNode is node for index access, which represents dynamic object property access or
// array index access.
type IndexAccessNode struct {
	// Operand is an expression at operand of index access, which should be array or object.
	Operand ExprNode
	// Index is an expression at index, which should be integer or string.
	Index ExprNode
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *IndexAccessNode) Token() *Token {
	return n.Operand.Token()
}

// Note: Currently only ! is a logical unary operator

// NotOpNode is node for unary ! operator.
type NotOpNode struct {
	// Operand is an expression at operand of ! operator.
	Operand ExprNode
	tok     *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *NotOpNode) Token() *Token {
	return n.tok
}

// CompareOpNodeKind is a kind of compare operators; ==, !=, <, <=, >, >=.
type CompareOpNodeKind int

const (
	// CompareOpNodeKindInvalid is invalid and initial value of CompareOpNodeKind values.
	CompareOpNodeKindInvalid CompareOpNodeKind = iota
	// CompareOpNodeKindLess is kind for < operator.
	CompareOpNodeKindLess
	// CompareOpNodeKindLessEq is kind for <= operator.
	CompareOpNodeKindLessEq
	// CompareOpNodeKindGreater is kind for > operator.
	CompareOpNodeKindGreater
	// CompareOpNodeKindGreaterEq is kind for >= operator.
	CompareOpNodeKindGreaterEq
	// CompareOpNodeKindEq is kind for == operator.
	CompareOpNodeKindEq
	// CompareOpNodeKindNotEq is kind for != operator.
	CompareOpNodeKindNotEq
)

// CompareOpNode is node for binary expression to compare values; ==, !=, <, <=, > or >=.
type CompareOpNode struct {
	// Kind is a kind of this expression to show which operator is used.
	Kind CompareOpNodeKind
	// Left is an expression for left hand side of the binary operator.
	Left ExprNode
	// Right is an expression for right hand side of the binary operator.
	Right ExprNode
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *CompareOpNode) Token() *Token {
	return n.Left.Token()
}

// LogicalOpNodeKind is a kind of logical operators; && and ||.
type LogicalOpNodeKind int

const (
	// LogicalOpNodeKindInvalid is an invalid and initial value of LogicalOpNodeKind.
	LogicalOpNodeKindInvalid LogicalOpNodeKind = iota
	// LogicalOpNodeKindAnd is a kind for && operator.
	LogicalOpNodeKindAnd
	// LogicalOpNodeKindOr is a kind for || operator.
	LogicalOpNodeKindOr
)

func (k LogicalOpNodeKind) String() string {
	switch k {
	case LogicalOpNodeKindAnd:
		return "&&"
	case LogicalOpNodeKindOr:
		return "||"
	default:
		return "INVALID LOGICAL OPERATOR"
	}
}

// LogicalOpNode is node for logical binary operators; && or ||.
type LogicalOpNode struct {
	// Kind is a kind to show which operator is used.
	Kind LogicalOpNodeKind
	// Left is an expression for left hand side of the binary operator.
	Left ExprNode
	// Right is an expression for right hand side of the binary operator.
	Right ExprNode
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *LogicalOpNode) Token() *Token {
	return n.Left.Token()
}

// FuncCallNode represents function call in expression.
// Note that currently only calling builtin functions is supported.
type FuncCallNode struct {
	// Callee is a name of called function. This is string value because currently only built-in
	// functions can be called.
	Callee string
	// Args is arguments of the function call.
	Args []ExprNode
	tok  *Token
}

// Token returns the first token of the node. This method is useful to get position of this node.
func (n *FuncCallNode) Token() *Token {
	return n.tok
}

// Parser API

func errorAtToken(t *Token, msg string) *ExprError {
	return &ExprError{
		Message: msg,
		Offset:  t.Offset,
		Line:    t.Line,
		Column:  t.Column,
	}
}

// ExprParser is a parser for expression syntax. To know the details, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type ExprParser struct {
	input []*Token
}

// NewExprParser creates new ExprParser instance.
func NewExprParser() *ExprParser {
	return &ExprParser{}
}

func (p *ExprParser) error(msg string) *ExprError {
	return errorAtToken(p.input[0], msg)
}

func (p *ExprParser) errorf(format string, args ...interface{}) *ExprError {
	msg := fmt.Sprintf(format, args...)
	return errorAtToken(p.input[0], msg)
}

func (p *ExprParser) unexpected(where string, expected []TokenKind) *ExprError {
	qs := make([]string, 0, len(expected))
	for _, e := range expected {
		qs = append(qs, strconv.Quote(DescribeTokenKind(e)))
	}
	var what string
	if p.input[0].Kind == TokenKindEnd {
		what = "end of input"
	} else {
		what = fmt.Sprintf("token %q", DescribeTokenKind(p.input[0].Kind))
	}
	msg := fmt.Sprintf("unexpected %s while parsing %s. expecting %s", what, where, strings.Join(qs, ", "))
	return p.error(msg)
}

func (p *ExprParser) next() *Token {
	// Do not consume final token to remember position of the end
	if p.input[0].Kind == TokenKindEnd {
		return p.input[0]
	}
	t := p.input[0]
	p.input = p.input[1:]
	return t
}

func (p *ExprParser) peek() *Token {
	return p.input[0]
}

func (p *ExprParser) parseIdent() (ExprNode, *ExprError) {
	ident := p.next() // eat ident
	switch p.peek().Kind {
	case TokenKindLeftParen:
		// Parse function call as primary expression though generally function call is parsed as
		// postfix expression. The reason is that only built-in function call is allowed in workflow
		// expression syntax, meant that callee is always built-in function name, not a general
		// expression.
		p.next() // eat '('
		args := []ExprNode{}
		if p.peek().Kind == TokenKindRightParen {
			// no arguments
			p.next() // eat ')'
		} else {
		LoopArgs:
			for {
				arg, err := p.parseLogicalOr()
				if err != nil {
					return nil, err
				}

				args = append(args, arg)

				switch p.peek().Kind {
				case TokenKindComma:
					p.next() // eat ','
					// continue to next argument
				case TokenKindRightParen:
					p.next() // eat ')'
					break LoopArgs
				default:
					return nil, p.unexpected("arguments of function call", []TokenKind{TokenKindComma, TokenKindRightParen})
				}
			}
		}
		return &FuncCallNode{ident.Value, args, ident}, nil
	default:
		// Handle keywords
		switch ident.Value {
		case "null":
			return &NullNode{ident}, nil
		case "true":
			return &BoolNode{true, ident}, nil
		case "false":
			return &BoolNode{false, ident}, nil
		default:
			return &VariableNode{ident.Value, ident}, nil
		}
	}
}

func (p *ExprParser) parseNestedExpr() (ExprNode, *ExprError) {
	p.next() // eat '('

	nested, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}

	if p.peek().Kind == TokenKindRightParen {
		p.next() // eat ')'
	} else {
		return nil, p.unexpected("closing ')' of nexted expression (...)", []TokenKind{TokenKindRightParen})
	}

	return nested, nil
}

func (p *ExprParser) parseInt() (ExprNode, *ExprError) {
	t := p.peek()
	i, err := strconv.ParseInt(t.Value, 0, 32)
	if err != nil {
		return nil, p.errorf("parsing invalid integer literal %q: %s", t.Value, err)
	}

	p.next() // eat int

	return &IntNode{int(i), t}, nil
}

func (p *ExprParser) parseFloat() (ExprNode, *ExprError) {
	t := p.peek()
	f, err := strconv.ParseFloat(t.Value, 64)
	if err != nil {
		return nil, p.errorf("parsing invalid float literal %q: %s", t.Value, err)
	}

	p.next() // eat float

	return &FloatNode{f, t}, nil
}

func (p *ExprParser) parseString() ExprNode {
	t := p.next() // eat string
	s := t.Value
	s = s[1 : len(s)-1]                  // strip first and last single quotes
	s = strings.ReplaceAll(s, "''", "'") // unescape ''
	return &StringNode{s, t}
}

func (p *ExprParser) parsePrimaryExpr() (ExprNode, *ExprError) {
	switch p.peek().Kind {
	case TokenKindIdent:
		return p.parseIdent()
	case TokenKindLeftParen:
		return p.parseNestedExpr()
	case TokenKindInt:
		return p.parseInt()
	case TokenKindFloat:
		return p.parseFloat()
	case TokenKindString:
		return p.parseString(), nil
	default:
		return nil, p.unexpected(
			"variable access, function call, null, bool, int, float or string",
			[]TokenKind{
				TokenKindIdent,
				TokenKindLeftParen,
				TokenKindInt,
				TokenKindFloat,
				TokenKindString,
			},
		)
	}
}

func (p *ExprParser) parsePostfixOp() (ExprNode, *ExprError) {
	ret, err := p.parsePrimaryExpr()
	if err != nil {
		return nil, err
	}

	for {
		switch p.peek().Kind {
		case TokenKindDot:
			p.next() // eat '.'
			switch p.peek().Kind {
			case TokenKindStar:
				p.next() // eat '*'
				ret = &ArrayDerefNode{ret}
			case TokenKindIdent:
				t := p.next() // eat 'b' of 'a.b'
				ret = &ObjectDerefNode{ret, t.Value}
			default:
				return nil, p.unexpected(
					"object property dereference like 'a.b' or array element dereference like 'a.*'",
					[]TokenKind{TokenKindIdent, TokenKindStar},
				)
			}
		case TokenKindLeftBracket:
			p.next() // eat '['
			idx, err := p.parseLogicalOr()
			if err != nil {
				return nil, err
			}
			ret = &IndexAccessNode{ret, idx}
			if p.peek().Kind != TokenKindRightBracket {
				return nil, p.unexpected("closing bracket ']' for index access", []TokenKind{TokenKindRightBracket})
			}
			p.next() // eat ']'
		default:
			return ret, nil
		}
	}
}

func (p *ExprParser) parsePrefixOp() (ExprNode, *ExprError) {
	t := p.peek()
	if t.Kind != TokenKindNot {
		return p.parsePostfixOp()
	}
	p.next() // eat '!' token

	o, err := p.parsePostfixOp()
	if err != nil {
		return nil, err
	}

	return &NotOpNode{o, t}, nil
}

func (p *ExprParser) parseCompareBinOp() (ExprNode, *ExprError) {
	l, err := p.parsePrefixOp()
	if err != nil {
		return nil, err
	}

	k := CompareOpNodeKindInvalid
	switch p.peek().Kind {
	case TokenKindLess:
		k = CompareOpNodeKindLess
	case TokenKindLessEq:
		k = CompareOpNodeKindLessEq
	case TokenKindGreater:
		k = CompareOpNodeKindGreater
	case TokenKindGreaterEq:
		k = CompareOpNodeKindGreaterEq
	case TokenKindEq:
		k = CompareOpNodeKindEq
	case TokenKindNotEq:
		k = CompareOpNodeKindNotEq
	default:
		return l, nil
	}
	p.next() // eat the operator token

	r, err := p.parseCompareBinOp()
	if err != nil {
		return nil, err
	}

	return &CompareOpNode{k, l, r}, nil
}

func (p *ExprParser) parseLogicalAnd() (ExprNode, *ExprError) {
	l, err := p.parseCompareBinOp()
	if err != nil {
		return nil, err
	}
	if p.peek().Kind != TokenKindAnd {
		return l, nil
	}
	p.next() // eat &&
	r, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}
	return &LogicalOpNode{LogicalOpNodeKindAnd, l, r}, nil
}

func (p *ExprParser) parseLogicalOr() (ExprNode, *ExprError) {
	l, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}
	if p.peek().Kind != TokenKindOr {
		return l, nil
	}
	p.next() // eat ||
	r, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}
	return &LogicalOpNode{LogicalOpNodeKindOr, l, r}, nil
}

// Parse parses given token sequence into syntax tree.
// Token sequence t must end with TokenKindEnd token and it cannot be empty.
func (p *ExprParser) Parse(t []*Token) (ExprNode, *ExprError) {
	// Init
	if len(t) == 0 {
		panic("tokens must not be empty")
	}
	p.input = t

	root, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}

	if p.peek().Kind != TokenKindEnd {
		// It did not reach the end of sequence
		qs := make([]string, 0, len(p.input)-1)
		for _, t := range p.input {
			if t.Kind == TokenKindEnd {
				break
			}
			qs = append(qs, strconv.Quote(DescribeTokenKind(t.Kind)))
		}
		return nil, p.errorf("parser did not reach end of input after parsing expression. remaining tokens are %s", strings.Join(qs, ", "))
	}

	return root, nil
}
