package layout

import (
	"fmt"
	"log"
	"testing"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize/convex/lp"
)

func TestLayout(t *testing.T) {
	emptybuffercontainer := func(dim Dimensions) { fmt.Println("emptybuffercontainer", dim) }
	statusline := func(dim Dimensions) { fmt.Println("statusline", dim) }
	linenumbers := func(dim Dimensions) { fmt.Println("linenumbers", dim) }
	buffer := func(dim Dimensions) { fmt.Println("buffer", dim) }

	flex := Flex{
		Dir: Y,
		Items: []FlexItem{
			{Size: 0.5, Box: emptybuffercontainer, Flex: &Flex{
				Dir: X,
				Items: []FlexItem{
					{Size: 0.5, Box: linenumbers, Flex: nil},
					{Size: 0.5, Box: buffer, Flex: nil},
				},
			}},
			{Size: 0.5, Box: statusline, Flex: nil},
		},
	}

	flex.StartLayouting(200, 200)
}

func TestSimplex(t *testing.T) {
	c := []float64{-1, -2, 0, 0}
	A := mat.NewDense(2, 4, []float64{-1, 2, 1, 0, 3, 1, 0, 1})
	b := []float64{4, 9}

	//fmt.Printf("c: %v\n", A.)

	opt, x, err := lp.Simplex(c, A, b, 0, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("opt: %v\n", opt)
	fmt.Printf("x: %v\n", x)
}

func TestSimplexBox(t *testing.T) {
	// objective function: 1b + 9b'
	c := []float64{-1, -1, 0}

	A := mat.NewDense(2, 3, []float64{1, 1, 1,
		0, 1, 1})
	// window width = 1000
	b := []float64{1000, 900}

	//fmt.Printf("c: %v\n", A.)

	opt, x, err := lp.Simplex(c, A, b, 0, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("opt: %v\n", opt)
	fmt.Printf("x: %v\n", x)
}

// Regular
func TestSimplexBoxRegularForm(t *testing.T) {
	// width := 1000
	// c := []float64{1, 9}
	// G := mat.NewDense(2, 2, []float64{1, 9,
	// 								  1, 1,})
	// h := []float64{1000, 1000}
	// A := mat.NewDense(2, 2, []float64{1, 9,
	// 								  1, 1,})
	// b := []float64{1000, 1000}

	// lp.Convert(c, G, h)
}
