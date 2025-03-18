// Package dsl 实现AST语法解析
// From https://github.com/Tencent/AI-Infra-Guard
package dsl

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Expr 定义了表达式接口
// 所有表达式类型都需要实现 Name() 方法
type Expr interface {
	String() string
}

// Rule 表示一个规则，包含多个表达式
type Rule struct {
	root Expr
}

type matchExpr struct {
	op        string
	left      string
	right     string
	cacheRegx *regexp.Regexp
}

func (m matchExpr) String() string {
	return fmt.Sprintf("matchExpr: left: %s, op: %s, right: %s", m.left, m.op, m.right)
}

type matchNumberExpr struct {
	op    string
	left  string
	right int
}

func (m matchNumberExpr) String() string {
	return fmt.Sprintf("matchNumberExpr: left: %s, op: %s, right: %d", m.left, m.op, m.right)
}

type logicExpr struct {
	op    string
	left  Expr
	right Expr
}

func (l logicExpr) String() string {
	return fmt.Sprintf("logicExpr: left: %s, op: %s, right: %s", l.left.String(), l.op, l.right.String())
}

type bracketExpr struct {
	inner Expr
}

func (b bracketExpr) String() string {
	return fmt.Sprintf("bracketExpr: inner: %s", b.inner.String())
}

// TransFormExpr 将token序列转换为表达式规则
// 输入tokens切片，返回Rule对象和error
// 主要功能：解析tokens并构建DSL表达式、逻辑表达式和括号表达式
func TransFormExpr(lexer *Lexer) (*Rule, error) {
	root, err := parseExpr(lexer)
	if err != nil {
		return nil, err
	}

	if lexer.hasNext() {
		return nil, errors.New("unexpected tokens after expression")
	}

	return &Rule{root: root}, nil
}

// parseExpr 解析表达式
func parseExpr(lexer *Lexer) (Expr, error) {
	expr, err := parsePrimaryExpr(lexer)
	if err != nil {
		return nil, err
	}

	for lexer.hasNext() {
		token, err := lexer.next()
		if err != nil {
			return nil, err
		}
		if token.name == tokenAnd || token.name == tokenOr {
			right, err := parsePrimaryExpr(lexer)
			if err != nil {
				return nil, err
			}
			// 提高括号表达式的优先级
			if _, ok := right.(*bracketExpr); ok {
				expr = &logicExpr{op: token.content, left: right, right: expr}
			} else {
				expr = &logicExpr{op: token.content, left: expr, right: right}
			}
		} else {
			lexer.rewind()
			break
		}
	}
	return expr, nil
}

// parsePrimary 解析括号语句和基础表达式
func parsePrimaryExpr(lexer *Lexer) (Expr, error) {
	tmpToken, err := lexer.next()
	if err != nil {
		return nil, err
	}

	switch tmpToken.name {
	case tokenStatus, tokenIcon:
		p2, err := lexer.next()
		if err != nil {
			return nil, err
		}
		if !(p2.name == tokenFullEqual ||
			p2.name == tokenNotEqual ||
			p2.name == tokenGte ||
			p2.name == tokenLte ||
			p2.name == tokenGt ||
			p2.name == tokenLt) {
			return nil, errors.New("synax error in " + tmpToken.content + " " + p2.content)
		}
		p3, err := lexer.next()
		if err != nil {
			return nil, err
		}
		if p3.name != tokenNumber {
			return nil, errors.New("synax error in" + tmpToken.content + " " + p2.content + " " + p3.content)
		}
		return &matchNumberExpr{left: tmpToken.content, op: p2.content, right: p3.number}, nil
	case tokenBody, tokenHeader:
		p2, err := lexer.next()
		if err != nil {
			return nil, err
		}
		if !(p2.name == tokenContains ||
			p2.name == tokenFullEqual ||
			p2.name == tokenNotEqual ||
			p2.name == tokenRegexEqual) {
			return nil, errors.New("synax error in " + tmpToken.content + " " + p2.content)
		}
		p3, err := lexer.next()
		if err != nil {
			return nil, err
		}
		if p3.name != tokenText {
			return nil, errors.New("synax error in" + tmpToken.content + " " + p2.content + " " + p3.content)
		}
		// 正则缓存对象
		var expr matchExpr
		if p2.name == tokenRegexEqual {
			compile, err := regexp.Compile(p3.content)
			if err != nil {
				return nil, err
			}
			expr = matchExpr{left: tmpToken.content, op: p2.content, cacheRegx: compile}
		} else {
			expr = matchExpr{left: tmpToken.content, op: p2.content, right: p3.content}
		}
		return &expr, nil
	case tokenLeftBracket:
		inner, err := parseExpr(lexer)
		if err != nil {
			return nil, err
		}
		closingToken, err := lexer.next()
		if err != nil || closingToken.name != tokenRightBracket {
			return nil, errors.New("missing or invalid closing bracket")
		}
		return &bracketExpr{inner: inner}, nil
	default:
		return nil, errors.New("unexpected token: " + tmpToken.content)
	}
}

