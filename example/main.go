package main

import (
	"fmt"
	"github.com/jtejido/qthc"
)

func main() {
	qthc.DEBUG = true
	qt := qthc.NewDefaultQuadTree(2)

	things := [][]float64{
		[]float64{0, 0},
		[]float64{3, 1},
		[]float64{1, 2},
		[]float64{8, 6},
		[]float64{10, 3},
		[]float64{11, 7},
		[]float64{2, 6},
		[]float64{3, 6},
		[]float64{2, 8},
		[]float64{3, 8},
	}

	for _, thing := range things {
		qt.Insert(thing, nil)
	}

	q := qt.SearchIntersect([]float64{2, 1}, []float64{12, 7})

	for q.HasNext() {
		fmt.Println(q.Next())
	}

	// 3, 1
	// 2, 6
	// 3, 6
	// 10, 3
	// 8, 6
	// 11, 7

}
