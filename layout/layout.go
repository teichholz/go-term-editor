package layout

type Point struct {
	X, Y int
}

type Layout interface {
}

type Flex struct {
	Dir   Direction // direction of the main axis
	Items []FlexItem
}

func (f Flex) StartLayouting(width, height int) {
	c := Context{
		constraints: Constraints{
			tl: Point{X: 0, Y: 0},
			br: Point{X: width, Y: height},
		},
	}
	f.Layout(c)
}

func (f Flex) Layout(c Context) {
	// calculate dimensinos
	width, height := c.constraints.br.X-c.constraints.tl.X, c.constraints.br.Y-c.constraints.tl.Y
	contextmap := make(map[int]Dimensions, len(f.Items))

	orig := c.constraints.tl
	for i, item := range f.Items {
		var dim Dimensions
		if f.Dir == Y {
			dim = Dimensions{width, int(float64(height) * item.Size), orig}
			orig = Point{orig.X, orig.Y + dim.Height}
		} else {
			dim = Dimensions{int(float64(width) * item.Size), height, orig}
			orig = Point{orig.X + dim.Width, orig.Y}
		}
		contextmap[i] = dim
		item.Box(dim)
	}

	for i, item := range f.Items {
		if item.Flex != nil {
			dim := contextmap[i]
			newContext := Context{Constraints{dim.Origin, Point{dim.Origin.X + dim.Width, dim.Origin.Y + dim.Height}}}
			item.Flex.Layout(newContext)
		}
	}
}

type FlexItem struct {
	Size float64 // [0, 1]
	Box  Box
	Flex *Flex
}

type Direction int

const (
	Y = iota
	X
)

type Context struct {
	constraints Constraints // tracks all the constraints for the current box
}

// Width or height. Can be absolute or relative. 200px or 50%
type Constraints struct {
	tl, br Point
}

// Resolve dimensions for a box
type Dimensions struct {
	Width, Height int
	Origin        Point // TL corner
}

// Should they be content aware?
type Box func(Dimensions)

func EmptyBox(Dimensions) {}
