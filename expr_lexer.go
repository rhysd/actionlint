package actionlint

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

type TokenKind int

const (
	TokenKindUnknown TokenKind = iota
	TokenKindEnd
	TokenKindIdent
	TokenKindString
	TokenKindInt
	TokenKindFloat
	TokenKindLeftParen
	TokenKindRightParen
	TokenKindLeftBracket
	TokenKindRightBracket
	TokenKindDot
	TokenKindNot
	TokenKindLess
	TokenKindLessEq
	TokenKindGreater
	TokenKindGreaterEq
	TokenKindEq
	TokenKindNotEq
	TokenKindAnd
	TokenKindOr
	TokenKindStar
	TokenKindComma
)

func GetTokenKindName(t TokenKind) string {
	switch t {
	case TokenKindUnknown:
		return "Unknown"
	case TokenKindEnd:
		return "End"
	case TokenKindIdent:
		return "Ident"
	case TokenKindString:
		return "String"
	case TokenKindInt:
		return "Int"
	case TokenKindFloat:
		return "Float"
	case TokenKindLeftParen:
		return "LeftParen"
	case TokenKindRightParen:
		return "RightParen"
	case TokenKindLeftBracket:
		return "LeftBracket"
	case TokenKindRightBracket:
		return "RightBracket"
	case TokenKindDot:
		return "Dot"
	case TokenKindNot:
		return "Not"
	case TokenKindLess:
		return "Less"
	case TokenKindLessEq:
		return "LessEq"
	case TokenKindGreater:
		return "Greater"
	case TokenKindGreaterEq:
		return "GreaterEq"
	case TokenKindEq:
		return "Eq"
	case TokenKindNotEq:
		return "NotEq"
	case TokenKindAnd:
		return "And"
	case TokenKindOr:
		return "Or"
	case TokenKindStar:
		return "Star"
	case TokenKindComma:
		return "Comma"
	default:
		return "Unknown"
	}
}

type Token struct {
	Kind   TokenKind
	Value  string
	Offset int
	Line   int
	Column int
}

func (t *Token) String() string {
	return fmt.Sprintf("%s:%d:%d:%d", GetTokenKindName(t.Kind), t.Line, t.Column, t.Offset)
}

type ExprError struct {
	Message string
	Offset  int
	Line    int
	Column  int
}

func (e *ExprError) Error() string {
	return fmt.Sprintf("%d:%d:%d: %s", e.Line, e.Column, e.Offset, e.Message)
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\r' || r == '\t'
}

func isAlpha(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
}

func isNum(r rune) bool {
	return '0' <= r && r <= '9'
}

func isHexNum(r rune) bool {
	return isNum(r) || 'a' < r && r < 'f' || 'A' < r && r < 'F'
}

func isAlnum(r rune) bool {
	return isAlpha(r) || isNum(r)
}

func appendPuncts(rs []rune) []rune {
	return append(rs, '\'', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',')
}

func appendDigits(rs []rune) []rune {
	for r := '0'; r <= '9'; r++ {
		rs = append(rs, r)
	}
	return rs
}

func appendAlphas(rs []rune) []rune {
	for r := 'a'; r <= 'z'; r++ {
		rs = append(rs, r)
	}
	for r := 'A'; r <= 'Z'; r++ {
		rs = append(rs, r)
	}
	return rs
}

type ExprLexer struct {
	src     string
	scan    scanner.Scanner
	scanErr *ExprError
	start   int
}

func NewExprLexer() *ExprLexer {
	return &ExprLexer{}
}

func (lex *ExprLexer) token(kind TokenKind) *Token {
	p := lex.scan.Pos()
	s, e := lex.start, p.Offset
	t := &Token{
		Kind:   kind,
		Value:  lex.src[s:e],
		Offset: lex.scan.Offset,
		Line:   p.Line,
		Column: p.Column,
	}
	lex.start = p.Offset
	return t
}

func (lex *ExprLexer) skipWhite() {
	for {
		if r := lex.scan.Peek(); !isWhitespace(r) {
			return
		}
		lex.scan.Next()
		lex.start = lex.scan.Pos().Offset
	}
}

