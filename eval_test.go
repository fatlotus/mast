package mast_test

import (
	"fmt"
	. "github.com/fatlotus/mast"
)

func handleError(e error) {
	panic(e)
}

// Evaluate a simple linear equation, handling the error.
func ExampleEval() {
	y := []float64{0, 0}
	A := [][]float64{[]float64{1, 2}, []float64{3, 4}}
	x := []float64{5, 6}
	b := []float64{7, 8}

	if err := Eval("y = A * x + b", &y, &A, &x, &b); err != nil {
		handleError(err)
		return
	}
	fmt.Printf("y = [%.2f %.2f]^T", y[0], y[1])
	// Output: y = [24.00 47.00]^T
}

// Evaluate a simple linear equation, but panic if something goes wrong.
func ExampleMustEval() {
	y := []float64{0, 0}
	A := [][]float64{[]float64{1, 2}, []float64{3, 4}}
	x := []float64{5, 6}
	b := []float64{7, 8}

	MustEval("y = A * x + b", &y, &A, &x, &b)
	fmt.Printf("y = [%.2f %.2f]^T", y[0], y[1])
	// Output: y = [24.00 47.00]^T
}