// PrintAST 递归打印表达式
func (r *Rule) PrintAST() {
	if r.root == nil {
		return
	}

	var printExpr func(expr Expr, level int)
	printExpr = func(expr Expr, level int) {
		indent := strings.Repeat("  ", level)

		switch e := expr.(type) {
		case *matchExpr:
			if e.cacheRegx != nil {
				fmt.Printf("%s    dslExp: %s %s regex('%s')\n", indent, e.left, e.op, e.cacheRegx.String())
			} else {
				fmt.Printf("%s    dslExp: %s %s '%s'\n", indent, e.left, e.op, e.right)
			}

		case *logicExpr:
			fmt.Printf("%s logicExp: %s\n", indent, e.op)
			fmt.Printf("%s  - left:\n", indent)
			printExpr(e.left, level+1)
			fmt.Printf("%s  - right:\n", indent)
			printExpr(e.right, level+1)

		case *bracketExpr:
			fmt.Printf("%s bracketExp:\n", indent)
			printExpr(e.inner, level+1)

		default:
			fmt.Printf("%s Unknown expression type\n", indent)
		}
	}

	printExpr(r.root, 0)
}

// Eval 评估规则是否匹配
// 输入配置对象，返回布尔值表示是否匹配
// 使用栈实现后缀表达式求值
func (r *Rule) Eval(config *Config, debug bool) bool {
	if r.root == nil {
		return false
	}

	var evalExpr func(expr Expr, config *Config) bool
	evalExpr = func(expr Expr, config *Config) bool {
		switch next := expr.(type) {
		case *matchExpr:
			var s1 string
			switch next.left {
			case tokenBody:
				s1 = config.Body
			case tokenHeader:
				s1 = config.Header
			default:
				panic("unknown left token")
			}
			s1 = strings.ToLower(s1)
			text := strings.ToLower(next.right)
			var r bool
			switch next.op {
			case tokenFullEqual:
				r = text == s1
			case tokenContains:
				r = strings.Contains(s1, text)
			case tokenNotEqual:
				r = !strings.Contains(s1, text)
			case tokenRegexEqual:
				r = next.cacheRegx.MatchString(s1)
			default:
				panic("unknown op token")
			}

			// TODO For DEBUG
			if debug {
				fmt.Printf("eval: %s, %v\n", next.String(), r)
			}

			return r
		case *matchNumberExpr:
			var n1 int
			switch next.left {
			case tokenStatus:
				n1 = config.Status
			case tokenIcon:
				n1 = int(config.Icon)
			default:
				panic("unknown left token")
			}
			var r bool
			switch next.op {
			case tokenFullEqual:
				r = next.right == n1
			case tokenNotEqual:
				r = next.right != n1
			case tokenGt:
				r = next.right > n1
			case tokenGte:
				r = next.right >= n1
			case tokenLt:
				r = next.right < n1
			case tokenLte:
				r = next.right <= n1
			default:
				panic("unknown op token")
			}

			// TODO For DEBUG
			if debug {
				fmt.Printf("eval: %s, %v\n", next.String(), r)
			}

			return r
		case *logicExpr:
			switch next.op {
			case tokenAnd:
				leftVal := evalExpr(next.left, config)
				if !leftVal { // short-circuit evaluation

					if debug {
						fmt.Printf("logicExpr: logic: %s, short-circuit evaluation, left: %v\n", next.op, leftVal)
					}

					return false
				}

				if debug {
					fmt.Printf("logicExpr: %v %s %s\n", leftVal, next.op, next.right)
				}
				return evalExpr(next.right, config)
			case tokenOr:
				leftVal := evalExpr(next.left, config)
				if leftVal { // short-circuit evaluation

					if debug {
						fmt.Printf("logicExpr: logic: %s, short-circuit evaluation, left: %v\n", next.op, leftVal)
					}

					return true
				}

				if debug {
					fmt.Printf("logicExpr: %v %s %s\n", leftVal, next.op, next.right)
				}

				return evalExpr(next.right, config)
			default:
				panic("unknown logic type")
			}
		case *bracketExpr:
			return evalExpr(next.inner, config)
		default:
			panic("error eval")
		}
	}
	return evalExpr(r.root, config)
}
