// +build gofuzz
package actionlint_fuzz

import (
	"unicode/utf8"

	"github.com/rhysd/actionlint"
)

func FuzzExprParse(data []byte) int {
	if !utf8.Valid(data) {
		return 0
	}

	l := actionlint.NewExprLexer()
	t, _, err := l.Lex(string(data))
	if err != nil {
		return 0
	}

	p := actionlint.NewExprParser()
	e, err := p.Parse(t)
	if err != nil {
		return 0
	}

	c := actionlint.NewExprSemanticsChecker()
	c.Check(e)

	return 1
}