func (lex *ExprLexer) unexpected(r rune, where string, expected []rune) *ExprError {
	qs := make([]string, 0, len(expected))
	for _, e := range expected {
		qs = append(qs, strconv.QuoteRune(e))
	}
	msg := fmt.Sprintf(
		"got unexpected character %s while lexing %s, expecting %s",
		strconv.QuoteRune(r),
		where,
		strings.Join(qs, ", "),
	)
	p := lex.scan.Pos()
	return &ExprError{
		Message: msg,
		Offset:  p.Offset,
		Line:    p.Line,
		Column:  p.Column,
	}
}

func (lex *ExprLexer) unexpectedEOF() *ExprError {
	p := lex.scan.Pos()
	return &ExprError{
		Message: "unexpected EOF while lexing expression",
		Offset:  p.Offset,
		Line:    p.Line,
		Column:  p.Column,
	}
}

func (lex *ExprLexer) lexIdent() (*Token, *ExprError) {
	for {
		lex.scan.Next()
		// a-Z, 0-9, - or _
		// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
		if r := lex.scan.Peek(); !isAlnum(r) && r != '_' && r != '-' {
			return lex.token(TokenKindIdent), lex.scanErr
		}
	}
}

func (lex *ExprLexer) lexNum() (*Token, *ExprError) {
	r := lex.scan.Next() // precond: r is digit or '-'

	if r == '-' {
		r = lex.scan.Next()
		if !isNum(r) {
			return nil, lex.unexpected(r, "number after -", appendDigits([]rune{}))
		}
	}

	if r == '0' {
		r = lex.scan.Peek()
		if r == 'x' {
			lex.scan.Next()
			return lex.lexHexInt()
		}
		if isAlnum(r) && r != 'e' && r != 'E' {
			e := []rune{}
			e = appendPuncts(e)
			e = append(e, 'e', 'E')
			return nil, lex.unexpected(r, "number after 0", e)
		}
	} else {
		// r is 1..9
		for {
			r = lex.scan.Peek()
			if !isNum(r) {
				break
			}
			lex.scan.Next()
		}
	}

	k := TokenKindInt

	if r == '.' {
		lex.scan.Next() // eat '.'
		r = lex.scan.Next()
		if !isNum(r) {
			return nil, lex.unexpected(r, "fraction part of float number", appendDigits([]rune{}))
		}

		for {
			r = lex.scan.Peek()
			if !isNum(r) {
				break
			}
			lex.scan.Next()
		}

		k = TokenKindFloat
	}

	if r == 'e' || r == 'E' {
		lex.scan.Next() // eat 'e' or 'E'
		r = lex.scan.Next()
		if r == '-' {
			r = lex.scan.Next()
		}
		if !isNum(r) {
			return nil, lex.unexpected(r, "exponent part of float number", appendDigits([]rune{}))
		}

		for {
			r = lex.scan.Peek()
			if !isNum(r) {
				break
			}
			lex.scan.Next()
		}

		k = TokenKindFloat
	}

	return lex.token(k), lex.scanErr
}

func (lex *ExprLexer) lexHexInt() (*Token, *ExprError) {
	r := lex.scan.Next()
	if !isHexNum(r) {
		e := appendDigits([]rune{})
		for r := 'a'; r <= 'f'; r++ {
			e = append(e, r)
		}
		for r := 'A'; r <= 'F'; r++ {
			e = append(e, r)
		}
		return nil, lex.unexpected(r, "hex integer", e)
	}

	for {
		r = lex.scan.Peek()
		if !isHexNum(r) {
			break
		}
		lex.scan.Next()
	}

	return lex.token(TokenKindInt), lex.scanErr
}

func (lex *ExprLexer) lexString() (*Token, *ExprError) {
	lex.scan.Next() // eat '
	for {
		if lex.scan.Next() == '\'' {
			if lex.scan.Peek() == '\'' {
				lex.scan.Next() // eat second ' in ''
			} else {
				return lex.token(TokenKindString), lex.scanErr
			}
		}
	}
}

func (lex *ExprLexer) lexEnd() (*Token, *ExprError) {
	lex.scan.Next() // eat '}'
	if r := lex.scan.Next(); r != '}' {
		return nil, lex.unexpected(r, "end marker }}", []rune{'}'})
	}
	// }} is an end marker of interpolation
	return lex.token(TokenKindEnd), lex.scanErr
}

