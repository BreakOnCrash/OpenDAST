package dsl

import "testing"

func TestEval(t *testing.T) {
	config := &Config{
		Status: 200,
		Header: "Server: Nginx",
		Body:   "<h1>hello nginx!<h1>",
		Icon:   123,
	}

	// test number
	// for _, rule := range []string{
	// 	"status==200",
	// 	"status!=200",
	// 	"status>200",
	// 	"status<=200",
	// 	"status>=200",

	// 	"icon>200",
	// 	"icon==123",
	// 	"icon!=123",

	// 	`icon="123"`,     // error
	// 	"status=\"200\"", // error
	// 	"status=200",     // error
	// } {
	// 	eval(t, config, rule)
	// }

	for _, rule := range []string{
		`status==200 && (header="nginx" || body="nginx")`,
		`status==200 && (header="nginx" && body="nginx" && icon==123)`,
	} {
		eval(t, config, rule)
	}
}

func eval(t *testing.T, config *Config, rule string) {
	lexer, err := NewLexer(rule)
	if err != nil {
		t.Fatal(err)
	}

	dsl, err := TransFormExpr(lexer)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(dsl.Eval(config, true))
}
