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
)

type Token struct {
	Kind   TokenKind
	Value  string
	Offset int
	Line   int
	Column int
}

type LexError struct {
	Message string
	Offset  int
	Line    int
	Column  int
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
	return append(rs, '\'', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|')
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
	scanErr *LexError
	start   int
}

func NewExprLexer() *ExprLexer {
	return &ExprLexer{}
}

func (lex *ExprLexer) token(kind TokenKind) *Token {
	s, e := lex.start, lex.scan.Offset
	t := &Token{
		Kind:   kind,
		Value:  lex.src[s:e],
		Offset: lex.scan.Offset,
		Line:   lex.scan.Line,
		Column: lex.scan.Column,
	}
	lex.start = lex.scan.Offset
	return t
}

func (lex *ExprLexer) skipWhite() {
	for {
		if r := lex.scan.Peek(); !isWhitespace(r) {
			return
		}
		lex.scan.Next()
		lex.start = lex.scan.Offset
	}
}

func (lex *ExprLexer) unexpected(r rune, expected []rune) *LexError {
	qs := make([]string, 0, len(expected))
	for _, e := range expected {
		qs = append(qs, strconv.QuoteRune(e))
	}
	msg := fmt.Sprintf("got unexpected character %s at offset %d, expecting %s", strconv.QuoteRune(r), lex.scan.Offset, strings.Join(qs, ", "))
	return &LexError{
		Message: msg,
		Offset:  lex.scan.Offset,
		Line:    lex.scan.Line,
		Column:  lex.scan.Column,
	}
}

func (lex *ExprLexer) unexpectedEOF() *LexError {
	return &LexError{
		Message: "Unexpected EOF while lexing expression",
		Offset:  lex.scan.Offset,
		Line:    lex.scan.Line,
		Column:  lex.scan.Column,
	}
}

func (lex *ExprLexer) lexIdent() (*Token, *LexError) {
	for {
		lex.scan.Next()
		if r := lex.scan.Peek(); !isAlnum(r) {
			return lex.token(TokenKindIdent), lex.scanErr
		}
	}
}

func (lex *ExprLexer) lexNum() (*Token, *LexError) {
	r := lex.scan.Next() // precond: r is digit or '-'

	if r == '-' {
		r = lex.scan.Next()
		if !isNum(r) {
			return nil, lex.unexpected(r, appendDigits([]rune{}))
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
			return nil, lex.unexpected(r, e)
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
			return nil, lex.unexpected(r, appendDigits([]rune{}))
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
			return nil, lex.unexpected(r, appendDigits([]rune{}))
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

func (lex *ExprLexer) lexHexInt() (*Token, *LexError) {
	r := lex.scan.Next()
	if !isHexNum(r) {
		e := appendDigits([]rune{})
		for r := 'a'; r <= 'f'; r++ {
			e = append(e, r)
		}
		for r := 'A'; r <= 'F'; r++ {
			e = append(e, r)
		}
		return nil, lex.unexpected(r, e)
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

func (lex *ExprLexer) lexString() (*Token, *LexError) {
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

func (lex *ExprLexer) lexEnd() (*Token, *LexError) {
	lex.scan.Next() // eat '}'
	if r := lex.scan.Next(); r != '}' {
		return nil, lex.unexpected(r, []rune{'}'})
	}
	// }} is an end marker of interpolation
	return lex.token(TokenKindEnd), lex.scanErr
}

func (lex *ExprLexer) lexLess() (*Token, *LexError) {
	lex.scan.Next() // eat '<'
	k := TokenKindLess
	if lex.scan.Peek() == '=' {
		k = TokenKindLessEq
		lex.scan.Next()
	}
	return lex.token(k), lex.scanErr
}

func (lex *ExprLexer) lexGreater() (*Token, *LexError) {
	lex.scan.Next() // eat '>'
	k := TokenKindGreater
	if lex.scan.Peek() == '=' {
		k = TokenKindGreaterEq
		lex.scan.Next()
	}
	return lex.token(k), lex.scanErr
}

func (lex *ExprLexer) lexEq() (*Token, *LexError) {
	lex.scan.Next() // eat '='
	if r := lex.scan.Next(); r != '=' {
		return nil, lex.unexpected(r, []rune{'='})
	}
	return lex.token(TokenKindEq), lex.scanErr
}

func (lex *ExprLexer) lexBang() (*Token, *LexError) {
	lex.scan.Next() // eat '!'
	k := TokenKindNot
	if lex.scan.Peek() == '=' {
		k = TokenKindNotEq
	}
	return lex.token(k), lex.scanErr
}

func (lex *ExprLexer) lexAnd() (*Token, *LexError) {
	lex.scan.Next() // eat '&'
	if r := lex.scan.Next(); r != '&' {
		return nil, lex.unexpected(r, []rune{'&'})
	}
	return lex.token(TokenKindAnd), lex.scanErr
}

func (lex *ExprLexer) lexOr() (*Token, *LexError) {
	lex.scan.Next() // eat '|'
	if r := lex.scan.Next(); r != '|' {
		return nil, lex.unexpected(r, []rune{'|'})
	}
	return lex.token(TokenKindOr), lex.scanErr
}

func (lex *ExprLexer) lexToken() (*Token, *LexError) {
	lex.skipWhite()

	r := lex.scan.Peek()
	if r == scanner.EOF {
		return nil, lex.unexpectedEOF()
	}

	if isAlpha(r) {
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
	default:
		e := []rune{}
		e = appendPuncts(e)
		e = appendDigits(e)
		e = appendAlphas(e)
		return nil, lex.unexpected(r, e)
	}
}

func (lex *ExprLexer) init(src string) {
	lex.src = src
	lex.start = 0
	lex.scanErr = nil
	lex.scan.Init(strings.NewReader(src))
	lex.scan.Error = func(s *scanner.Scanner, m string) {
		lex.scanErr = &LexError{
			Message: fmt.Sprintf("error while lexing expression: %s", m),
			Offset:  s.Offset,
			Line:    s.Line,
			Column:  s.Column,
		}
	}
}

func (lex *ExprLexer) Lex(src string) ([]*Token, int, *LexError) {
	lex.init(src)
	ts := []*Token{}
	for {
		t, err := lex.lexToken()
		if err != nil {
			return nil, 0, err
		}
		if t.Kind == TokenKindEnd {
			return ts, lex.scan.Offset, nil
		}
		ts = append(ts, t)
	}
}
