//go:build gofuzz

package actionlint_fuzz

import (
	"github.com/rhysd/actionlint"
	"gopkg.in/yaml.v3"
)

func canParseByGoYAML(data []byte) (ret bool) {
	ret = true
	defer func() {
		if err := recover(); err != nil {
			ret = false
		}
	}()
	var n yaml.Node
	yaml.Unmarshal(data, &n)
	return
}

func FuzzParse(data []byte) int {
	if !canParseByGoYAML(data) {
		return 0
	}

	if _, errs := actionlint.Parse(data); len(errs) > 0 {
		return 0
	}

	return 1
}
