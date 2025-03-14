package js

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
	"github.com/robertkrimen/otto/token"
)

func ParseJSCode(jsCode, funcname string) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	program, err := parser.ParseFile(nil, "", jsCode, 0)
	if err != nil {
		return "", err
	}

	var (
		callExp     *ast.CallExpression
		localvalues = make(map[string]ast.Expression) // 临时变量缓存
	)

	for _, node := range program.Body {
		switch stmt := node.(type) {
		case *ast.ExpressionStatement:
			switch exp := stmt.Expression.(type) {
			case *ast.CallExpression:
				if callExp != nil {
					continue
				}

				if analyzeCallExpression(exp) == funcname {
					callExp = exp
				}
			case *ast.AssignExpression:
				if exp.Operator == token.ASSIGN {
					// 限制长度
					if len(localvalues) > 6 {
						break
					}
					if left, ok := exp.Left.(*ast.Identifier); ok {
						// 缓存参数
						localvalues[left.Name] = exp.Right
					}
				}
			case *ast.BinaryExpression:
				if callExp != nil {
					continue
				}
				// 右边为函数时
				if call, ok := exp.Right.(*ast.CallExpression); ok {
					if analyzeCallExpression(call) == funcname {
						callExp = call
					}
				}
			}
		}
	}

	if callExp == nil {
		return "", errors.New("not found function")
	}

	var arguments string
	for _, arg := range callExp.ArgumentList {
		paramStr := expressionToString(arg, localvalues)
		if paramStr != "" {
			arguments = arguments + "," + paramStr
		}
	}

	return arguments, nil
}

// analyzeCallExpression 分析函数调用
func analyzeCallExpression(call *ast.CallExpression) string {
	var funcName string
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		funcName = callee.Name
	case *ast.DotExpression:
		if ident, ok := callee.Left.(*ast.Identifier); ok {
			funcName = fmt.Sprintf("%s.%s", ident.Name, callee.Identifier.Name)
		}
	}
	return funcName
}

func expressionToString(expr ast.Expression, localvalues map[string]ast.Expression) string {
	switch e := expr.(type) {
	case *ast.ObjectLiteral:
		return objectLiteralToString(e, localvalues)
	case *ast.ArrayLiteral:
		return arrayLiteralToString(e, localvalues)
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	case *ast.NumberLiteral:
		return fmt.Sprintf("%v", e.Value)
	case *ast.BooleanLiteral:
		return fmt.Sprintf("%v", e.Value)
	case *ast.Identifier:
		value, ok := localvalues[e.Name]
		if !ok {
			return e.Name
		}
		return expressionToString(value, localvalues)
	}
	return ""
}

// objectLiteralToString 将对象字面量转换为字符串
func objectLiteralToString(obj *ast.ObjectLiteral, localvalues map[string]ast.Expression) string {
	var pairs []string
	for _, prop := range obj.Value {
		value := expressionToString(prop.Value, localvalues)
		pairs = append(pairs, fmt.Sprintf("%s:%s", prop.Key, value))
	}
	return "{" + strings.Join(pairs, ",") + "}"
}

// arrayLiteralToString 将数组字面量转换为字符串
func arrayLiteralToString(arr *ast.ArrayLiteral, localvalues map[string]ast.Expression) string {
	var elements []string
	for _, value := range arr.Value {
		elements = append(elements, expressionToString(value, localvalues))
	}
	return "[" + strings.Join(elements, ",") + "]"
}
