package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	Buffer "main/buffer"
	Files "main/files"
	"main/rope"
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
	buffer     Buffer.ExtendedBuffer
	cursor     *Cursor
	BufferArea CursorArea
	window     *Window

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
		cursor.x = app.BufferArea.minX + app.buffer.LastNonWhitespaceChar(cursor.y) + 1
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
			x, y := cursor.x - app.BufferArea.minX, cursor.y - app.BufferArea.minY
			app.log.Printf("Inserting '%c' into rope '%v' at Cursor (x=%v, y=%v)", ev.Rune(), app.buffer.String(), x, y);
			app.buffer.Buffer = app.buffer.InsertChar(y, x, ev.Rune());
			cursor.x++
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			x, y := cursor.x - app.BufferArea.minX, cursor.y - app.BufferArea.minY
			app.buffer.Buffer = app.buffer.DeleteAt(y, x)
			cursor.x--
			app.log.Printf("Deleting character. Rope is now:\n '%v'", app.buffer.String());
		} else if ev.Key() == tcell.KeyEnter {
			x, y := cursor.x - app.BufferArea.minX, cursor.y - app.BufferArea.minY
			app.log.Printf("Inserting '\\n' into rope '%v' at Cursor (x=%v, y=%v)", app.buffer.String(), x, y);
			app.buffer.Buffer = app.buffer.InsertChar(y, x, '\n');
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

func (app *Application) drawStatusBar(s tcell.Screen) CursorArea {
	window := app.window
	drawBox(s, 0, window.height-3, window.width-1, window.height-1, DefaultStyle, "Normal Mode")
	return CursorArea{app.BufferArea.minX, app.BufferArea.maxX, app.BufferArea.minY, app.BufferArea.maxY - 3}
}

func (app *Application) drawLineNumbers(s tcell.Screen) CursorArea {
	// erase existing line numbers
	for i := 0; i < app.BufferArea.maxY; i++ {
		drawText(s, 0, i, 4, i, DefaultStyle, " ")
	}

	// draw new correct ones
	var lineCount int
	if app.buffer.LineCount() == 0 {
		lineCount = 1
	} else {
		lineCount = app.buffer.LineCount()
	}
	for i := 0; i < lineCount; i++ {
		drawText(s, 0, i, 4, i, DefaultStyle, fmt.Sprintf("%v", i))
	}
	return CursorArea{app.BufferArea.minX + 4, app.BufferArea.maxX, app.BufferArea.minY, app.BufferArea.maxY}
}

func (app *Application) quit(s tcell.Screen) {
	maybePanic := recover()
	s.Fini()

	file := *app.file
	err := Files.Write(file, app.buffer.Buffer)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	app.log.Printf("Wrote rope to file %v:\n'%v'", file, app.buffer)

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

func (app *Application) eraseBuffer(s tcell.Screen) {
	for i := app.BufferArea.minY; i < app.BufferArea.maxY; i++ {
		for j := app.BufferArea.minX; j < app.BufferArea.maxX; j++ {
			s.SetContent(j, i, ' ', nil, DefaultStyle)
		}
	}
}

func (app *Application) redrawBuffer(s tcell.Screen) {
	app.eraseBuffer(s)
	drawText(s, app.BufferArea.minX, app.BufferArea.minY, app.BufferArea.maxX, app.BufferArea.maxY, DefaultStyle, app.buffer.String())
	s.Show()
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
		buffer:     Buffer.ExtendedBuffer{Buffer: rope.NewString("")},
		window:     window,
		BufferArea: cursorArea,
		log:        log,
	}

	flag.Parse()
	file := flag.Arg(0)

	if file == "" {
		rope := rope.NewString("")
		application.buffer.Buffer = &rope
		log.Print("Started program without any files. Created new rope.")
	} else {
		application.file = &file
		var err error
		rope, err := Files.Read(file)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		log.Printf("Read rope from file %v:\n'%v'", file, rope)
		application.buffer.Buffer = rope
	}


	// You have to catch panics in a defer, clean up, and
	// re-raise them - otherwise your application can
	// die without leaving any diagnostic trace.
	defer application.quit(s)

	// Event loop
	for {
		window.update(s.Size())
		application.BufferArea = cursorArea
		application.BufferArea = application.drawStatusBar(s)
		application.BufferArea = application.drawLineNumbers(s)
		application.redrawBuffer(s)

		// Clamp cursor position and move move cursor up and down if necessary
		application.clampCursor()
		s.ShowCursor(cursor.x, cursor.y)

		// drawText(s, 2, window.height-7, window.width-1, window.height-1, DefaultStyle, application.buffer.String())
		// log.Printf("Rope: %v", application.rope)

		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		application.handleInput(s, ev)
	}
}
