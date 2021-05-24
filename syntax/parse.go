package syntax

// import (
// 	"fmt"
//
// 	"github.com/goccy/go-yaml/ast"
// 	"github.com/goccy/go-yaml/parser"
// )
//
// type Dumper struct{}
//
// func (d *Dumper) Visit(n ast.Node) ast.Visitor {
// 	t := n.Type()
// 	p := n.GetToken().Position
// 	fmt.Printf("%s at (%d, %d):\n%s\n\n", t, p.Line, p.Column, n)
// 	return d
// }
//
// func ParseFile(file string) error {
// 	f, err := parser.ParseFile(file, 0)
// 	if err != nil {
// 		return err
// 	}
//
// 	fmt.Printf("name: %s\n", f.Name)
//
// 	v := &Dumper{}
// 	for _, d := range f.Docs {
// 		ast.Walk(v, d)
// 	}
//
// 	return nil
// }

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
