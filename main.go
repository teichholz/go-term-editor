package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	Buffer "main/buffer"
	"main/commands"
	"main/config"
	"main/layout"
	. "main/layout"
	"os"

	"github.com/gdamore/tcell/v2"
)

type Cursor struct {
	Origin
	saved Origin
}

type Window struct {
	width, height int
}

type Box struct {
	min, max Origin
}

type Origin struct {
	x, y int
}


type InputSink func(tcell.Event)
type InputAreaType int


const (
	bufferArea InputAreaType = iota
	commandArea
)

// Buffer (in memory of file)
// Window is viewport on buffer. Eeach window is its own InputArea. 
// Tab page is collection of windows

// TODO flesh out

type Application struct {
	// TODO refactor this into multiple buffers, buffers need to factor in an open tab, and the input area
  currentBuffer *Buffer.Buffer
	buffers *Buffer.Buffers

  // editor configuration
	config *config.Config

  // command to execute is build up and stored here
	currentCommand string
  // map of all possible commands and their implementations
	commands       *commands.Commands

	activeInputArea *InputArea
	inputAreas      map[InputAreaType]*InputArea

  currentArea Area
  commandArea *CommandArea 

  // current terminal window size. Gets updated by the value tcell sends us
	window *Window
  // tcell state holder
	screen tcell.Screen

  // whether the application is still running
	isAlive bool

	log *log.Logger
}

type WindowArea struct {
	box    *Box
	cursor *Cursor
}

type InputArea struct {
	typ  InputAreaType
	area WindowArea
	sink InputSink
}

type BufferWindow struct {
  inputArea *InputArea
  buffer *Buffer.Buffer
}

func (win *BufferWindow) send(ev tcell.Event) { 
  win.inputArea.sink(ev)
}

type CommandArea struct {
  inputArea *InputArea
}

func (cmda *CommandArea) send(ev tcell.Event) { 
  cmda.inputArea.sink(ev)
}


type Area interface {
  // Areas can be "active" in which case events are send to them
  send(tcell.Event) 
}

func (app *Application) switchInputArea(inputArea InputAreaType) {
	app.activeInputArea = app.inputAreas[inputArea]
}

func (win *Window) update(width, height int) {
	win.width, win.height = width, height
}

func (app *Application) broadcastInputSink(sinks ...InputSink) InputSink {
	return func(ev tcell.Event) {
		for _, sink := range sinks {
			sink(ev)
		}
	}
}

func (app *Application) handleInputBufferArea(ev tcell.Event) {
	cursor := app.activeInputArea.area.cursor
	window := app.window
	s := app.screen
	box := app.activeInputArea.area.box

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
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == ':' {
			// switch into command mode
			app.activeInputArea = app.inputAreas[commandArea]
		} else if ev.Key() == tcell.KeyRune {
			x, y := cursor.x-box.min.x, cursor.y-box.min.y
			app.log.Printf("Inserting '%c' into rope '%v' at Cursor (x=%v, y=%v)", ev.Rune(), app.currentBuffer.Rope.String(), x, y)
			app.currentBuffer.Rope = app.currentBuffer.Rope.InsertChar(y, x, ev.Rune())
			cursor.x++
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			minx, miny := app.activeInputArea.area.box.min.x, app.activeInputArea.area.box.min.y
			x, y := cursor.x-minx, cursor.y-miny
			app.currentBuffer.Rope = app.currentBuffer.Rope.DeleteAt(y, x)
			cursor.x--
			app.log.Printf("Deleting character. Rope is now:\n '%v'", app.currentBuffer.Rope.String())
		} else if ev.Key() == tcell.KeyEnter {
			x, y := cursor.x-box.min.x, cursor.y-box.min.y
			app.log.Printf("Inserting '\\n' into rope '%v' at Cursor (x=%v, y=%v)", app.currentBuffer.Rope.String(), x, y)
			app.currentBuffer.Rope = app.currentBuffer.Rope.InsertChar(y, x, '\n')
			cursor.x = app.activeInputArea.area.box.min.x
			cursor.y++
		}
	case *tcell.EventMouse:
		x, y := ev.Position()
		if ev.Buttons() == tcell.Button1 {
			cursor.x, cursor.y = x, y
		}
	}
}

func (app *Application) handleInputCommandArea(ev tcell.Event) {
	cursor := app.activeInputArea.area.cursor
	window := app.window
	s := app.screen

	// TODO we need to be able to change the current command at any position via the cursor
	switch ev := ev.(type) {
	case *tcell.EventResize:
		window.update(ev.Size())
		s.Sync()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			app.currentCommand = ""
			app.activeInputArea = app.inputAreas[bufferArea]
			// invalidate cursor, causing them to be clamped again next time
			cursor.x, cursor.y = -1, -1
			app.activeInputArea = app.inputAreas[bufferArea]
		} else if ev.Key() == tcell.KeyCtrlC {
			app.quit(s)
		} else if ev.Key() == tcell.KeyLeft {
			cursor.x--
		} else if ev.Key() == tcell.KeyRight {
			cursor.x++
		} else if ev.Key() == tcell.KeyCtrlL {
			s.Sync()
		} else if ev.Key() == tcell.KeyRune {
			rune := ev.Rune()
			app.currentCommand += string(rune)
			cursor.x++
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			if len(app.currentCommand) > 0 {
				app.currentCommand = app.currentCommand[:len(app.currentCommand)-1]
			}
			cursor.x--
		} else if ev.Key() == tcell.KeyEnter {
			app.commands.Exec(app.currentCommand)
			// invalidate cursor, causing them to be clamped again next time
			cursor.x, cursor.y = -1, -1
			app.currentCommand = ""
			app.activeInputArea = app.inputAreas[bufferArea]
		}
	}
}

