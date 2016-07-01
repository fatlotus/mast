# Mast: a Math AST for Golang

[![CircleCI](https://img.shields.io/circleci/project/fatlotus/mast.svg?maxAge=2592000)](https://circleci.com/gh/fatlotus/mast) 

Because Go does not support operator overloading, most numerical libraries are 
messy to use. Mast is an attempt to write a domain-specific language for math,
so that code can be written in a more simple way.

If it helps, think [regular expressions](http://godoc.org/pkg/regexp), but for
algebra.

```go
// with Mast
mast.Eval("x = A' * b + c", &x, &A, &b, &c)

// without Mast
Transpose(&aT, &A)
Times(&Atb, &At, &b)
Plus(&x, &Atb, &c)
```

## Parser

Mast mostly exists to create parsers for math-like languages. To use it,
first create a *mast.Parser object, configuring the various operations
and their precidence.

```go
// A Parser configures the given operators in terms of precedence.
type Parser struct {
	// Define which operators appear in this language. These operators are
	// given in order of loosest to tightest (so addition/subtraction should
	// probably come before multiplication/division).
	Operators []Prec

	// Define which matching grouping operators are used.
	Groups []Group

	// If true, then "sin x" is legal and parses as "sin(x)" would. If false,
	// that is a syntax error.
	AdjacentIsApplication bool
}
```

Then you invoke the `(p *mast.Parser).Parse(string) (*mast.Equation, error)`
function.

```go
// Parses a single equation in the given source. On success, e is a
// parsed *Equation; iff not, error is non-nil and of type Unexpected{}.
func (p Parser) Parse(source string) (e *Equation, error) {
```

### Example

Suppose we want to make a basic calculator parser. First, define which 
operations and grouping operators we want to support:

```go
integers := &mast.Parser{
	// allow parentheses
	Groups: []Group{
		{"(", ")"},
	},
	
	// add plus, minux, multiply, divide, and modulo
	Operators: []Prec{
		{[]string{"+", "-"}, InfixLeft},
		{[]string{"*", "/", "%"}, InfixLeft},
		{[]string{"-", "+"}, Prefix},
	},
}
```

To parse a string using this language, invoke
`(p *mast.Parser) Parse(string) (*mast.Equation, error)` with the source code
to evaluate. For example, if we run:

```go
tree, err := integers.Parse("y = (-a) * b + c")
fmt.Printf("tree: %v\n", tree)
```

then `err` will be `nil` and `tree` will be as follows:

```go
tree := &Equation{
	Left: &Var{"y"},
	Right: &Binary{
		Op: "+",
		Left: &Binary{
			Op: "*",
			Left: &Unary{
				Op: "-",
				Elem: &Var{"a"},
			},
			Right: &Var{"b"},
		}
		Right: &Var{"c"}
	}
}
```

By iterating over the tree, your DSL can evaluate the mathematical
expression while maintaining type integrity.

## Evaluator

Mast includes a toy evaluator that handles matrices as [][]float64.
To use it, invoke the `Eval(code string, args ...interface{}) error`
function, passing pointers to the respective arguments.

```go
// Evaluate the given expression with the given variables. Variables are
// assigned left to right based on first usage.
func Eval(code string, args ...interface{}) error {
```

Think `%`-arguments to `fmt.Printf`. To make setting up variables easier,
arguments can be specified in three ways:

- a `[][]float64` for an `n x m` matrix;
- a `[]float64` for an `1 x n` column vector; or
- a `float64`, for a `1 x 1` scalar.

All other types panic.

### Example

Suppose we want to compute a linear transform (multiplying a vector by
a matrix, and adding a vector). First, we set up the variables to compute:

```go
A := [][]float64{
	[]float64{1, 2},
}

x := []float64{3, 4}
// same as := [][]float64{[]float64{5}, []float64{6}}

b := 5.0
// same as := [][]float64{[]float64{5.0}} 

y := 0
```

Once those are set up, the computation is fairly easy.

```go
if err := mast.Eval("y = A * x + b", &y, &A, &x, &b); err != nil {
	handleError(err)
	return
}
```

The result is then available in y.