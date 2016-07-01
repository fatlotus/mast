package mast

import (
	"fmt"
)

func addVars(e Expr, vars *[]string) {
	switch e := e.(type) {
	case *Var:
		for _, v := range *vars {
			if v == e.Name {
				return
			}
		}
		*vars = append(*vars, e.Name)
	case *Apply:
		addVars(e.Operator, vars)
		addVars(e.Operand, vars)
	case *Unary:
		addVars(e.Elem, vars)
	case *Binary:
		addVars(e.Left, vars)
		addVars(e.Right, vars)
	case *Equation:
		addVars(e.Left, vars)
		addVars(e.Right, vars)
	default:
		panic(fmt.Sprintf("strange Expr: %#v", e))
	}
}

func readMat(x interface{}) [][]float64 {
	switch x := x.(type) {
	case *float64:
		return [][]float64{[]float64{*x}}
	case *[]float64:
		result := make([][]float64, len(*x))
		for i, v := range *x {
			result[i] = []float64{v}
		}
		return result
	case *[][]float64:
		return *x
	default:
		panic(fmt.Errorf("unsupported type %#x", x))
	}
}

func writeMat(x interface{}, result [][]float64) {
	switch x := x.(type) {
	case *float64:
		if len(result) == 1 && len(result[0]) == 1 {
			*x = result[0][0]
		} else {
			panic("attempt to assign non-scalar value to scalar")
		}
	case *[]float64:
		if *x == nil {
			*x = make([]float64, len(result))
		} else if len(result) != len(*x) {
			panic("attempt to assign vectors of differing size")
		}

		for _, row := range result {
			if len(row) != 1 {
				panic("attempt to assign non-vector value to vector")
			}
		}

		for i, row := range result {
			(*x)[i] = row[0]
		}
	case *[][]float64:
		*x = result // FIXME
	default:
		panic(fmt.Errorf("unsupported type %#x", x))
	}
}

func addMats(a, b [][]float64) [][]float64 {
	// TODO: error handling

	result := make([][]float64, len(a))
	for i := range a {
		result[i] = make([]float64, len(a[0]))
		for j := range a[i] {
			result[i][j] = a[i][j] + b[i][j]
		}
	}
	return result
}

func dim(x [][]float64) (rows int, cols int) {
	if rows = len(x); rows == 0 {
		return
	}
	cols = len(x[0])
	for i, r := range x[1:] {
		if cols != len(r) {
			panic(fmt.Sprintf(
				"array size mismatch: [0] was an %d-slice, [%d] was an %d-slice",
				cols, i+1, len(r)))
		}
	}
	return
}

func transposeMat(a [][]float64) [][]float64 {
	n, m := dim(a)
	result := make([][]float64, m)

	for i := range result {
		result[i] = make([]float64, n)
	}

	for i, row := range a {
		for j := range row {
			result[j][i] = a[i][j]
		}
	}

	return result
}

func multMats(a, b [][]float64) [][]float64 {
	na, ma := dim(a)
	nb, mb := dim(b)

	if ma != nb {
		panic(fmt.Sprintf("cannot multiply %d-by-%d and %d-by-%d matrices", na, ma, nb, mb))
	}

	result := make([][]float64, len(a))
	for i := range a {
		result[i] = make([]float64, len(b[0]))
		for j := range b[0] {
			for k := range a[0] {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}

func eval(e Expr, vars map[string][][]float64) [][]float64 {
	switch e := e.(type) {
	case *Var:
		val, ok := vars[e.Name]
		if !ok {
			panic(fmt.Sprintf("undefined variable %#v\n", e.Name))
		}
		return val

	case *Apply: // treat all application as multiplication
		return multMats(eval(e.Operator, vars), eval(e.Operand, vars))

	case *Unary:
		switch e.Op {
		case "'":
			return transposeMat(eval(e.Elem, vars))
		default:
			panic(fmt.Sprintf("unknown unary operation: %s", e.Op))
		}

	case *Binary:
		switch e.Op {
		case "+":
			return addMats(eval(e.Left, vars), eval(e.Right, vars))
		case "*":
			return multMats(eval(e.Left, vars), eval(e.Right, vars))
		default:
			panic(fmt.Sprintf("unknown binary operation: %s", e.Op))
		}

	default:
		panic(fmt.Sprintf("strange Expr: %#v", e))
	}
}

// Evaluate the given expression with the given variables. Variables are
// assigned left to right based on first usage.
func Eval(code string, args ...interface{}) error {
	tree, err := PEMDAS.Parse(code)
	if err != nil {
		return err
	}

	if _, ok := tree.Left.(*Var); !ok {
		return fmt.Errorf("expression was %#v, but must be of the form \"y = ...\"", tree)
	}

	vars := []string{}
	addVars(tree, &vars)

	if len(vars) != len(args) {
		return fmt.Errorf("got %#v args, hoping for %d (to make %v)",
			len(args), len(vars), vars)
	}

	scope := map[string][][]float64{}
	for i, v := range vars[1:] {
		scope[v] = readMat(args[i+1])
	}

	writeMat(args[0], eval(tree.Right, scope))

	return nil
}

func MustEval(code string, args ...interface{}) {
	if err := Eval(code, args...); err != nil {
		panic(err)
	}
}
