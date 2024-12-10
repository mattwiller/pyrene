package fhirpath

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
)

const (
	initialPrecendence uint8 = iota
	plusPrecedence
	minusPrecedence
	dotPrecedence
	lparenPrecedence
	rparenPrecedence
	commaPrecedence
)

var (
	dot    = []byte{'.'}
	comma  = []byte{','}
	lparen = []byte{'('}
	rparen = []byte{')'}
)

var booleanLiteral = regexp.MustCompile(`^(true|false)$`)

type Parser struct {
	input  []byte
	peeked Parselet
	idx    int
}

func NewParser(input []byte) *Parser {
	return &Parser{
		input:  input,
		peeked: Parselet{},
	}
}

var whitespace = []byte{' ', '\r', '\n', '\t'}

func (p *Parser) next() (Parselet, error) {
	if p.peeked.token != nil {
		next := p.peeked
		p.peeked = Parselet{}
		return next, nil
	} else if p.idx >= len(p.input) {
		return Parselet{}, io.EOF
	}
	c := p.input[p.idx]
	// Skip whitespace
	for bytes.ContainsRune(whitespace, rune(c)) {
		p.idx++
		c = p.input[p.idx]
	}
	switch c {
	case '.':
		parselet := Parselet{
			precedence: dotPrecedence,
			token:      p.input[p.idx : p.idx+1],
			infix: func(p *Parser, precedence uint8, left *Atom, token []byte) (*Atom, error) {
				right, err := p.parseExpression(precedence)
				if err != nil {
					return nil, fmt.Errorf("error parsing invocation: %w", err)
				}
				return InvocationExpression(left, right), nil
			},
		}
		p.idx++
		return parselet, nil
	case ',':
		parselet := Parselet{
			precedence: commaPrecedence,
			token:      p.input[p.idx : p.idx+1],
			infix: func(p *Parser, precedence uint8, left *Atom, token []byte) (*Atom, error) {
				right, err := p.parseExpression(precedence)
				if err != nil {
					return nil, fmt.Errorf("error parsing union: %w", err)
				}
				return Union(left, right), nil
			},
		}
		p.idx++
		return parselet, nil
	case '(':
		parselet := Parselet{
			precedence: lparenPrecedence,
			token:      p.input[p.idx : p.idx+1],
			prefix:     nil, // TODO: Implement paren group
			infix: func(p *Parser, precedence uint8, left *Atom, token []byte) (*Atom, error) {
				right, err := p.parseExpression(precedence)
				if err != nil {
					return nil, fmt.Errorf("error parsing function invocation: %w", err)
				}
				return Function(left, right), nil
			},
		}
		p.idx++
		return parselet, nil
	case ')':
		parselet := Parselet{
			precedence: rparenPrecedence,
			token:      p.input[p.idx : p.idx+1],
			prefix: func(p *Parser, precedence uint8, token []byte) (*Atom, error) {
				return ParamList(nil), nil
			},
			infix: func(p *Parser, precedence uint8, left *Atom, token []byte) (*Atom, error) {
				return ParamList(left), nil
			},
		}
		p.idx++
		return parselet, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		endIdx := p.idx + bytes.IndexFunc(p.input[p.idx:], func(r rune) bool {
			return (r < '0' || r > '9') && r != '.'
		})
		if endIdx == -1 {
			endIdx = len(p.input)
		}
		parselet := Parselet{
			precedence: initialPrecendence,
			token:      p.input[p.idx:endIdx],
			prefix: func(p *Parser, precedence uint8, token []byte) (*Atom, error) {
				return Number(token), nil
			},
		}
		p.idx = endIdx
		return parselet, nil
	case '\'':
		endIdx := p.idx + 1 + bytes.IndexRune(p.input[p.idx+1:], '\'')
		if endIdx == -1 {
			return Parselet{}, fmt.Errorf("unterminated string literal: %s", p.input[p.idx+1:])
		}
		parselet := Parselet{
			precedence: initialPrecendence,
			token:      p.input[p.idx+1 : endIdx],
			prefix: func(p *Parser, precedence uint8, token []byte) (*Atom, error) {
				return String(token), nil
			},
		}
		p.idx = endIdx + 1
		return parselet, nil
	default:
		if (c >= '0' && c <= '9') || !isValidIdentifierCharacter(p.input[p.idx]) {
			return Parselet{}, fmt.Errorf("invalid identifier character: %c", c)
		}
		endIdx := p.idx + bytes.IndexFunc(p.input[p.idx:], func(r rune) bool {
			return !isValidIdentifierCharacter(byte(r))
		})
		var parselet Parselet
		if booleanLiteral.Match(p.input[p.idx:endIdx]) {
			parselet = Parselet{
				precedence: initialPrecendence,
				token:      p.input[p.idx:endIdx],
				prefix: func(p *Parser, precedence uint8, token []byte) (*Atom, error) {
					return Boolean(token), nil
				},
			}
		} else {
			parselet = Parselet{
				precedence: initialPrecendence,
				token:      p.input[p.idx:endIdx],
				prefix: func(p *Parser, precedence uint8, token []byte) (*Atom, error) {
					return Identifier(token), nil
				},
			}
		}
		p.idx = endIdx
		return parselet, nil
	}
}

