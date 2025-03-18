// From https://github.com/Tencent/AI-Infra-Guard
package dsl

import "errors"

type Lexer struct {
	tokens      []Token // slice of tokens to process 要处理的token切片
	index       int     // current position in the stream 当前处理位置
	tokenLength int     // total number of tokens 总token数量
}

func NewLexer(s1 string) (lexer *Lexer, err error) {
	lexer = new(Lexer)
	lexer.tokens, err = ParseTokens(s1)
	if err != nil {
		return nil, err
	}

	lexer.tokenLength = len(lexer.tokens)
	return lexer, nil
}

// rewind moves the current position back by one
// 将当前位置回退一步
func (l *Lexer) rewind() {
	l.index -= 1
}

// next returns the next token in the stream and advances the position
// 返回流中的下一个token并前进位置
func (l *Lexer) next() (Token, error) {
	// Fix the logic error: check bounds before accessing token
	if l.index >= len(l.tokens) {
		return Token{}, errors.New("token index great token's length")
	}
	token := l.tokens[l.index]
	l.index += 1
	return token, nil
}

// hasNext checks if there are more tokens available in the stream
// 检查流中是否还有更多token可用
func (l *Lexer) hasNext() bool {
	return l.index < l.tokenLength
}
