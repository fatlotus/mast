package mast_test

import (
	"fmt"
	"github.com/fatlotus/mast"
)

func Example_parse() {
	tree, err := mast.PEMDAS.Parse("res = inv(B * B')")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", tree)
	// Output:
	// res = (inv (B * (' B)))
}

func Example_evaluate() {
	res := [][]float64{[]float64{0, 0}, []float64{0, 0}}
	b := [][]float64{[]float64{1, 2}, []float64{3, 4}}

	mast.MustEval("res = B * B'", &res, &b)
	fmt.Printf("res = %v\n", res)

	// Output:
	// res = [[5 11] [11 25]]
}
