package dsl

import "testing"

func TestTransFormExp(t *testing.T) {
	s := `header="123" || (body="abc" && (header="efg" || icon="123"))`
	lexer, err := NewLexer(s)
	if err != nil {
		t.Fatal(err)
	}
	expr, err := TransFormExpr(lexer)
	if err != nil {
		t.Fatal(err)
	}

	expr.PrintAST()
}
