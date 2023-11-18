package eval

import (
	"bufio"
	"errors"
	"testrand-vm/reader/lexer"
	"testrand-vm/reader/token"
)

type reader struct {
	lexer.Lexer
	token.Token
	nestingLevel int
}

type Reader interface {
	Read() (SExpression, error)
}

func (r *reader) getCdr() (SExpression, error) {
	if r.Token.GetKind() == token.TokenKindRPAREN {
		return NewConsCell(NewNil(), NewNil()), nil
	}
	if r.Token.GetKind() == token.TokenKindDot {
		nextToken, err := r.Lexer.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		return sexp, nil
	}
	car, err := r.sExpression()
	if err != nil {
		return nil, err
	}
	cdr, err := r.getCdr()
	return NewConsCell(car, cdr), nil
}

func (r *reader) sExpression() (SExpression, error) {
	if r.Token.GetKind() == token.TokenKindNumber {
		value := r.GetInt()
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return NewInt(value), nil
	}

	if r.Token.GetKind() == token.TokenKindString {
		value := r.GetString()
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return NewString(value), nil
	}

	if r.Token.GetKind() == token.TokenKindSymbol {
		value := r.GetSymbol()
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return NewSymbol(value), nil
	}
	if r.Token.GetKind() == token.TokenKindBoolean {
		value := r.GetBool()
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return NewBool(value), nil
	}
	if r.Token.GetKind() == token.TokenKindNil {
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return NewNil(), nil
	}
	if r.Token.GetKind() == token.TokenKindQuote {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		return NewConsCell(NewSymbol("quote"), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}

	if r.Token.GetKind() == token.TokenKindUnquote {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}

		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		return NewConsCell(NewSymbol("unquote"), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}

	if r.Token.GetKind() == token.TokenKindUnquoteSplicing {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		return NewConsCell(NewSymbol("unquote-splicing"), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}

	if r.Token.GetKind() == token.TokenKindQuasiquote {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		return NewConsCell(NewSymbol("quasiquote"), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}
	if r.Token.GetKind() == token.TokenKindLparen {
		r.nestingLevel += 1
		nextToken, err := r.Lexer.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		if r.Token.GetKind() == token.TokenKindRPAREN {
			r.nestingLevel -= 1
			if r.nestingLevel != 0 {
				nextToken, err = r.Lexer.GetNextToken()
				if err != nil {
					return nil, err
				}
				r.Token = nextToken
			}
			return NewConsCell(NewNil(), NewNil()), nil
		}
		car, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		cdr, err := r.getCdr()
		if err != nil {
			return nil, err
		}
		r.nestingLevel -= 1
		if r.nestingLevel != 0 {
			nextToken, err = r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return NewConsCell(car, cdr), nil
	}
	return nil, errors.New("Invalid expression: " + r.Token.String())
}

func (r *reader) Read() (SExpression, error) {
	r.nestingLevel = 0
	t, err := r.Lexer.GetNextToken()
	if err != nil {
		return nil, err
	}
	r.Token = t
	return r.sExpression()
}

func NewReader(in *bufio.Reader) Reader {
	return &reader{
		Lexer:        lexer.New(in),
		Token:        nil,
		nestingLevel: 0,
	}
}
