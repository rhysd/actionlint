package actionlint

import "testing"

func TestParseExpression(t *testing.T) {
	l := NewExprLexer()
	tok, _, err := l.Lex("foo(1, true, 'hello')}}")
	if err != nil {
		panic(err)
	}

	p := NewExprParser()
	n, err := p.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}

	c, ok := n.(*FuncCallNode)
	if !ok {
		t.Fatalf("not a function call node at root: %#v", n)
	}

	if c.Callee != "foo" {
		t.Errorf("wanted \"foo\" as callee name but got %q", c.Callee)
	}
	if len(c.Args) != 3 {
		t.Errorf("wanted 3 arguments but got %#v", c.Args)
	}
}