func (app *Application) quit(s tcell.Screen) {
	maybePanic := recover()
	s.Fini()

	err := app.buffers.WriteClose(app.currentBuffer.File)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	app.log.Printf("Wrote rope to file %v:\n'%v'", app.currentBuffer.File, app.currentBuffer.Rope)

	if maybePanic != nil {
		panic(maybePanic)
	} else {
		os.Exit(0)
	}
}

var nFlag = flag.Int("n", 1234, "help message for flag n")
// TODO make these flags work
var oFlag = flag.String("o", "", "Files to open horizontally split on startup")
var OFlag = flag.String("O", "", "Files to open vertically split on startup")

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

	if app.activeInputArea.typ == bufferArea {
		cursor := app.activeInputArea.area.cursor
		pad := xmax - xmin

		for i := ymin; i < ymax; i++ {
			drawText(s, xmin, i, xmax, i, DefaultStyle, " ")
		}

		if !app.config.EditorConfig.RelativeLineNumbers {
			var lineCount int
			if app.currentBuffer.Rope.LineCount() == 0 {
				lineCount = 1
			} else {
				lineCount = app.currentBuffer.Rope.LineCount()
			}
			for i := 0; i < lineCount; i++ {
				drawText(s, xmin, i, xmax, i, DefaultStyle, fmt.Sprintf("%*v", pad, i))
			}
		} else {
      // TODO relative liner numbers are not working properly
			var lineCount int
			if app.currentBuffer.Rope.LineCount() == 0 {
				lineCount = 1
			} else {
				lineCount = app.currentBuffer.Rope.LineCount()
			}
			if lineCount == 1 {
				drawText(s, xmin, cursor.y, xmax, cursor.y, DefaultStyle, fmt.Sprintf("%*v", pad, 0))
			} else {
				for top := 0; top < cursor.y; top++ {
					drawText(s, xmin, top, xmax, top, LightStyle, fmt.Sprintf("%*v", pad, cursor.y-top))
				}
				drawText(s, xmin, cursor.y, xmax, cursor.y, DefaultStyle, fmt.Sprintf("%*v", pad, cursor.y))
				for bottom := cursor.y + 1; bottom < lineCount; bottom++ {
					drawText(s, xmin, bottom, xmax, bottom, LightStyle, fmt.Sprintf("%*v", pad, bottom-cursor.y))
				}
			}
		}
	}
}

func (app *Application) bufferBox(dims layout.Dimensions) {
	app.log.Printf("Drawing buffer box")
	s := app.screen
	xmin, ymin, xmax, ymax := dims.Origin.X, dims.Origin.Y, dims.Origin.X+dims.Width, dims.Origin.Y+dims.Height

	if app.activeInputArea.typ == bufferArea {
		drawRunes(s, xmin, ymin, xmax, ymax, DefaultStyle, app.currentBuffer.Rope.Runes())
		box := Box{Origin{xmin, ymin}, Origin{xmax, ymax}}
		inputArea := app.inputAreas[bufferArea]
		inputArea.area.box = &box
	}
}

func (app *Application) clampAreaCursor() {
	cursor := app.activeInputArea.area.cursor
	box := app.activeInputArea.area.box
	minx, miny := box.min.x, box.min.y
	maxx, maxy := box.max.x, box.max.y

	// move cursor to end of previous line
	if app.activeInputArea.typ == bufferArea {
		if cursor.x < minx && cursor.y > miny {
			cursor.y--
			cursor.x = minx + app.currentBuffer.Rope.LastCharInRow(cursor.y) + 1
		}
	}

	// keep cursor in left and right bounds
	cursor.x = max(cursor.x, minx)
	cursor.x = min(cursor.x, maxx)

	// keep cursor in top and bottom bounds
	cursor.y = max(cursor.y, miny)
	cursor.y = min(cursor.y, maxy)

  // clamp cursor to last line of buffer, should be in range [0, app.currentBuffer.Rope.LineCount())
  if app.activeInputArea.typ == bufferArea {
    cursor.y = min(cursor.y, app.currentBuffer.Rope.LineCount()-1)
  }

	// app.log.Printf("Clamped cursor to (%v, %v)", cursor.x, cursor.y)
}


