package qthc

import (
	"log"
)

type Node struct {
	center  []float64
	radius  float64
	values  []*Entry
	subs    []interface{}
	nValues int
	isLeaf  bool
}

func newNode(center []float64, radius float64) *Node {
	ans := new(Node)
	ans.center = center
	ans.radius = radius
	ans.values = make([]*Entry, 2)
	ans.isLeaf = true

	return ans
}

func newNodeWithSub(center []float64, radius float64, subNode *Node, subNodePos int) *Node {
	ans := new(Node)
	ans.center = center
	ans.radius = radius
	ans.values = nil
	ans.subs = make([]interface{}, 1<<uint(len(center)))
	ans.subs[subNodePos] = subNode
	ans.isLeaf = false

	return ans
}

func (n *Node) tryPut(e *Entry, maxNodeSize int, enforceLeaf bool) *Node {
	if DEBUG && !e.enclosedFromCenter(n.center, n.radius) {
		log.Printf("entry at %.4f: center/radius at %.4f/%.4f \n", e.point, n.center, n.radius)
	}

	//traverse subs?
	if !n.isLeaf {
		return n.getOrCreateSub(e, maxNodeSize, enforceLeaf)
	}

	//add if:
	//a) we have space
	//b) we have maxDepth
	//c) elements are equal (work only for n=1, avoids splitting
	//   in cases where splitting won't help. For n>1 the
	//   local limit is (temporarily) violated.
	if n.nValues < maxNodeSize || enforceLeaf || n.areAllPointsIdentical(e) {
		n.addValue(e, maxNodeSize)
		return nil
	}

	//split
	vals := n.values
	nVal := n.nValues
	n.clearValues()
	n.subs = make([]interface{}, 1<<uint(len(n.center)))
	n.isLeaf = false
	for i := 0; i < nVal; i++ {
		e2 := vals[i]
		sub := n.getOrCreateSub(e2, maxNodeSize, enforceLeaf)
		for sub != nil {
			//This may recurse if all entries fall
			//into the same subnode
			sub = sub.tryPut(e2, maxNodeSize, false)
		}
	}

	return n.getOrCreateSub(e, maxNodeSize, enforceLeaf)
}

func (n *Node) areAllPointsIdentical(e *Entry) bool {
	//This discovers situation where a node overflows, but splitting won't help because all points are identical
	for i := 0; i < n.nValues; i++ {
		if !e.equals(n.values[i]) {
			return false
		}
	}

	return true
}

func (n *Node) addValue(e *Entry, maxNodeSize int) {
	//Allow overflow over max node size (for example for lots of identical values in node)
	maxLen := maxNodeSize
	if n.nValues >= maxNodeSize {
		maxLen = n.nValues * 2
	}

	if n.nValues >= len(n.values) {
		l := maxLen
		if n.nValues*3 <= maxLen {
			l = n.nValues * 3
		}
		t := make([]*Entry, l)
		copy(t, n.values)
		n.values = t
	}
	n.values[n.nValues] = e
	n.nValues++
}

func (n *Node) removeValue(pos int) {
	if n.isLeaf {
		n.nValues--
		if pos < n.nValues {
			copy(n.values[pos+1:(pos+1)+(n.nValues-pos)], n.values[pos:pos+(n.nValues-pos)])
		}
	} else {
		n.nValues--
		n.subs[pos] = nil
	}
}

func (n *Node) clearValues() {
	n.values = nil
	n.nValues = 0
}

func (n *Node) getOrCreateSub(e *Entry, maxNodeSize int, enforceLeaf bool) *Node {
	pos := n.calcSubPosition(e.point)
	nn := n.subs[pos]

	if v, ok := nn.(*Node); ok {
		return v
	}

	if nn == nil {
		n.subs[pos] = e
		n.nValues++
		return nil
	}

	e2, _ := nn.(*Entry)
	n.nValues--
	sub := n.createSubForEntry(pos)
	n.subs[pos] = sub
	sub.tryPut(e2, maxNodeSize, enforceLeaf)

	return sub
}

func (n *Node) createSubForEntry(subNodePos int) *Node {
	centerSub := make([]float64, len(n.center))
	mask := 1 << uint(len(n.center))
	//This ensures that the subsnodes completely cover the area of
	//the parent node.
	radiusSub := n.radius / 2.0
	for d := 0; d < len(n.center); d++ {
		mask >>= 1
		if (subNodePos & mask) > 0 {
			centerSub[d] = n.center[d] + radiusSub
		} else {
			centerSub[d] = n.center[d] - radiusSub
		}
	}

	return newNode(centerSub, radiusSub)
}

func (n *Node) calcSubPosition(p []float64) int {
	subNodePos := 0
	for d := 0; d < len(n.center); d++ {
		subNodePos <<= 1
		if p[d] >= n.center[d] {
			subNodePos |= 1
		}
	}

	return subNodePos
}

