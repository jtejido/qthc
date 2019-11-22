package qthc

import (
	"math"
)

type QueryIterator interface {
	HasNext() bool
	Next() *Entry
	Reset(min, max []float64)
}

const (
	EPS_MUL = 1.000000001
)

func isPointEnclosed(point, min, max []float64) bool {
	for d := 0; d < len(min); d++ {
		if point[d] < min[d] || point[d] > max[d] {
			return false
		}
	}
	return true
}

func isPointEnclosedFromCenter(point, center []float64, radius float64) bool {
	for d := 0; d < len(center); d++ {
		if point[d] < center[d]-radius || point[d] > center[d]+radius {
			return false
		}
	}
	return true
}

func isPointEqual(p1, p2 []float64) bool {
	for d := 0; d < len(p1); d++ {
		if p1[d] != p2[d] {
			return false
		}
	}
	return true
}

func overlap(min, max, min2, max2 []float64) bool {
	for d := 0; d < len(min); d++ {
		if max[d] < min2[d] || min[d] > max2[d] {
			return false
		}
	}
	return true
}

func isRectEnclosed(centerEnclosed []float64, radiusEnclosed float64, centerOuter []float64, radiusOuter float64) bool {
	for d := 0; d < len(centerOuter); d++ {
		radOuter := radiusOuter
		radEncl := radiusEnclosed
		if (centerOuter[d]+radOuter) < (centerEnclosed[d]+radEncl) || (centerOuter[d]-radOuter) > (centerEnclosed[d]-radEncl) {
			return false
		}
	}
	return true
}

func distance(p1, p2 []float64) float64 {
	var dist float64
	for i := 0; i < len(p1); i++ {
		d := p1[i] - p2[i]
		dist += d * d
	}
	return math.Sqrt(dist)
}

func distToRectNode(point, nodeCenter []float64, nodeRadius float64) float64 {
	var dist float64
	for i := 0; i < len(point); i++ {
		d := 0.
		if point[i] > nodeCenter[i]+nodeRadius {
			d = point[i] - (nodeCenter[i] + nodeRadius)
		} else if point[i] < nodeCenter[i]-nodeRadius {
			d = nodeCenter[i] - nodeRadius - point[i]
		}
		dist += d * d
	}
	return math.Sqrt(dist)
}