func (lex *ExprLexer) lexLess() (*Token, *ExprError) {
	lex.scan.Next() // eat '<'
	k := TokenKindLess
	if lex.scan.Peek() == '=' {
		k = TokenKindLessEq
		lex.scan.Next()
	}
	return lex.token(k), lex.scanErr
}

func (lex *ExprLexer) lexGreater() (*Token, *ExprError) {
	lex.scan.Next() // eat '>'
	k := TokenKindGreater
	if lex.scan.Peek() == '=' {
		k = TokenKindGreaterEq
		lex.scan.Next()
	}
	return lex.token(k), lex.scanErr
}

func (lex *ExprLexer) lexEq() (*Token, *ExprError) {
	lex.scan.Next() // eat '='
	if r := lex.scan.Next(); r != '=' {
		return nil, lex.unexpected(r, "== operator", []rune{'='})
	}
	return lex.token(TokenKindEq), lex.scanErr
}

func (lex *ExprLexer) lexBang() (*Token, *ExprError) {
	lex.scan.Next() // eat '!'
	k := TokenKindNot
	if lex.scan.Peek() == '=' {
		lex.scan.Next() // eat '='
		k = TokenKindNotEq
	}
	return lex.token(k), lex.scanErr
}

func (lex *ExprLexer) lexAnd() (*Token, *ExprError) {
	lex.scan.Next() // eat '&'
	if r := lex.scan.Next(); r != '&' {
		return nil, lex.unexpected(r, "&& operator", []rune{'&'})
	}
	return lex.token(TokenKindAnd), lex.scanErr
}

func (lex *ExprLexer) lexOr() (*Token, *ExprError) {
	lex.scan.Next() // eat '|'
	if r := lex.scan.Next(); r != '|' {
		return nil, lex.unexpected(r, "|| operator", []rune{'|'})
	}
	return lex.token(TokenKindOr), lex.scanErr
}

func (lex *ExprLexer) lexToken() (*Token, *ExprError) {
	lex.skipWhite()

	r := lex.scan.Peek()
	if r == scanner.EOF {
		return nil, lex.unexpectedEOF()
	}

	// Ident starts with a-Z or _
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	if isAlpha(r) || r == '_' {
		return lex.lexIdent()
	}

	if isNum(r) {
		return lex.lexNum()
	}

	switch r {
	case '-':
		return lex.lexNum()
	case '\'':
		return lex.lexString()
	case '}':
		return lex.lexEnd()
	case '!':
		return lex.lexBang()
	case '<':
		return lex.lexLess()
	case '>':
		return lex.lexGreater()
	case '=':
		return lex.lexEq()
	case '&':
		return lex.lexAnd()
	case '|':
		return lex.lexOr()
	}

	// Single character tokens
	lex.scan.Next()
	switch r {
	case '(':
		return lex.token(TokenKindLeftParen), lex.scanErr
	case ')':
		return lex.token(TokenKindRightParen), lex.scanErr
	case '[':
		return lex.token(TokenKindLeftBracket), lex.scanErr
	case ']':
		return lex.token(TokenKindRightBracket), lex.scanErr
	case '.':
		return lex.token(TokenKindDot), lex.scanErr
	case '*':
		return lex.token(TokenKindStar), lex.scanErr
	case ',':
		return lex.token(TokenKindComma), lex.scanErr
	}

	e := []rune{'_'} // Ident can start with _
	e = appendPuncts(e)
	e = appendDigits(e)
	e = appendAlphas(e)
	return nil, lex.unexpected(r, "expression", e)
}

func (lex *ExprLexer) init(src string) {
	lex.src = src
	lex.start = 0
	lex.scanErr = nil
	lex.scan.Init(strings.NewReader(src))
	lex.scan.Error = func(s *scanner.Scanner, m string) {
		lex.scanErr = &ExprError{
			Message: fmt.Sprintf("error while lexing expression: %s", m),
			Offset:  s.Offset,
			Line:    s.Line,
			Column:  s.Column,
		}
	}
}

func (lex *ExprLexer) Lex(src string) ([]*Token, int, *ExprError) {
	lex.init(src)
	ts := []*Token{}
	for {
		t, err := lex.lexToken()
		if err != nil {
			return nil, 0, err
		}
		ts = append(ts, t)
		if t.Kind == TokenKindEnd {
			return ts, lex.scan.Pos().Offset, nil
		}
	}
}
