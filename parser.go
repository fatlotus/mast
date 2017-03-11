package mast

import (
	"fmt"
	"unicode"
)

// A Parser configures the given operators in terms of precedence.
type Parser struct {
	// Define which operators appear in this language. These operators are
	// given in order of loosest to tightest (so addition/subtraction should
	// probably come before multiplication/division).
	Operators []Prec

	// Define the invisible grouping operators. These are not part of the tree,
	// but do override order of operations [as in (a + b) * c].
	Parens []Group

	// Define structural grouping operators. These behave like parens, except
	// they also also empty groups, and turn into Unary and Var nodes.
	//
	// Examples:
	//   []     = Var{"[]"}
	//   [x]    = Unary{"[]", Var{"x"}}
	//   [a, b] = Unary{"[]", Binary{",", "a", "b"}}
	Brackets []Group

	// If true, then "sin x" is legal and parses as "sin(x)" would. If false,
	// that is a syntax error.
	AdjacentIsApplication bool
}

// PEMDAS defines a typical multiply-first math language.
var PEMDAS Parser = Parser{
	Parens: []Group{
		{"(", ")"},
		{"[", "]"},
	},
	Brackets: []Group{
		{"{", "}"},
	},
	Operators: []Prec{
		{[]string{","}, InfixLeft},
		{[]string{"+", "-"}, InfixLeft},
		{[]string{"*", "/", "\\"}, InfixLeft},
		{[]string{"^"}, InfixRight},
		{[]string{"-", "+"}, Prefix},
		{[]string{"'"}, Suffix},
	},
	AdjacentIsApplication: true,
}

// A Group is a grouping operator, such as (), {}, or [].
type Group struct {
	Left  string
	Right string
}

// A Prec is a series of operators that share the same precedence,
// such as plus and minus. The Type represents the associativity
// and position of this operator.
type Prec struct {
	Glyphs []string
	Type   OpType
}

// An OpType represents the direction the operator associates and
// where it goes relative to the operands.
type OpType int

const (
	// eg. a + b + c == (a + b) + c
	InfixLeft OpType = iota

	// eg. a ^ b ^ c == a ^ (b ^ c)
	InfixRight

	// eg. - - a == - (- a)
	Prefix

	// eg. A'' = (A')'
	Suffix
)

// The Syntax tree returned by .Parse() is composed of Expr elements.
// Exprs are always of one of the following types:
//
//   Apply   sin(t)
//   Var     x
//   Unary   -w
//   Binary  a + b
//
type Expr interface {
	String() string
}

// A Var is a named variable in the environment. Variables can be single or
// multiple letters.
type Var struct {
	Name string
}

// Represent this Var as a context.
func (v *Var) String() string {
	return v.Name
}

// Represents function application, where an expression is invoked as an
// operator. Examples:
//
//   sin x == Apply{Var{"sin"}, Var{"x"}}
//   inv(A + B) == Apply{Var{"inv"}, Binary{"+", Var{"A"}, Var{"B"}}}
//
type Apply struct {
	Operator Expr
	Operand  Expr
}

// Represents this application as a string.
func (a *Apply) String() string {
	return fmt.Sprintf("(%v %v)", a.Operator, a.Operand)
}

// A Unary operator is one with one operator. Examples:
//
//   -a == Unary{"-", Var{"a"}}
//   A' == Unary{"'", Var{"A"}}
//
type Unary struct {
	Op   string
	Elem Expr
}

// Represent this unary operator as a string.
func (u *Unary) String() string {
	return fmt.Sprintf("(%s %s)", u.Op, u.Elem)
}

// A Binary operator is one with two operands. Examples:
//
//   a + b == Binary{"+", Var{"a"}, Var{"b"}}
//   a * b + c == Binary{"+", Binary{"*", Var{"a"}, Var{"b"}}, Var{"c"}}
//
type Binary struct {
	Op    string
	Left  Expr
	Right Expr
}

// Represent this binary operator as a string.
func (o *Binary) String() string {
	return fmt.Sprintf("(%s %s %s)", o.Left, o.Op, o.Right)
}

// An equation is an assignment of one side to the other. The engine provided
// can only evaluate an equation with a single variable on the left, but more
// advanced algebra systems could go further. Example:
//
//   x = A\b  ==  Equation{Var{"x"}, Binary{"\\", Var{"A"}, Var{"b"}}}
//
type Equation struct {
	Left  Expr
	Right Expr
}

// Represent this Equation as a string.
func (e *Equation) String() string {
	return fmt.Sprintf("%s = %s", e.Left, e.Right)
}

func isVar(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	return s != ""
}

func isEq(s string) bool {
	return s == "="
}

func isEof(s string) bool {
	return s == ""
}

func isOp(s string, ops []string) bool {
	if s == "" {
		return false
	}
	for _, op := range ops {
		if op == s {
			return true
		}
	}
	return false
}

// A parse error; all errors returned from .Parse are of this form. These
// indicate which token was found, and which tokens should have been provided.
type Unexpected struct {
	Found     string
	Expecting string
}

// Represent this Unexpected as a string.
func (u Unexpected) Error() string {
	found := u.Found
	if found == "" {
		found = "end-of-input"
	}
	result := fmt.Sprintf("unexpected %#v", u.Found)
	if u.Expecting != "" {
		result += fmt.Sprintf(", expecting %s", u.Expecting)
	}
	return result
}

