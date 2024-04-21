package Rope

const (
	maxDepth    = 64
	maxLeafSize = 1024
)

// [start, end)
type Interval struct {
	Start, End int
}

type NodeInfo[V any, L Leaf[V]] interface {
	Accumulate(other *NodeInfo[V, L])
	ComputeInfo(leaf L) NodeInfo[V, L]
}

type HasInterval interface {
	Interval(len int) Interval
}

type DefaultNodeInfoInterval struct {}
func (node DefaultNodeInfoInterval) Interval(len int) Interval {
	return Interval{0, len}
}

type Leaf[V any] interface {
	Len() int
	PushMaybeSplit(l V) *V
}

type NodeBody[V any, L Leaf[V], N NodeInfo[V, L]] struct {
	height int
	len    int
	info   NodeVal
}

type NodeVal interface {}
type LeafVal[V any, L Leaf[V], N NodeInfo[V, L]] struct {
	val N
}
type InteralVal[V any, L Leaf[V], N NodeInfo[V, L]] struct {
	val []N
}




















type StringLeaf struct {
	*string
}

func (s StringLeaf) Len() int {
	return len(*s.string)
}

func (s StringLeaf) PushMaybeSplit(other StringLeaf) *StringLeaf {
	*s.string = *s.string + *other.string
	l := len(*s.string)
	if (l > maxLeafSize) {
		// TODO be smart here
        *s.string = (*s.string)[:l/2]
        secondHalf := (*s.string)[l/2:]
		return &StringLeaf{&secondHalf}
	}
	return nil
}