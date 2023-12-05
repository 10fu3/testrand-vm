package compile

import (
	"fmt"
)

type TokenKind int

const (
	TokenKindNumber TokenKind = iota
	TokenKindFloat
	TokenKindBoolean
	TokenKindSymbol
	TokenKindLparen
	TokenKindRPAREN
	TokenKindDot
	TokenKindQuote
	TokenKindQuasiquote
	TokenKindUnquote
	TokenKindUnquoteSplicing
	TokenKindNil
	TokenKindString
)

type token struct {
	_kind   TokenKind
	_int    int64
	_float  float64
	_bool   bool
	_symbol string
	_string string
}

func (t *token) GetKind() TokenKind {
	return t._kind
}

func (t *token) GetInt() int64 {
	return t._int
}

func (t *token) GetFloat() float64 {
	return t._float
}

func (t *token) GetBool() bool {
	return t._bool
}

func (t *token) GetString() string {
	return t._string
}

func (t *token) String() string {
	// 数値
	if t._kind == TokenKindNumber {
		return fmt.Sprintf("Token (Number, %d )", t._int)
	}
	// 浮動小数点数
	if t._kind == TokenKindFloat {
		return fmt.Sprintf("Token (Float, %f )", t._float)
	}
	// 真理値
	if t._kind == TokenKindBoolean {
		return fmt.Sprintf("Token (Boolean, %t )", t._bool)
	}
	// 記号
	if t._kind == TokenKindSymbol {
		return fmt.Sprintf("Token (Symbol, %s )", t._symbol)
	}
	// 左括弧
	if t._kind == TokenKindLparen {
		return "Token (LeftPar)"
	}
	// 右括弧
	if t._kind == TokenKindRPAREN {
		return "Token (RightPar)"
	}
	// ドット
	if t._kind == TokenKindDot {
		return "Token (Dot)"
	}
	// クォート
	if t._kind == TokenKindQuote {
		return "Token (Quote)"
	}
	if t._kind == TokenKindUnquote {
		return "Token (Unquote)"
	}
	if t._kind == TokenKindUnquoteSplicing {
		return "Token (UnquoteSplicing)"
	}

	return "Token (Unknown)"
}

func (t *token) GetSymbol() string {
	return t._symbol
}

func NewTokenByInt(value int64) Token {
	return &token{
		_kind: TokenKindNumber,
		_int:  value,
	}
}

func NewTokenByFloat(value float64) Token {
	return &token{
		_kind:  TokenKindFloat,
		_float: value,
	}
}

func NewTokenByBool(value bool) Token {
	return &token{
		_kind: TokenKindBoolean,
		_bool: value,
	}
}

func NewTokenBySymbol(value string) Token {
	return &token{
		_kind:   TokenKindSymbol,
		_symbol: value,
	}
}

func NewTokenByKind(kind TokenKind) Token {
	return &token{
		_kind: kind,
	}
}

func NewTokenByNil() Token {
	return &token{_kind: TokenKindNil}
}

func NewTokenByString(value string) Token {
	return &token{
		_kind:   TokenKindString,
		_string: value,
	}
}
