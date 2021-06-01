package actionlint

import (
	"fmt"
	"strconv"
	"strings"
)

// ExprNode is a node of expression syntax tree. To know the syntax, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type ExprNode interface {
	Token() *Token
}

// Variable

type VariableNode struct {
	Name string
	tok  *Token
}

func (n *VariableNode) Token() *Token {
	return n.tok
}

// Literals

type NullNode struct {
	tok *Token
}

func (n *NullNode) Token() *Token {
	return n.tok
}

type BoolNode struct {
	Value bool
	tok   *Token
}

func (n *BoolNode) Token() *Token {
	return n.tok
}

type IntNode struct {
	Value int
	tok   *Token
}

func (n *IntNode) Token() *Token {
	return n.tok
}

type FloatNode struct {
	Value float64
	tok   *Token
}

func (n *FloatNode) Token() *Token {
	return n.tok
}

type StringNode struct {
	Value string
	tok   *Token
}

func (n *StringNode) Token() *Token {
	return n.tok
}

// Operators

// ObjectDerefNode represents property dereference of object like 'foo.bar'.
type ObjectDerefNode struct {
	Receiver ExprNode
	Property string
}

func (n ObjectDerefNode) Token() *Token {
	return n.Receiver.Token()
}

// ArrayDerefNode represents elements dereference of arrays like '*' in 'foo.bar.*.piyo'.
type ArrayDerefNode struct {
	Receiver ExprNode
}

func (n ArrayDerefNode) Token() *Token {
	return n.Receiver.Token()
}

type IndexAccessNode struct {
	Operand ExprNode
	Index   ExprNode
}

func (n *IndexAccessNode) Token() *Token {
	return n.Operand.Token()
}

// Note: Currently only ! is a logical unary operator

type NotOpNode struct {
	Operand ExprNode
	tok     *Token
}

func (n *NotOpNode) Token() *Token {
	return n.tok
}

type CompareOpNodeKind int

const (
	CompareOpNodeKindInvalid CompareOpNodeKind = iota
	CompareOpNodeKindLess
	CompareOpNodeKindLessEq
	CompareOpNodeKindGreater
	CompareOpNodeKindGreaterEq
	CompareOpNodeKindEq
	CompareOpNodeKindNotEq
)

type CompareOpNode struct {
	Kind  CompareOpNodeKind
	Left  ExprNode
	Right ExprNode
}

func (n *CompareOpNode) Token() *Token {
	return n.Left.Token()
}

type LogicalOpNodeKind int

const (
	LogicalOpNodeKindInvalid LogicalOpNodeKind = iota
	LogicalOpNodeKindAnd
	LogicalOpNodeKindOr
)

type LogicalOpNode struct {
	Kind  LogicalOpNodeKind
	Left  ExprNode
	Right ExprNode
}

func (n *LogicalOpNode) Token() *Token {
	return n.Left.Token()
}

// FuncCallNode represents function call in expression.
// Note that currently only calling builtin functions is supported.
type FuncCallNode struct {
	Callee string
	Args   []ExprNode
	tok    *Token
}

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

type ExprParser struct {
	input []*Token
}

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

// Parse parses token sequence into syntax tree.
// t must end with TokenKindEnd token and it cannot be empty slice.
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
