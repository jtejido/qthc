/*
 * Copyright 2016-2017 Tilmann Zaeschke
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package qthc

import (
	"log"
	"math"
	"sort"
)

const (
	DEFAULT_MAX_NODE_SIZE int = 10
	MAX_DEPTH                 = 50
)

var (
	DEBUG = false
)

type QuadTree struct {
	dim, maxNodeSize, size int
	root                   *Node
}

func NewQuadTree(dim, maxNodeSize int) *QuadTree {
	if DEBUG {
		log.Println("Warning: DEBUG enabled")
		log.Println("Starting QuadTree...")
	}
	ans := new(QuadTree)
	ans.dim = dim
	ans.maxNodeSize = maxNodeSize
	return ans
}

func NewDefaultQuadTree(dim int) *QuadTree {
	maxNodeSize := DEFAULT_MAX_NODE_SIZE
	if 2*dim > DEFAULT_MAX_NODE_SIZE {
		maxNodeSize = 2 * dim
	}
	return NewQuadTree(dim, maxNodeSize)
}

func (qt *QuadTree) Insert(key []float64, value interface{}) {
	qt.size++
	e := NewEntry(key, value)
	if qt.root == nil {
		qt.initializeRoot(key)
	}

	qt.ensureCoverage(e)

	depth := 0
	r := qt.root
	for r != nil {
		r = r.tryPut(e, qt.maxNodeSize, depth > MAX_DEPTH)
		depth++
	}
}

func (qt *QuadTree) initializeRoot(key []float64) {
	lo := math.MaxFloat64
	hi := -math.MaxFloat64
	for d := 0; d < qt.dim; d++ {
		if lo > key[d] {
			lo = key[d]
		}
		if hi < key[d] {
			hi = key[d]
		}
	}
	if lo == 0 && hi == 0 {
		hi = 1.0
	}
	maxDistOrigin := lo
	if math.Abs(hi) > math.Abs(lo) {
		maxDistOrigin = hi
	}
	maxDistOrigin = math.Abs(maxDistOrigin)
	//no we use (0,0)/(+-maxDistOrigin*2,+-maxDistOrigin*2) as root.
	center := make([]float64, qt.dim)
	for d := 0; d < qt.dim; d++ {
		center[d] = -maxDistOrigin
		if key[d] > 0 {
			center[d] = maxDistOrigin
		}
	}

	qt.root = newNode(center, maxDistOrigin)
}

func (qt *QuadTree) Contains(key []float64) bool {
	if qt.root == nil {
		return false
	}

	return qt.root.getExact(key) != nil
}

func (qt *QuadTree) Get(key []float64) interface{} {
	if qt.root == nil {
		return nil
	}

	e := qt.root.getExact(key)

	if e == nil {
		return nil
	}

	return e.value
}

func (qt *QuadTree) Remove(key []float64) interface{} {
	if qt.root == nil {
		if DEBUG {
			log.Printf("Remove failure. Root is nil: %v \n", key)
		}

		return nil
	}
	e := qt.root.remove(nil, key, qt.maxNodeSize)
	if e == nil {
		if DEBUG {
			log.Printf("Remove failure. Not in root: %v \n", key)
		}
		return nil
	}

	qt.size--
	return e.value
}

func (qt *QuadTree) Update(oldKey, newKey []float64) interface{} {
	if qt.root == nil {
		return nil
	}
	requiresReinsert := []bool{false}
	e := qt.root.update(nil, oldKey, newKey, qt.maxNodeSize, requiresReinsert, 0, MAX_DEPTH)
	if e == nil {
		//not found
		if DEBUG {
			log.Printf("Reinsert failure: %v \n", newKey)
		}
		return nil
	}
	if requiresReinsert[0] {
		if DEBUG {
			log.Printf("Reinsert failure: %v \n", newKey)
		}
		//does not fit in root node...
		qt.ensureCoverage(e)
		depth := 0
		r := qt.root

		for r != nil {
			r = r.tryPut(e, qt.maxNodeSize, depth > MAX_DEPTH)
			depth++
		}
	}

	return e.value
}

func (qt *QuadTree) ensureCoverage(e *Entry) {
	p := e.point

	for !e.enclosedFromCenter(qt.root.center, qt.root.radius) {
		center := qt.root.center
		radius := qt.root.radius
		center2 := make([]float64, len(center))
		radius2 := radius * 2
		subNodePos := 0
		for d := 0; d < len(center); d++ {
			subNodePos <<= 1
			if p[d] < center[d]-radius {
				center2[d] = center[d] - radius
				//root will end up in upper quadrant in this
				//dimension
				subNodePos |= 1
			} else {
				//extend upwards, even if extension unnecessary for this dimension.
				center2[d] = center[d] + radius
			}
		}

		if DEBUG && !isRectEnclosed(center, radius, center2, radius2) {
			log.Printf("entry at %.4f: center/radius at %.4f/%.4f \n", e.point, center2, radius)
		}

		qt.root = newNodeWithSub(center2, radius2, qt.root, subNodePos)
	}
}

func (qt *QuadTree) Clear() {
	qt.size = 0
	qt.root = nil
}

func (qt *QuadTree) SearchIntersect(min, max []float64) QueryIterator {
	return newIterator(qt, min, max)
}

func (qt *QuadTree) NearestNeighbor(center []float64, k int) []*EntryDist {
	if qt.root == nil {
		return []*EntryDist{}
	}

	candidates := make([]*EntryDist, 0)
	rangeSearchKNN(qt.root, center, candidates, k, math.MaxFloat64)
	return candidates
}

func rangeSearchKNN(node *Node, center []float64, candidates []*EntryDist, k int, maxRange float64) float64 {
	posHC := node.calcSubPosition(center)
	entries := node.entries()
	var alreadyVisited interface{}
	if !node.isLeaf {
		//Search best node first
		ePos := entries[posHC]
		if v, ok := ePos.(*Node); ok {
			maxRange = rangeSearchKNN(v, center, candidates, k, maxRange)
			alreadyVisited = ePos
		}
	}
	//TODO first sort entries by distance!
	//TODO reuse buffer!
	buffer := make([]*KnnTemp, 0)
	for i := 0; i < len(entries); i++ {
		e := entries[i]
		if v, ok := e.(*Node); ok && e != alreadyVisited {
			n := v
			dist := distToRectNode(center, n.center, n.radius)
			addToBuffer(n, dist, maxRange, buffer)
		} else if v2, ok2 := e.(*Entry); ok2 {
			p := v2
			dist := distance(center, p.point)
			addToBuffer(p, dist, maxRange, buffer)
		}
	}
	sort.Sort(byDistKnn(buffer))

	for i := 0; i < len(buffer); i++ {
		t := buffer[i]
		if t.dist > maxRange {
			//check again, because maxDist may change during this loop
			continue
		}
		o := t.o
		if v, ok := o.(*Node); ok && o != alreadyVisited {
			maxRange = rangeSearchKNN(v, center, candidates, k, maxRange)
		} else if v2, ok2 := o.(*Entry); ok2 {
			p := v2
			candidates = append(candidates, NewEntryDist(p, t.dist))
			maxRange = adjustRegionKNN(candidates, k, maxRange)
		}
	}
	return maxRange
}

func addToBuffer(o interface{}, dist, maxDist float64, buffer []*KnnTemp) {
	if dist < maxDist {
		buffer = append(buffer, newKnnTemp(o, dist))
	}
}

func adjustRegionKNN(candidates []*EntryDist, k int, maxRange float64) float64 {
	if len(candidates) < k {
		//wait for more candidates
		return maxRange
	}

	//use stored distances instead of recalculating them
	sort.Sort(byDistEntry(candidates))

	for len(candidates) > k && len(candidates) > 0 {
		candidates = candidates[:len(candidates)-1]
	}

	return candidates[len(candidates)-1].dist
}

type KnnTemp struct {
	o    interface{}
	dist float64
}

func newKnnTemp(o interface{}, d float64) *KnnTemp {
	ans := new(KnnTemp)
	ans.o = o
	ans.dist = d
	return ans
}

type byDistKnn []*KnnTemp

func (a byDistKnn) Len() int           { return len(a) }
func (a byDistKnn) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDistKnn) Less(i, j int) bool { return a[i].dist < a[j].dist }
