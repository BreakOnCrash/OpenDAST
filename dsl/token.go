// Package dsl 实现词法分析
// From https://github.com/Tencent/AI-Infra-Guard
package dsl

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// Token represents a lexical unit in the expression parsing
// 表示表达式解析中的词法单元
type Token struct {
	name    string // token type name
	content string // actual content of the token
	number  int    // number token value
}

// Constants defining different types of tokens
const (
	// Content type tokens
	tokenStatus = "status" // matches status code
	tokenBody   = "body"   // matches body content
	tokenHeader = "header" // matches HTTP headers
	tokenIcon   = "icon"   // matches icon content
	tokenText   = "text"   // matches text content
	tokenNumber = "number" // matches number

	// Comparison operators
	tokenContains   = "="  // contains operator
	tokenFullEqual  = "==" // exact match operator
	tokenNotEqual   = "!=" // not equal operator
	tokenRegexEqual = "~=" // regex match operator

	// Logical operators
	tokenAnd = "&&" // logical AND
	tokenOr  = "||" // logical OR

	tokenGt  = ">" // greater than
	tokenGte = ">="
	tokenLt  = "<" // less than
	tokenLte = "<="

	// Parentheses
	tokenLeftBracket  = "("
	tokenRightBracket = ")"
)

// ParseTokens converts input string to token sequence, supporting text content(quoted),
// comparison ops(=,==,!=,~=), logical ops(&&,||), parentheses and keywords(body,header,icon)
func ParseTokens(s1 string) ([]Token, error) {
	return parseTokensWithOptions(s1, []string{tokenStatus, tokenBody, tokenHeader, tokenIcon})
}

// parseTokensWithOptions 提取Token的公共解析函数
func parseTokensWithOptions(s1 string, validKeywords []string) ([]Token, error) {
	brackets := map[rune]string{'(': tokenLeftBracket, ')': tokenRightBracket}

	s, tokens := []rune(s1), []Token{}
	for i := 0; i < len(s); {
		x := s[i]
		switch true {
		case runeContains(x, '"'):
			token, newPos, err := parseQuotedText(s, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			i = newPos + 1
		case runeContains(x, '=', '~', '!', '|', '&', '>', '<'):
			token, skip, err := parseOperator(s[i:])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			i += skip
		case runeContains(x, '(', ')'):
			tokens = append(tokens, Token{
				name:    brackets[x],
				content: string(x),
			})
			i++
		case unicode.IsDigit(x):
			token, newPos, err := parseNumber(s[i:])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			i += newPos
		case runeContains(x, ' ', '\t', '\n', '\r'): // skip whitespace
			i++
		default:
			token, newPos, err := parseKeyword(s[i:], validKeywords)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			i += newPos
		}
	}
	return tokens, nil
}

// 辅助函数：解析引号内的文本
func parseQuotedText(s []rune, start int) (Token, int, error) {
	var n []rune
	i := start + 1
	for i < len(s) {
		if s[i] == '\\' { // skip escape '\"'
			n = append(n, s[i+1])
			i += 2
		} else if s[i] == '"' { // end of quoted
			return Token{name: tokenText, content: string(n)}, i, nil
		} else {
			n = append(n, s[i])
			i++
		}
	}
	return Token{}, 0, errors.New("unknown text:" + string(s[start:]))
}

// 辅助函数：解析操作符
func parseOperator(s []rune) (Token, int, error) {
	ops := []struct {
		name, content string
		skip          int
	}{
		{tokenFullEqual, "==", 2},
		{tokenContains, "=", 1},
		{tokenRegexEqual, "~=", 2},
		{tokenNotEqual, "!=", 2},
		{tokenOr, "||", 2},
		{tokenAnd, "&&", 2},
		{tokenGte, ">=", 2},
		{tokenLte, "<=", 2},
		{tokenGt, ">", 1},
		{tokenLt, "<", 1},
	}
	for _, op := range ops {
		if strings.HasPrefix(string(s), op.content) {
			return Token{name: op.name, content: op.content}, op.skip, nil
		}
	}
	return Token{}, 0, errors.New("invalid operator")
}

// 辅助函数：解析数字
func parseNumber(s []rune) (Token, int, error) {
	var num []rune

	for _, char := range s {
		if !unicode.IsDigit(char) {
			break
		}
		num = append(num, char)
	}

	if len(num) == 0 {
		return Token{}, 0, errors.New("unknown number")
	}

	val, err := strconv.Atoi(string(num))
	if err != nil {
		return Token{}, 0, err
	}

	return Token{name: tokenNumber, content: string(num), number: val}, len(num), nil
}

// 辅助函数：解析关键字
func parseKeyword(s []rune, validKeywords []string) (Token, int, error) {
	textOption := string(s)
	for _, check := range validKeywords {
		if strings.HasPrefix(textOption, check) {
			return Token{
				name:    check,
				content: check,
			}, len(check), nil
		}
	}
	return Token{}, 0, errors.New("unknown text: " + textOption)
}

func runeContains(x rune, prefix ...rune) bool {
	for _, pre := range prefix {
		if x == pre {
			return true
		}
	}

	return false
}
