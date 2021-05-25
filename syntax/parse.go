package syntax

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func kindString(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "Document"
	case yaml.SequenceNode:
		return "Sequence"
	case yaml.MappingNode:
		return "Mapping"
	case yaml.ScalarNode:
		return "Scalar"
	case yaml.AliasNode:
		return "Arias"
	default:
		panic("unreachable")
	}
}

func dump(n *yaml.Node, level int) {
	fmt.Printf("%s%s (%s, %d,%d): %q\n", strings.Repeat("  ", level), kindString(n.Kind), n.Tag, n.Line, n.Column, n.Value)
	for _, c := range n.Content {
		dump(c, level+1)
	}
}

func Parse(b []byte) error {
	var n yaml.Node

	if err := yaml.Unmarshal(b, &n); err != nil {
		return err
	}

	dump(&n, 0)

	return nil
}
