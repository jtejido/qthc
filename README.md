# qthc
MX-quadtree with Hypercube Traversal

![mx-quadtree](https://media.springernature.com/lw785/springer-static/image/chp%3A10.1007%2F978-3-319-10789-9_2/MediaObjects/316199_1_En_2_Fig11_HTML.gif)


This is an MX-quadtree implementation with configurable maximum depth, maximum nodes size, and
(if desired) automatic guessing of root rectangle. 

For navigation during queries, it uses the ***inc()*** traversal algorithm as described in: 
**T. Zaeschke and M. Norrie, "Efficient Z-Ordered Traversal of Hypercube Indexes,  BTW proceedings, 2017.**

The tree stores an object's center point (as opposed to a rectangle), together with its data, on its nodes.



## Usage:


![graph](https://i.imgur.com/8WPBz10l.png)


```golang
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

```

