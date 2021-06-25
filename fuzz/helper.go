package actionlint_fuzz

import "gopkg.in/yaml.v3"

func canParseByGoYAML(data []byte) (ret bool) {
	defer func() {
		if err := recover(); err != nil {
			ret = false
		}
	}()
	var n yaml.Node
	yaml.Unmarshal(data, &n)
	return true
}
