package layout

import (
	"slices"
	"sync"
)

type Point struct {
	X, Y int
}

type Layout interface {
}

type Flex struct {
	Dir   Direction // direction of the main axis
	Items []FlexItem
}

func Column(items ...FlexItem) *Flex {
	return &Flex{Dir: Y, Items: items}
}

func Row(items ...FlexItem) *Flex {
	return &Flex{Dir: X, Items: items}
}


func (f Flex) StartLayouting(width, height int) {
	c := context{
		curDimensions: Dimensions{
			Origin: Point{X: 0, Y: 0},
			Width: width,
			Height: height,
		},
	}
	f.Layout(c)
}

// min size need to all be met
// max size are optional
// if min size is not met, do not layout box?
// rel size gets converted to abs size
func (f Flex) Layout(c context) {
	if f.Dir == Y {
		f.LayoutVertical(c)
	} else {
		f.LayoutHorizontal(c)
	}
}

func (f Flex) LayoutVertical(c context) {
	width, height := c.curDimensions.Origin.X + c.curDimensions.Width, c.curDimensions.Origin.Y + c.curDimensions.Height

	// 1. get abs size off all items, both min and max
	// 2. equally distribute remaining space to all items. If min size is not met, cancel that box layout and redistribute
	// 3. distribute remaining space to items, stopping if max size is met


	// only render items whose min size is actually met
	smallestPossibleSize := height / len(f.Items)
	itemsToLayout := filter(f.Items, func(i int, item FlexItem) bool { return item.Size.Min.toAbs(height)  <= smallestPossibleSize })

	sortedItemsToLayout := slices.Clone(itemsToLayout)
	slices.SortFunc(sortedItemsToLayout, func(a, b FlexItem) int {
		return a.Size.Max.toAbs(height) - b.Size.Max.toAbs(height);
	})

	filledSpace := make(map[int]int, len(sortedItemsToLayout))
	remainingSpace := height
	for tos, item := range sortedItemsToLayout {
		fill := item.Size.Max.toAbs(width)
		// if we can equally distribute remaining space to all items, do so
		if fill * len(sortedItemsToLayout[tos:]) <= remainingSpace {
			for _, item := range sortedItemsToLayout[tos:] {
				filledSpace[item.id] += fill
				remainingSpace -= fill
			}
		} else {
			fill = remainingSpace / len(sortedItemsToLayout[tos:])
			for _, item := range sortedItemsToLayout[tos:] {
				filledSpace[item.id] += fill
			}

			break
		}
	}

	// items must be layouted in the order they appear in the list, so that the origin is correct
	contextmap := make(map[int]Dimensions, len(itemsToLayout))
	orig := c.curDimensions.Origin
	for i, item := range itemsToLayout {
		dim := Dimensions{orig, width, filledSpace[item.id]}
		orig = Point{orig.X, orig.Y + dim.Height}
		contextmap[i] = dim
		item.Box(dim)
	}

	// recursiveley layout flex items
	for i, item := range itemsToLayout {
		if item.Flex != nil {
			dim := contextmap[i]
			newContext := context{dim}
			item.Flex.Layout(newContext)
		}
	}
}

func (f Flex) LayoutHorizontal(c context) {
	width, height := c.curDimensions.Origin.X + c.curDimensions.Width, c.curDimensions.Origin.Y + c.curDimensions.Height

	// 1. get abs size off all items, both min and max
	// 2. equally distribute remaining space to all items. If min size is not met, cancel that box layout and redistribute
	// 3. distribute remaining space to items, stopping if max size is met


	// only render items whose min size is actually met
	smallestPossibleSize := width / len(f.Items)
	itemsToLayout := filter(f.Items, func(i int, item FlexItem) bool { return item.Size.Min.toAbs(height)  <= smallestPossibleSize })

	sortedItemsToLayout := slices.Clone(itemsToLayout)
	slices.SortFunc(sortedItemsToLayout, func(a, b FlexItem) int {
		return a.Size.Max.toAbs(width) - b.Size.Max.toAbs(width);
	})

	filledSpace := make(map[int]int, len(sortedItemsToLayout))
	remainingSpace := width
	for tos, item := range sortedItemsToLayout {
		fill := item.Size.Max.toAbs(width)
		// if we can equally distribute remaining space to all items, do so
		if fill * len(sortedItemsToLayout[tos:]) <= remainingSpace {
			for _, item := range sortedItemsToLayout[tos:] {
				filledSpace[item.id] += fill
				remainingSpace -= fill
			}
		} else {
			fill = remainingSpace / len(sortedItemsToLayout[tos:])
			for _, item := range sortedItemsToLayout[tos:] {
				filledSpace[item.id] += fill
			}

			break
		}
	}

	// equally distribute remaining space to all items, stopping if > max size
	contextmap := make(map[int]Dimensions, len(itemsToLayout))
	orig := c.curDimensions.Origin
	for i, item := range itemsToLayout {
		dim := Dimensions{orig, filledSpace[item.id], height}
		orig = Point{orig.X + dim.Width, orig.Y}
		contextmap[i] = dim
		item.Box(dim)
	}

	// recursiveley layout flex items
	for i, item := range itemsToLayout {
		if item.Flex != nil {
			dim := contextmap[i]
			newContext := context{dim}
			item.Flex.Layout(newContext)
		}
	}
}

func filter[T any](ss []T, test func(i int, t T) bool) (ret []T) {
    for i, s := range ss {
        if test(i, s) {
            ret = append(ret, s)
        }
    }
    return
}

type AutoId struct {
	sync.Mutex
	id int
}

func (a *AutoId) ID() (id int) {
    a.Lock()
    defer a.Unlock()

    id = a.id
    a.id++
    return
}

var ai AutoId
type FlexItem struct {
	id int
	Box  LayoutBox
	Flex *Flex
	Size Constraint
}

func FlexItemBox(box LayoutBox, size Constraint, flex *Flex) FlexItem {
	return FlexItem{id: ai.ID(), Box: box, Size: size, Flex: flex}
}

type Constraint struct {
	Min, Max Size
}

func Exact(size Size) Constraint {
	return Constraint{Min: size, Max: size}
}

func Max(size Size) Constraint {
	return Constraint{Min: Abs(0), Max: size}
}

type Size struct {
	abs int // absolute size
	rel float64 // [0, 1]
}

func Abs(abs int) Size {
	return Size{abs: abs}
}

func Rel(rel float64) Size {
	return Size{rel: rel}
}

func (s Size) toAbs(size int) int {
	if s.abs != 0 {
		return s.abs
	}

	return int(s.rel * float64(size))
}

type Direction int

const (
	Y = iota
	X
)

type context struct {
	curDimensions Dimensions
}

// Resolve dimensions for a box
type Dimensions struct {
	Origin        Point // TL corner
	Width, Height int
}

// Should they be content aware?
type LayoutBox func(Dimensions)

func EmptyBox(Dimensions) {}
