//go:build gofuzz

package actionlint_fuzz

import (
	"unicode/utf8"

	"github.com/rhysd/actionlint"
)

func FuzzExprParse(data []byte) int {
	if !utf8.Valid(data) {
		return 0
	}

	l := actionlint.NewExprLexer(string(data))
	p := actionlint.NewExprParser()
	e, err := p.Parse(l)
	if err != nil {
		return 0
	}

	c := actionlint.NewExprSemanticsChecker(true, nil)
	c.Check(e)

	return 1
}
