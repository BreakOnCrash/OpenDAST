package js

import (
	"fmt"
	"testing"
)

func TestParseJSCode(t *testing.T) {
	fmt.Println(ParseJSCode(`callback({"data": {username:"name"}});`, "callback"))
	fmt.Println(ParseJSCode(`callback({"data": {a:"name",test:0, args:["123", 1,{username:"xx"}]}});`, "callback"))
	fmt.Println(ParseJSCode(`callback([{"info": {"username": "name"}}])`, "callback"))

	fmt.Println(ParseJSCode(`cb('  {"username":"name"}  ')`, "cb"))

	fmt.Println(ParseJSCode(`/*aa*/ window.cb({"username":"name"})`, "window.cb"))
	fmt.Println(ParseJSCode(`/*aa*/ window.cb && window.cb({"username":"name"})`, "window.cb"))
	fmt.Println(ParseJSCode(`/*aa*/ cb && cb({"username":"name"})`, "cb"))

	fmt.Println(ParseJSCode(`a={"username": "name"}; cb({"s": a});`, "cb"))
	fmt.Println(ParseJSCode(`a={"username": "name"}; cb(a);`, "cb"))
}