func (p Parser) parseSingle(tokens []string, inApp bool) (lo []string, e Expr, err error) {
	// Look for a single variable
	if isVar(tokens[0]) {
		lo = tokens[1:]
		e = &Var{tokens[0]}
		var e2 Expr

		if inApp {
			return
		}

		for {
			apply := false
			for _, group := range append(p.Parens, p.Brackets...) {
				if lo[0] == group.Left {
					apply = true
					break
				}
			}

			if apply || (isVar(lo[0]) && p.AdjacentIsApplication) {
				lo, e2, err = p.parseSingle(lo, true)
				if err != nil {
					return
				}
				e = &Apply{e, e2}
			} else {
				break
			}
		}
		return
	}

	// Look for an open parenthesis
	for _, group := range p.Parens {
		if tokens[0] == group.Left {
			lo, e, err = p.parseExpr(0, tokens[1:])
			if lo[0] != group.Right {
				return lo, nil, &Unexpected{lo[0], fmt.Sprintf("%#v", group.Right)}
			}
			lo = lo[1:]
			return
		}
	}

	// Look for an open bracket
	for _, group := range p.Brackets {
		if tokens[0] == group.Left {
			if len(tokens) > 1 && tokens[1] == group.Right {
				lo = tokens[2:]
				e = &Var{group.Left + group.Right}
				return
			}
			lo, e, err = p.parseExpr(0, tokens[1:])
			if lo[0] != group.Right {
				return lo, nil, &Unexpected{lo[0], fmt.Sprintf("%#v", group.Right)}
			}
			lo = lo[1:]
			e = &Unary{group.Left + group.Right, e}
			return
		}
	}

	// Otherwise, compute what we might've wanted
	options := ""
	for _, op := range p.Operators {
		if op.Type == Prefix {
			for _, glyph := range op.Glyphs {
				options += fmt.Sprintf("\"%s\", ", glyph)
			}
		}
	}
	if options != "" {
		options += "or a variable"
	}
	return tokens, nil, &Unexpected{tokens[0], options}
}

func (p Parser) parseExpr(prec int, tokens []string) (lo []string, e Expr, err error) {
	if prec >= len(p.Operators) {
		return p.parseSingle(tokens, false)
	}

	var e2 Expr

	op := p.Operators[prec]
	lo = tokens

	switch op.Type {
	case Prefix:
		if glyph := lo[0]; isOp(glyph, op.Glyphs) {
			lo, e, err = p.parseExpr(prec, lo[1:])
			if err != nil {
				return lo, nil, err
			}
			return lo, &Unary{glyph, e}, nil
		}
		return p.parseExpr(prec+1, tokens)

	case InfixLeft:
		lo, e, err = p.parseExpr(prec+1, lo)
		for isOp(lo[0], op.Glyphs) {
			glyph := lo[0]
			lo, e2, err = p.parseExpr(prec+1, lo[1:])
			if err != nil {
				return lo, nil, err
			}
			e = &Binary{glyph, e, e2}
		}
		return

	case InfixRight:
		lo, e, err = p.parseExpr(prec+1, lo)
		if err != nil {
			return lo, nil, err
		}

		if glyph := lo[0]; isOp(glyph, op.Glyphs) {
			lo, e2, err = p.parseExpr(prec, lo[1:])
			if err != nil {
				return lo[1:], nil, err
			}
			e = &Binary{glyph, e, e2}
		}
		return

	case Suffix:
		lo, e, err = p.parseExpr(prec+1, lo)
		if err != nil {
			return lo, nil, err
		}

		for isOp(lo[0], op.Glyphs) {
			e = &Unary{lo[0], e}
			lo = lo[1:]
		}
		return
	}

	panic("should not get here")
}

func (p Parser) parseEqn(tokens []string) (lo []string, r *Equation, err error) {
	lo, lhs, err := p.parseExpr(0, tokens)
	if err != nil {
		return
	}

	if lo[0] != "=" {
		err = &Unexpected{lo[0], "="}
		return
	}

	lo, rhs, err := p.parseExpr(0, lo[1:])
	if err != nil {
		return
	}

	return lo, &Equation{lhs, rhs}, nil
}

// Parses an expression from source. On success, Expr is an expression; iff not,
// error is non-nil and of type Unexpected{}
func (p Parser) ParseExpr(source string) (Expr, error) {
	tokens, err := p.tokenize(source)
	if err != nil {
		return nil, err
	}
	lo, e, err := p.parseExpr(0, tokens)
	if err != nil {
		return nil, err
	} else if len(lo) > 1 || lo[0] != "" {
		return nil, &Unexpected{lo[0], "end-of-input"}
	}
	return e, nil
}

// Parses a single equation in the given source. On success, Equation is a
// parsed Equation; iff not, error is non-nil and of type Unexpected{}.
func (p Parser) Parse(source string) (*Equation, error) {
	tokens, err := p.tokenize(source)
	if err != nil {
		return nil, err
	}
	lo, e, err := p.parseEqn(tokens)
	if (len(lo) > 1 || lo[0] != "") && err == nil {
		return nil, &Unexpected{lo[0], "end-of-input"}
	}
	return e, err
}
