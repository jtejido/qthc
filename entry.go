package qthc

type Entry struct {
	point []float64
	value interface{}
}

func NewEntry(key []float64, value interface{}) *Entry {
	ans := new(Entry)
	ans.point = key
	ans.value = value

	return ans
}

func (e *Entry) Point() []float64 {
	return e.point
}

func (e *Entry) Value() interface{} {
	return e.value
}

func (e *Entry) enclosed(min, max []float64) bool {
	return isPointEnclosed(e.point, min, max)
}

func (e *Entry) enclosedFromCenter(center []float64, radius float64) bool {
	return isPointEnclosedFromCenter(e.point, center, radius)
}

func (e *Entry) equals(ent *Entry) bool {
	return isPointEqual(e.point, ent.point)
}

type EntryDist struct {
	Entry
	dist float64
}

func NewEntryDist(e *Entry, distance float64) *EntryDist {
	ans := new(EntryDist)
	ans.point = e.point
	ans.value = e.value
	ans.dist = distance

	return ans
}

type byDistEntry []*EntryDist

func (a byDistEntry) Len() int           { return len(a) }
func (a byDistEntry) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDistEntry) Less(i, j int) bool { return a[i].dist < a[j].dist }
