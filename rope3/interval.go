package BRope

// [Lo, Hi).
// Interval can be thought of as a sorted set of integers.
//
// Consider:
//
// [1, 2, 3] = [1, 4)
//
// [2, 3, 4] = [2, 5)
//
// Useful operations are:
//
// Prefix: [1]
//
// Suffix: [4]
//
// Union: [1, 2, 3, 4]
//
// Intersection: [2, 3]
type Interval struct {
	Lo, Hi int
}

func (i Interval) Len() int {
	return i.Hi - i.Lo
}

func (i Interval) IsEmpty() bool {
	return i.Lo >= i.Hi
}


func (i Interval) union(o Interval) Interval {
	if i.IsEmpty() { return o }
	if o.IsEmpty() { return i }
	return Interval{min(i.Lo, o.Lo), max(i.Hi, o.Hi)}
}

func (i Interval) Intersection(o Interval) Interval {
	start := max(i.Lo, o.Lo)
	end := min(i.Hi, o.Hi)
	return Interval{start, max(start, end)}
}

func (i Interval) Prefix(o Interval) Interval {
	return Interval{min(i.Lo, o.Lo), min(i.Hi, o.Lo)}
}

func (i Interval) Suffix(o Interval) Interval {
	return Interval{max(i.Lo, o.Hi), max(i.Hi, o.Hi)}
}

func (i Interval) Translate(n int) Interval {
	return Interval{i.Lo + n, i.Hi + n}
}

func (i Interval) IsBefore(n int) bool {
	return i.Hi <= n
}

func (i Interval) String() string {
	return "[" + string(rune(i.Lo)) + ", " + string(rune(i.Hi)) + ")"
}