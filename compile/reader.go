package compile

import (
	"bufio"
	"errors"
)

type reader struct {
	compEnv *CompilerEnvironment
	Lexer
	Token
	nestingLevel int
}

type Reader interface {
	Read() (SExpression, error)
}

func (r *reader) getCdr() (SExpression, error) {
	if r.Token.GetKind() == TokenKindRPAREN {
		return NewConsCell(NewNil(), NewNil()), nil
	}
	if r.Token.GetKind() == TokenKindDot {
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
	if r.Token.GetKind() == TokenKindNumber {
		value := r.GetInt()
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return Number(value), nil
	}

	if r.Token.GetKind() == TokenKindString {
		value := r.GetString()
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		symbolIndex := r.compEnv.GetCompilerSymbol(value)
		return NewString(symbolIndex), nil
	}

	if r.Token.GetKind() == TokenKindSymbol {
		value := r.GetSymbol()
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		symbolIndex := r.compEnv.GetCompilerSymbol(value)
		return NewSymbol(symbolIndex), nil
	}
	if r.Token.GetKind() == TokenKindBoolean {
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
	if r.Token.GetKind() == TokenKindNil {
		if r.nestingLevel != 0 {
			nextToken, err := r.GetNextToken()
			if err != nil {
				return nil, err
			}
			r.Token = nextToken
		}
		return NewNil(), nil
	}
	if r.Token.GetKind() == TokenKindQuote {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		symbolIndex := r.compEnv.GetCompilerSymbol("quote")
		return NewConsCell(NewSymbol(symbolIndex), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}

	if r.Token.GetKind() == TokenKindUnquote {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}

		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		symbolIndex := r.compEnv.GetCompilerSymbol("unquote")
		return NewConsCell(NewSymbol(symbolIndex), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}

	if r.Token.GetKind() == TokenKindUnquoteSplicing {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		symbolIndex := r.compEnv.GetCompilerSymbol("unquote-splicing")
		return NewConsCell(NewSymbol(symbolIndex), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}

	if r.Token.GetKind() == TokenKindQuasiquote {
		nextToken, err := r.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		sexp, err := r.sExpression()
		if err != nil {
			return nil, err
		}
		symbolIndex := r.compEnv.GetCompilerSymbol("quasiquote")
		return NewConsCell(NewSymbol(symbolIndex), NewConsCell(sexp, NewConsCell(NewNil(), NewNil()))), nil
	}
	if r.Token.GetKind() == TokenKindLparen {
		r.nestingLevel += 1
		nextToken, err := r.Lexer.GetNextToken()
		if err != nil {
			return nil, err
		}
		r.Token = nextToken
		if r.Token.GetKind() == TokenKindRPAREN {
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

func NewReader(compEnv *CompilerEnvironment, in *bufio.Reader) Reader {
	return &reader{
		Lexer:        New(in),
		Token:        nil,
		nestingLevel: 0,
		compEnv:      compEnv,
	}
}
