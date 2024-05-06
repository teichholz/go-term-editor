package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	BRope "main/brope"
	Files "main/files"
	"main/layout"
	"os"

	"github.com/gdamore/tcell/v2"
)

type Cursor struct {
	x, y int
}

type Window struct {
	width, height int
}

type CursorArea struct {
	minX, maxX, minY, maxY int
}

type Application struct {
	file       *string
	rope       BRope.Rope
	cursor     *Cursor
	BufferArea CursorArea
	window     *Window
	screen     tcell.Screen

	log *log.Logger
}

func (win *Window) update(width, height int) {
	win.width, win.height = width, height
}

func (app *Application) clampCursor() {
	cursor := app.cursor

	// move cursor to end of previous line
	if cursor.x < app.BufferArea.minX && cursor.y > app.BufferArea.minY {
		cursor.y--
		cursor.x = app.BufferArea.minX + app.rope.LastCharInRow(cursor.y) + 1
	}

	// keep cursor in left and right bounds
	cursor.x = max(cursor.x, app.BufferArea.minX)
	cursor.x = min(cursor.x, app.BufferArea.maxX)

	// keep cursor in top and bottom bounds
	cursor.y = max(cursor.y, app.BufferArea.minY)
	cursor.y = min(cursor.y, app.BufferArea.maxY)
}

func (app *Application) handleInput(s tcell.Screen, ev tcell.Event) {
	cursor := app.cursor
	window := app.window

	switch ev := ev.(type) {
	case *tcell.EventResize:
		window.update(ev.Size())
		s.Sync()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
			app.quit(s)
		} else if ev.Key() == tcell.KeyUp {
			cursor.y = cursor.y - 1
		} else if ev.Key() == tcell.KeyDown {
			cursor.y++
		} else if ev.Key() == tcell.KeyLeft {
			cursor.x--
		} else if ev.Key() == tcell.KeyRight {
			cursor.x++
		} else if ev.Key() == tcell.KeyCtrlL {
			s.Sync()
		} else if ev.Key() == tcell.KeyRune {
			x, y := cursor.x-app.BufferArea.minX, cursor.y-app.BufferArea.minY
			app.log.Printf("Inserting '%c' into rope '%v' at Cursor (x=%v, y=%v)", ev.Rune(), app.rope.String(), x, y)
			app.rope = app.rope.InsertChar(y, x, ev.Rune())
			cursor.x++
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			x, y := cursor.x-app.BufferArea.minX, cursor.y-app.BufferArea.minY
			app.rope = app.rope.DeleteAt(y, x)
			cursor.x--
			app.log.Printf("Deleting character. Rope is now:\n '%v'", app.rope.String())
		} else if ev.Key() == tcell.KeyEnter {
			x, y := cursor.x-app.BufferArea.minX, cursor.y-app.BufferArea.minY
			app.log.Printf("Inserting '\\n' into rope '%v' at Cursor (x=%v, y=%v)", app.rope.String(), x, y)
			app.rope = app.rope.InsertChar(y, x, '\n')
			cursor.x = app.BufferArea.minX
			cursor.y++
		}
	case *tcell.EventMouse:
		x, y := ev.Position()
		if ev.Buttons() == tcell.Button1 {
			cursor.x, cursor.y = x, y
		}
	}
}

func (app *Application) quit(s tcell.Screen) {
	maybePanic := recover()
	s.Fini()

	file := *app.file
	err := Files.Write(file, app.rope)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	app.log.Printf("Wrote rope to file %v:\n'%v'", file, app.rope)

	if maybePanic != nil {
		panic(maybePanic)
	} else {
		os.Exit(0)
	}
}

var nFlag = flag.Int("n", 1234, "help message for flag n")

func NewLogger() *log.Logger {
	// Open a file for logging
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	// no need to defer I guess

	// Create a multi writer
	multi := io.MultiWriter(file)

	// Set the output for the standard logger
	return log.New(multi, "", log.LstdFlags|log.Lshortfile)
}

func (app *Application) lineNumberBox(dims layout.Dimensions) {
	s := app.screen
	xmin, ymin, xmax, ymax := dims.Origin.X, dims.Origin.Y, dims.Origin.X+dims.Width, dims.Origin.Y+dims.Height
	for i := ymin; i < ymax; i++ {
		drawText(s, xmin, i, xmax, i, DefaultStyle, " ")
	}

	// draw new correct ones
	var lineCount int
	if app.rope.LineCount() == 0 {
		lineCount = 1
	} else {
		lineCount = app.rope.LineCount()
	}
	for i := 0; i < lineCount; i++ {
		pad := xmax - xmin
		drawText(s, xmin, i, xmax, i, DefaultStyle, fmt.Sprintf("%*v", pad, i))
	}
}
func (app *Application) bufferBox(dims layout.Dimensions) {
	s := app.screen
	xmin, ymin, xmax, ymax := dims.Origin.X, dims.Origin.Y, dims.Origin.X+dims.Width, dims.Origin.Y+dims.Height
	app.BufferArea = CursorArea{xmin, xmax, ymin, ymax}
	drawText(s, xmin, ymin, xmax, ymax, DefaultStyle, app.rope.String())
}
func (app *Application) statusLineBox(dims layout.Dimensions) {
	s := app.screen
	xmin, ymin, xmax, ymax := dims.Origin.X, dims.Origin.Y, dims.Origin.X+dims.Width, dims.Origin.Y+dims.Height
	drawBox(s, xmin, ymin, xmax-1, ymax, DefaultStyle, "Normal Mode")
}

func main() {
	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(DefaultStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	width, height := s.Size()
	window := &Window{width, height}
	cursor := &Cursor{x: 0, y: 0}
	cursorArea := CursorArea{0, window.width - 1, 0, window.height - 1}
	log := NewLogger()
	application := &Application{
		file:       nil,
		cursor:     cursor,
		rope:       BRope.NewRopeString(""),
		window:     window,
		BufferArea: cursorArea,
		screen:     s,
		log:        log,
	}

	flag.Parse()
	file := flag.Arg(0)

	if file == "" {
		rope := BRope.NewRopeString("")
		application.rope = rope
		log.Print("Started program without any files. Created new rope.")
	} else {
		application.file = &file
		var err error
		rope, err := Files.Read(file)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		log.Printf("Read rope from file %v:\n'%v'", file, rope)
		application.rope = rope
	}

	// You have to catch panics in a defer, clean up, and
	// re-raise them - otherwise your application can
	// die without leaving any diagnostic trace.
	defer application.quit(s)

	flex := layout.Flex{
		Dir: layout.Y,
		Items: []layout.FlexItem{
			{Size: 0.95, Box: layout.EmptyBox, Flex: &layout.Flex{
				Dir: layout.X,
				Items: []layout.FlexItem{
					{Size: 0.05, Box: application.lineNumberBox, Flex: nil},
					{Size: 0.95, Box: application.bufferBox, Flex: nil},
				},
			}},
			{Size: 0.05, Box: application.statusLineBox, Flex: nil},
		},
	}

	// Event loop
	for {
		window.update(s.Size())
		s.Clear()
		flex.StartLayouting(window.width, window.height)

		// Clamp cursor position and move move cursor up and down if necessary
		application.clampCursor()
		s.ShowCursor(cursor.x, cursor.y)

		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		application.handleInput(s, ev)
	}
}
