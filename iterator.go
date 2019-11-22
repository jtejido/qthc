package qthc

import (
	"math"
)

type iterator struct {
	tree     *QuadTree
	stack    *IteratorStack
	next     *Entry
	min, max []float64
}

func newIterator(tree *QuadTree, min, max []float64) *iterator {
	ans := new(iterator)
	ans.stack = newIteratorStack()
	ans.tree = tree
	ans.Reset(min, max)

	return ans
}

func (it *iterator) HasNext() bool {
	return it.next != nil
}

func (it *iterator) Next() *Entry {
	ret := it.next
	it.findNext()
	return ret
}

/**
 * Reset the iterator. This iterator can be reused in order to reduce load on the
 * garbage collector.
 */
func (it *iterator) Reset(min, max []float64) {
	it.stack.clear()
	it.min = min
	it.max = max
	it.next = nil
	if it.tree.root != nil {
		it.stack.prepareAndPush(it.tree.root, min, max)
		it.findNext()
	}
}

func (it *iterator) findNext() {
	for !it.stack.isEmpty() {
		se := it.stack.peek()
		for se.pos < int64(se.len) {
			if se.isLeaf {
				e := se.entries[int(se.pos)].(*Entry)
				se.pos++
				if e.enclosed(it.min, it.max) {
					it.next = e
					return
				}
			} else {
				pos := int(se.pos)
				se.inc()
				//abort in next round if no increment is detected
				if se.pos <= int64(pos) {
					se.pos = math.MaxInt64
				}

				e := se.entries[pos]
				if e != nil {
					if v, ok := e.(*Node); ok {
						node := v
						se = it.stack.prepareAndPush(node, it.min, it.max)
					} else {
						qe := e.(*Entry)
						if qe.enclosed(it.min, it.max) {
							it.next = qe
							return
						}
					}
				}
			}
		}
		it.stack.pop()
	}
	it.next = nil
}

type IteratorStack struct {
	stack []*StackEntry
	size  int
}

func newIteratorStack() *IteratorStack {
	return &IteratorStack{make([]*StackEntry, 0), 0}
}

func (it *IteratorStack) isEmpty() bool {
	return it.size == 0
}

func (it *IteratorStack) prepareAndPush(node *Node, min, max []float64) *StackEntry {
	if it.size == len(it.stack) {
		it.stack = append(it.stack, new(StackEntry))
	}
	ni := it.stack[it.size]
	it.size++

	ni.set(node, min, max)
	return ni
}

func (it *IteratorStack) peek() *StackEntry {
	return it.stack[it.size-1]
}

func (it *IteratorStack) pop() *StackEntry {
	it.size--
	return it.stack[it.size]

}

func (it *IteratorStack) clear() {
	it.size = 0
}

type StackEntry struct {
	pos, m0, m1 int64
	entries     []interface{}
	isLeaf      bool
	len         int
}

func (se *StackEntry) set(node *Node, min, max []float64) {
	se.entries = node.entries()
	se.isLeaf = node.isLeaf

	if se.isLeaf {
		se.len = node.nValues
		se.pos = 0
	} else {
		se.len = len(se.entries)
		se.m0 = 0
		se.m1 = 0
		center := node.center
		for d := 0; d < len(center); d++ {
			se.m0 <<= 1
			se.m1 <<= 1
			if max[d] >= center[d] {
				se.m1 |= 1
				if min[d] >= center[d] {
					se.m0 |= 1
				}
			}
		}
		se.pos = se.m0
	}
}

func (se *StackEntry) inc() {
	//first, fill all 'invalid' bits with '1' (bits that can have only one value).
	r := se.pos | (^se.m1)
	//increment. The '1's in the invalid bits will cause bitwise overflow to the next valid bit.
	r++
	//remove invalid bits.
	se.pos = (r & se.m1) | se.m0

	//return -1 if we exceed 'max' and cause an overflow or return the original value. The
	//latter can happen if there is only one possible value (all filter bits are set).
	//return (r <= v) ? -1 : r;
}