func (p *Parser) peek() Parselet {
	if p.peeked.token != nil {
		return p.peeked
	}
	next, _ := p.next()
	p.peeked = next
	return next
}

func (p *Parser) parseExpression(basePrecedence uint8) (*Atom, error) {
	parselet, err := p.next()
	if err == io.EOF {
		return nil, errors.New("unexpected end of input")
	} else if err != nil {
		return nil, err
	} else if parselet.prefix == nil {
		return nil, fmt.Errorf("unexpected token: %s", parselet.token)
	}

	left, err := parselet.prefix(p, parselet.precedence, parselet.token)
	if err != nil {
		return nil, fmt.Errorf("prefix parse error: %w", err)
	}
	for next := p.peek(); next.precedence > basePrecedence && next.infix != nil; next = p.peek() {
		p.next()
		left, err = next.infix(p, next.precedence, left, next.token)
		if err != nil {
			return nil, fmt.Errorf("infix parse error: %w", err)
		}
	}
	return left, nil
}

func isValidIdentifierCharacter(r byte) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

type PrefixFn = func(p *Parser, precedence uint8, token []byte) (*Atom, error)
type InfixFn = func(p *Parser, precedence uint8, left *Atom, token []byte) (*Atom, error)

type Parselet struct {
	token      []byte
	prefix     PrefixFn
	infix      InfixFn
	precedence uint8
}

func Parse(input []byte) (*Atom, error) {
	parser := NewParser(input)
	return parser.parseExpression(initialPrecendence)
}

type AtomType uint8

const (
	InvalidAtom AtomType = iota
	BooleanAtom
	StringAtom
	NumberAtom
	IdentifierAtom
	MemberInvocationAtom
	InvocationExpressionAtom
	FunctionAtom
	ParamListAtom
	UnionAtom
)

func (at AtomType) String() string {
	switch at {
	case BooleanAtom:
		return "Boolean"
	case StringAtom:
		return "String"
	case NumberAtom:
		return "Number"
	case IdentifierAtom:
		return "Identifier"
	case MemberInvocationAtom:
		return "MemberInvocation"
	case InvocationExpressionAtom:
		return "InvocationExpression"
	case FunctionAtom:
		return "Function"
	case ParamListAtom:
		return "ParamList"
	case UnionAtom:
		return "Union"
	default:
		return "<INVALID>"
	}
}

type Atom struct {
	left     *Atom
	right    *Atom
	token    []byte
	atomType AtomType
}

func (atom *Atom) Type() AtomType {
	return atom.atomType
}

func (atom *Atom) Equals(other *Atom) bool {
	// Handle nil pointers
	if atom == nil && other == nil {
		return true // nil == nil
	} else if atom == nil || other == nil {
		return false // nil != Atom{}
	}

	if atom.atomType != other.atomType {
		return false
	} else if !bytes.Equal(atom.token, other.token) {
		return false
	} else if !atom.left.Equals(other.left) {
		return false
	} else if !atom.right.Equals(other.right) {
		return false
	}
	return true
}

func InvocationExpression(left, right *Atom) *Atom {
	return &Atom{
		token:    dot,
		atomType: InvocationExpressionAtom,
		left:     left,
		right:    right,
	}
}

func Union(left, right *Atom) *Atom {
	return &Atom{
		token:    comma,
		atomType: UnionAtom,
		left:     left,
		right:    right,
	}
}

func Function(identifier, args *Atom) *Atom {
	return &Atom{
		token:    lparen,
		atomType: FunctionAtom,
		left:     identifier,
		right:    args,
	}
}

func ParamList(args *Atom) *Atom {
	return &Atom{
		token:    rparen,
		atomType: ParamListAtom,
		left:     args,
	}
}

func Number(literal []byte) *Atom {
	return &Atom{
		token:    literal,
		atomType: NumberAtom,
	}
}

func Boolean(literal []byte) *Atom {
	return &Atom{
		token:    literal,
		atomType: BooleanAtom,
	}
}

func Identifier(token []byte) *Atom {
	return &Atom{
		token:    token,
		atomType: IdentifierAtom,
	}
}

func String(token []byte) *Atom {
	return &Atom{
		token:    token,
		atomType: StringAtom,
	}
}
