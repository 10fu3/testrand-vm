package lexer

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"testrand-vm/reader/token"
)

const WHITESPACE_AT_EOL rune = ' '

var signs = map[rune]bool{
	'!': true,
	'$': true,
	'%': true,
	'&': true,
	'*': true,
	'+': true,
	'-': true,
	'.': true,
	'/': true,
	':': true,
	'<': true,
	'=': true,
	'>': true,
	'?': true,
	'@': true,
	'^': true,
	'_': true,
	'~': true,
	'|': true,
	'`': true,
	',': true,
}

type lexer struct {
	in        *bufio.Reader
	line      []rune
	lineIndex int
	nextRune  rune
}

type Lexer interface {
	GetNextToken() (token.Token, error)
}

func New(in *bufio.Reader) Lexer {
	return &lexer{
		in:        in,
		line:      make([]rune, 0),
		lineIndex: -1,
		nextRune:  ' ',
	}
}

func (l *lexer) updateNextChar() error {
	if l.lineIndex == len(l.line)-1 { // 次の行を読む.
		newLine, err := l.in.ReadString('\n')
		if err != nil {
			return err
		}
		l.line = []rune(fmt.Sprintf("%s%c", newLine, WHITESPACE_AT_EOL)) // 行末には必ず空白文字があることにする.
		l.lineIndex = 0
		l.nextRune = l.line[l.lineIndex]
	} else { // それ以外
		l.lineIndex++
		l.nextRune = l.line[l.lineIndex]
	}
	return nil
}

func (l *lexer) GetNextToken() (token.Token, error) {
	r := l.nextRune
	for isWhiteSpaceRune(r) {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		r = l.nextRune
	}
	if r == '(' {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		return token.NewTokenByKind(token.TokenKindLparen), nil
	}
	if r == ')' {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		return token.NewTokenByKind(token.TokenKindRPAREN), nil
	}
	if r == '.' {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		return token.NewTokenByKind(token.TokenKindDot), nil
	}
	if r == '#' {
		temp := make([]rune, 1)
		temp[0] = r
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		r = l.nextRune
		for isSymbolChar(r) {
			temp = append(temp, r)
			if err := l.updateNextChar(); err != nil {
				return nil, err
			}
			r = l.nextRune
		}
		temporarySymbol := string(temp)
		switch temporarySymbol {
		case "#t":
			return token.NewTokenByBool(true), nil
		case "#f":
			return token.NewTokenByBool(false), nil
		case "#nil":
			return token.NewTokenByNil(), nil
		}
		return nil, errors.New("invalid # constant")
	}
	if r == '\'' {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		return token.NewTokenByKind(token.TokenKindQuote), nil
	}
	if r == '`' {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		return token.NewTokenByKind(token.TokenKindQuasiquote), nil
	}
	if r == ',' {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		r = l.nextRune
		if r == '@' {
			if err := l.updateNextChar(); err != nil {
				return nil, err
			}
			return token.NewTokenByKind(token.TokenKindUnquoteSplicing), nil
		}
		return token.NewTokenByKind(token.TokenKindUnquote), nil
	}
	if isSymbolChar(r) {
		isBeginWithDigit := isDigit(r)
		var temp []rune
		for isSymbolChar(r) {
			temp = append(temp, r)
			if err := l.updateNextChar(); err != nil {
				return nil, err
			}
			r = l.nextRune
		}
		symbolSequence := string(temp)
		parseInt, err := strconv.ParseInt(symbolSequence, 10, 64)
		if err == nil {
			return token.NewTokenByInt(parseInt), nil
		}
		parseFloat, err := strconv.ParseFloat(symbolSequence, 64)
		if err == nil {
			return token.NewTokenByFloat(parseFloat), nil
		}
		if isBeginWithDigit {
			return nil, errors.New(fmt.Sprintf("unexpected word: %s", symbolSequence))
		}
		return token.NewTokenBySymbol(symbolSequence), nil
	}
	if r == '"' {
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		r = l.nextRune
		var temp []rune
		for r != '"' {
			temp = append(temp, r)
			if err := l.updateNextChar(); err != nil {
				return nil, err
			}
			r = l.nextRune
		}
		if err := l.updateNextChar(); err != nil {
			return nil, err
		}
		return token.NewTokenByString(string(temp)), nil
	}
	if err := l.updateNextChar(); err != nil {
		return nil, err
	}
	return nil, errors.New(fmt.Sprintf("unknown char: %s", string(r)))
}

func isWhiteSpaceRune(r rune) bool {
	return r == ' ' || r == '\r' || r == '\n' || r == '\t'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlphabet(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isSign(r rune) bool {
	if _, has := signs[r]; has {
		return true
	}
	return false
}

func isSymbolChar(r rune) bool {
	return isDigit(r) || isAlphabet(r) || isSign(r)
}