func (app *Application) statusLineBox(dims layout.Dimensions) {
	s := app.screen
	xmin, ymin, xmax, ymax := dims.Origin.X, dims.Origin.Y, dims.Origin.X+dims.Width, dims.Origin.Y+dims.Height
	// app.log.Printf("Drawing status line box at (%v, %v) to (%v, %v)", xmin, ymin, xmax, ymax)
	drawBox(s, xmin, ymin, xmax-1, ymax-1, DefaultStyle)
	drawText(s, xmax-7, ymin+1, xmax-1, ymax-1, DefaultStyle, "Normal")

	prefix := "Cmd: "
	offset := len(prefix) + 1
	if app.activeInputArea.typ == commandArea {
		// Cursor needs to consider 'Cmd: ' prefx
		drawText(s, xmin+1, ymin+1, xmax-1, ymax-1, DefaultStyle, prefix+app.currentCommand)
		box := Box{Origin{xmin + offset, ymin + 1}, Origin{xmax - 7, ymax - 1}}
		inputArea := app.inputAreas[commandArea]
		inputArea.area.box = &box
	}
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
	log := NewLogger()
	config := config.NewConfig(log)
	config.Init()
	defer config.Cleanup()
	commands := commands.NewCommands(log)
	app := &Application{
		config:     config,
		buffers:     Buffer.NewBuffers(log),
		commands:   commands,
		inputAreas: make(map[InputAreaType]*InputArea, 10),
		window:     window,
		screen:     s,
		log:        log,
		isAlive:    true,
	}
 
	bufferInputArea := &InputArea{
		typ:  bufferArea,
		area: WindowArea{&Box{Origin{0, 0}, Origin{window.width - 1, window.height - 1}}, &Cursor{Origin{0, 0}, Origin{0, 0}}},
		sink: app.handleInputBufferArea,
	}

	commandInputArea := &InputArea{
		typ:  commandArea,
		area: WindowArea{&Box{Origin{0, 0}, Origin{window.width - 1, window.height - 1}}, &Cursor{Origin{0, 0}, Origin{0, 0}}},
		sink: app.handleInputCommandArea,
	}

  ca := &CommandArea{
    inputArea: commandInputArea,
  }
  app.commandArea = ca

	app.inputAreas[bufferArea] = bufferInputArea
	app.inputAreas[commandArea] = commandInputArea
	app.activeInputArea = bufferInputArea

	commands.Register("help", app.helpCmd)
	commands.Register("quit", app.quitCmd)
	commands.Register("exit", app.quitCmd)
	commands.Register("write", app.writeCmd)
	commands.Register("read", app.readCmd)
	commands.Register("hsplit", app.hsplitCmd)
	commands.Register("files", app.filesCmd)

	flag.Parse()
	file := flag.Arg(0)

	if file == "" {
    var erro error
    app.currentBuffer, erro = app.buffers.OpenTemp() 

    if erro != nil {
      log.Fatalf("Could not open file: %v", err)
    }

    bw := &BufferWindow{
      inputArea: bufferInputArea,
      buffer: app.currentBuffer,
    }
    app.currentArea = bw

		log.Print("Started program without any files. Created new rope.")
	} else {
    var erro error
    app.currentBuffer, erro = app.buffers.OpenFile(file)

    if erro != nil {
      log.Fatalf("Could not open file: %v", err)
    }

    bw := &BufferWindow{
      inputArea: bufferInputArea,
      buffer: app.currentBuffer,
    }
    app.currentArea = bw
    
		log.Printf("Read rope from file %v:\n'%v'", file, app.currentBuffer.Rope)
	}

	// You have to catch panics in a defer, clean up, and
	// re-raise them - otherwise your application can
	// die without leaving any diagnostic trace.
	defer app.quit(s)

	layouter := layout.NewLayouter(log)

	layout := Column(
		FlexItemBox(EmptyBox, Max(Rel(1)), Row(
			FlexItemBox(app.lineNumberBox, Exact(Abs(3)), nil),
			FlexItemBox(app.bufferBox, Max(Rel(1)), nil),
		)),
		FlexItemBox(app.statusLineBox, Exact(Abs(3)), nil),
	)

	// Event loop
	for app.isAlive {
		window.update(s.Size())
		s.Clear()
		layouter.StartLayouting(layout, window.width, window.height)
		app.clampAreaCursor()

		cx, cy := app.activeInputArea.area.cursor.x, app.activeInputArea.area.cursor.y
		s.ShowCursor(cx, cy)

		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		app.activeInputArea.sink(ev)
	}
}

func (app *Application) helpCmd() {
	app.log.Println("Sadly there is no help yet.")
}

func (app *Application) quitCmd() {
	app.isAlive = false
}

func (app *Application) writeCmd() {
	err := app.buffers.Write(app.currentBuffer.File)
	if err != nil {
		app.log.Fatalf("Could not write buffer content to file: %v", err)
	}
}

func (app *Application) readCmd() {
	app.log.Printf("read command not implemented yet.")
}

func (app *Application) hsplitCmd() {
	app.log.Printf("hsplit command not implemented yet.")
}

func (app *Application) filesCmd() {
	app.log.Printf("files command not implemented yet.")
}
