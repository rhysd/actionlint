package actionlint_fuzz

import (
	"github.com/rhysd/actionlint"
)

func FuzzParse(data []byte) int {
	if !canParseByGoYAML(data) {
		return 0
	}

	if _, errs := actionlint.Parse(data); len(errs) > 0 {
		return 0
	}

	return 1
}