func (n *Node) remove(parent *Node, key []float64, maxNodeSize int) *Entry {
	if !n.isLeaf {
		pos := n.calcSubPosition(key)
		o := n.subs[pos]
		if v, ok := o.(*Node); ok {
			return v.remove(n, key, maxNodeSize)
		} else if v2, ok2 := o.(*Entry); ok2 {
			e := v2
			if n.removeSub(parent, key, pos, e, maxNodeSize) {
				return e
			}
		}

		return nil
	}

	for i := 0; i < n.nValues; i++ {
		e := n.values[i]
		if n.removeSub(parent, key, i, e, maxNodeSize) {
			return e
		}
	}

	return nil
}

func (n *Node) removeSub(parent *Node, key []float64, pos int, e *Entry, maxNodeSize int) bool {
	if isPointEqual(e.point, key) {
		n.removeValue(pos)
		//TODO provide threshold for re-insert
		//i.e. do not always merge.
		if parent != nil {
			parent.checkAndMergeLeafNodes(maxNodeSize)
		}
		return true
	}
	return false
}

func (n *Node) update(parent *Node, keyOld, keyNew []float64, maxNodeSize int, requiresReinsert []bool, currentDepth, maxDepth int) *Entry {
	if !n.isLeaf {
		pos := n.calcSubPosition(keyOld)
		e := n.subs[pos]
		if e == nil {
			return nil
		}
		if v, ok := e.(*Node); ok {
			sub := v
			ret := sub.update(n, keyOld, keyNew, maxNodeSize, requiresReinsert, currentDepth+1, maxDepth)
			if ret != nil && requiresReinsert[0] && isPointEnclosedFromCenter(ret.point, n.center, n.radius/EPS_MUL) {
				requiresReinsert[0] = false
				r := n
				for r != nil {
					r = r.tryPut(ret, maxNodeSize, currentDepth > maxDepth)
					currentDepth++
				}
			}

			return ret
		}
		//Entry
		qe, _ := e.(*Entry)
		if isPointEqual(qe.point, keyOld) {
			n.removeValue(pos)
			qe.point = keyNew
			if isPointEnclosedFromCenter(keyNew, n.center, n.radius/EPS_MUL) {
				//reinsert locally;
				r := n
				for r != nil {
					r = r.tryPut(qe, maxNodeSize, currentDepth > maxDepth)
					currentDepth++
				}
				requiresReinsert[0] = false
			} else {
				requiresReinsert[0] = true
				if parent != nil {
					parent.checkAndMergeLeafNodes(maxNodeSize)
				}
			}
			return qe
		}
	}

	for i := 0; i < n.nValues; i++ {
		e := n.values[i]
		if isPointEqual(e.point, keyOld) {
			n.removeValue(i)
			e.point = keyNew
			n.updateSub(keyNew, e, parent, maxNodeSize, requiresReinsert)
			return e
		}
	}

	requiresReinsert[0] = false
	return nil
}

func (n *Node) updateSub(keyNew []float64, e *Entry, parent *Node, maxNodeSize int, requiresReinsert []bool) {
	if isPointEnclosedFromCenter(keyNew, n.center, n.radius/EPS_MUL) {
		//reinsert locally;
		n.addValue(e, maxNodeSize)
		requiresReinsert[0] = false
	} else {
		requiresReinsert[0] = true
		//TODO provide threshold for re-insert
		//i.e. do not always merge.
		if parent != nil {
			parent.checkAndMergeLeafNodes(maxNodeSize)
		}
	}
}

func (n *Node) checkAndMergeLeafNodes(maxNodeSize int) {
	//check: We start with including all local values: nValues
	nTotal := n.nValues
	for i := 0; i < len(n.subs); i++ {
		e := n.subs[i]
		if v, ok := e.(*Node); ok {
			sub := v
			if !sub.isLeaf {
				//can't merge directory nodes.
				//Merge only makes sense if we switch to list-mode, for which we don;t support subnodes!
				return
			}
			nTotal += sub.nValues
			if nTotal > maxNodeSize {
				//too many children
				return
			}
		}
	}

	//okay, let's merge
	n.values = make([]*Entry, nTotal)
	n.nValues = 0
	for i := 0; i < len(n.subs); i++ {
		e := n.subs[i]
		if v, ok := e.(*Node); ok {
			sub := v
			for j := 0; j < sub.nValues; j++ {
				n.values[n.nValues] = sub.values[j]
				n.nValues++
			}
		} else if v2, ok2 := e.(*Entry); ok2 {
			n.values[n.nValues] = v2
			n.nValues++
		}
	}

	n.subs = nil
	n.isLeaf = true
}

func (n *Node) getExact(key []float64) *Entry {
	if !n.isLeaf {
		pos := n.calcSubPosition(key)
		sub := n.subs[pos]
		if v, ok := sub.(*Node); ok {
			return v.getExact(key)
		} else if sub != nil {
			e, _ := sub.(*Entry)
			if isPointEqual(e.point, key) {
				return e
			}
		}
		return nil
	}

	for i := 0; i < n.nValues; i++ {
		e := n.values[i]
		if isPointEqual(e.point, key) {
			return e
		}
	}

	return nil
}

func (n *Node) entries() []interface{} {
	if n.isLeaf {
		r := make([]interface{}, len(n.values))
		for i := 0; i < len(n.values); i++ {
			r[i] = n.values[i]
		}

		return r
	}

	return n.subs
}
