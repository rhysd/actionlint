package actionlint

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

// TokenKind is kind of token.
type TokenKind int

const (
	// TokenKindUnknown is a default value of token as unknown token value.
	TokenKindUnknown TokenKind = iota
	// TokenKindEnd is a token for end of token sequence. Sequence without this
	// token means invalid.
	TokenKindEnd
	// TokenKindIdent is a token for identifier.
	TokenKindIdent
	// TokenKindString is a token for string literals.
	TokenKindString
	// TokenKindInt is a token for integers including hex integers.
	TokenKindInt
	// TokenKindFloat is a token for float numbers.
	TokenKindFloat
	// TokenKindLeftParen is a token for '('.
	TokenKindLeftParen
	// TokenKindRightParen is a token for ')'.
	TokenKindRightParen
	// TokenKindLeftBracket is a token for '['.
	TokenKindLeftBracket
	// TokenKindRightBracket is a token for ']'.
	TokenKindRightBracket
	// TokenKindDot is a token for '.'.
	TokenKindDot
	// TokenKindNot is a token for '!'.
	TokenKindNot
	// TokenKindLess is a token for '<'.
	TokenKindLess
	// TokenKindLessEq is a token for '<='.
	TokenKindLessEq
	// TokenKindGreater is a token for '>'.
	TokenKindGreater
	// TokenKindGreaterEq is a token for '>='.
	TokenKindGreaterEq
	// TokenKindEq is a token for '=='.
	TokenKindEq
	// TokenKindNotEq is a token for '!='.
	TokenKindNotEq
	// TokenKindAnd is a token for '&&'.
	TokenKindAnd
	// TokenKindOr is a token for '||'.
	TokenKindOr
	// TokenKindStar is a token for '*'.
	TokenKindStar
	// TokenKindComma is a token for ','.
	TokenKindComma
)

func (t TokenKind) String() string {
	switch t {
	case TokenKindUnknown:
		return "UNKNOWN"
	case TokenKindEnd:
		return "END"
	case TokenKindIdent:
		return "IDENT"
	case TokenKindString:
		return "STRING"
	case TokenKindInt:
		return "INTEGER"
	case TokenKindFloat:
		return "FLOAT"
	case TokenKindLeftParen:
		return "("
	case TokenKindRightParen:
		return ")"
	case TokenKindLeftBracket:
		return "["
	case TokenKindRightBracket:
		return "]"
	case TokenKindDot:
		return "."
	case TokenKindNot:
		return "!"
	case TokenKindLess:
		return "<"
	case TokenKindLessEq:
		return "<="
	case TokenKindGreater:
		return ">"
	case TokenKindGreaterEq:
		return ">="
	case TokenKindEq:
		return "=="
	case TokenKindNotEq:
		return "!="
	case TokenKindAnd:
		return "&&"
	case TokenKindOr:
		return "||"
	case TokenKindStar:
		return "*"
	case TokenKindComma:
		return ","
	default:
		panic("unreachable")
	}
}

// Token is a token lexed from expression syntax. For more details, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type Token struct {
	// Kind is kind of the token.
	Kind TokenKind
	// Value is string representation of the token.
	Value string
	// Offset is byte offset of token string starting.
	Offset int
	// Line is line number of start position of the token. Note that this value is 1-based.
	Line int
	// Column is column number of start position of the token. Note that this value is 1-based.
	Column int
}

func (t *Token) String() string {
	return fmt.Sprintf("%s:%d:%d:%d", t.Kind.String(), t.Line, t.Column, t.Offset)
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
	return isNum(r) || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F'
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

// ExprLexer is a struct to lex expression syntax. To know the syntax, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type ExprLexer struct {
	src     string
	scan    scanner.Scanner
	scanErr *ExprError
	start   scanner.Position
}

// NewExprLexer makes new ExprLexer instance.
func NewExprLexer() *ExprLexer {
	return &ExprLexer{}
}

func (lex *ExprLexer) token(kind TokenKind) *Token {
	p := lex.scan.Pos()
	s := lex.start
	t := &Token{
		Kind:   kind,
		Value:  lex.src[s.Offset:p.Offset],
		Offset: s.Offset,
		Line:   s.Line,
		Column: s.Column,
	}
	lex.start = p
	return t
}

func (lex *ExprLexer) skipWhite() {
	for {
		if r := lex.scan.Peek(); !isWhitespace(r) {
			return
		}
		lex.scan.Next()
		lex.start = lex.scan.Pos()
	}
}

func (lex *ExprLexer) unexpected(r rune, where string, expected []rune) *ExprError {
	qs := make([]string, 0, len(expected))
	for _, e := range expected {
		qs = append(qs, strconv.QuoteRune(e))
	}

	var what string
	if r == scanner.EOF {
		what = "EOF"
	} else {
		what = "character " + strconv.QuoteRune(r)
	}

	msg := fmt.Sprintf(
		"got unexpected %s while lexing %s, expecting %s",
		what,
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
	// Note: Should we lex NaN?

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
	lex.start = scanner.Position{
		Offset: 0,
		Line:   1,
		Column: 1,
	}
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

// Lex lexes the given string as expression syntax. The parameter must contain '}}' which represents
// end of expression. Otherwise this function will report an error that it encountered unexpected
// EOF.
func (lex *ExprLexer) Lex(src string) ([]*Token, int, *ExprError) {
	lex.init(src)
	ts := []*Token{}
	for {
		t, err := lex.lexToken()
		if err != nil {
			return nil, lex.scan.Pos().Offset, err
		}
		ts = append(ts, t)
		if t.Kind == TokenKindEnd {
			return ts, lex.scan.Pos().Offset, nil
		}
	}
}
