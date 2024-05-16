package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	BRope "main/brope"
	"main/commands"
	"main/config"
	Files "main/files"
	"main/layout"
	. "main/layout"
	"os"

	"github.com/gdamore/tcell/v2"
)

type Cursor struct {
	Origin
	saved Origin
}

// TODO maybe store Cursor inside each Area, so that we can easily have more areas + input sinks
type CommandArea struct {
	min, max Origin
}

type Window struct {
	Box
}

type Box struct {
	width, height int
}

type Origin struct {
	x, y int
}

type BufferArea struct {
	minX, maxX, minY, maxY int
}

type InputSink func(tcell.Event)
type sink int

const (
	bufferArea sink = iota
	commandArea
)

type Application struct {
	// TODO refactor this into multiple buffers
	file   string
	rope   BRope.Rope
	cursor *Cursor
	config *config.Config

	// TODO refactor this into area type
	currentCommand string
	commandArea    CommandArea
	commands       *commands.Commands

	// TODO refactor this into sink type
	currentSink      sink
	currentInputSink InputSink

	bufferArea BufferArea
	window     *Window
	screen     tcell.Screen

	isAlive bool

	log *log.Logger
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
	cursor := app.cursor
	window := app.window
	s := app.screen

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
			app.cursor.saved = app.cursor.Origin
			app.cursor.Origin = app.commandArea.min
			app.currentInputSink = app.handleInputCommandArea
			app.currentSink = commandArea
		} else if ev.Key() == tcell.KeyRune {
			x, y := cursor.x-app.bufferArea.minX, cursor.y-app.bufferArea.minY
			app.log.Printf("Inserting '%c' into rope '%v' at Cursor (x=%v, y=%v)", ev.Rune(), app.rope.String(), x, y)
			app.rope = app.rope.InsertChar(y, x, ev.Rune())
			cursor.x++
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			x, y := cursor.x-app.bufferArea.minX, cursor.y-app.bufferArea.minY
			app.rope = app.rope.DeleteAt(y, x)
			cursor.x--
			app.log.Printf("Deleting character. Rope is now:\n '%v'", app.rope.String())
		} else if ev.Key() == tcell.KeyEnter {
			x, y := cursor.x-app.bufferArea.minX, cursor.y-app.bufferArea.minY
			app.log.Printf("Inserting '\\n' into rope '%v' at Cursor (x=%v, y=%v)", app.rope.String(), x, y)
			app.rope = app.rope.InsertChar(y, x, '\n')
			cursor.x = app.bufferArea.minX
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
	cursor := app.cursor
	window := app.window
	s := app.screen

	// TODO we need to be able to change the current command at any position via the cursor
	switch ev := ev.(type) {
	case *tcell.EventResize:
		window.update(ev.Size())
		s.Sync()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			app.currentInputSink = app.handleInputBufferArea
			app.currentCommand = ""
			app.currentSink = bufferArea
			app.cursor.Origin = app.cursor.saved
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
			app.currentCommand = app.currentCommand[:len(app.currentCommand)-1]
			cursor.x--
		} else if ev.Key() == tcell.KeyEnter {
			app.commands.Exec(app.currentCommand)
			app.currentCommand = ""
			app.currentInputSink = app.handleInputBufferArea
			app.cursor.Origin = app.cursor.saved
			app.currentSink = bufferArea
		}
	}
}

func (app *Application) quit(s tcell.Screen) {
	maybePanic := recover()
	s.Fini()

	file := app.file
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

	if app.currentSink == bufferArea {
		for i := ymin; i < ymax; i++ {
			drawText(s, xmin, i, xmax, i, DefaultStyle, " ")
		}

		pad := xmax - xmin

		// TODO check this vaule
		//app.log.Printf("Config: %v", app.config)
		if !app.config.EditorConfig.RelativeLineNumbers {
			var lineCount int
			if app.rope.LineCount() == 0 {
				lineCount = 1
			} else {
				lineCount = app.rope.LineCount()
			}
			for i := 0; i < lineCount; i++ {
				drawText(s, xmin, i, xmax, i, DefaultStyle, fmt.Sprintf("%*v", pad, i))
			}
		} else {
			var lineCount int
			if app.rope.LineCount() == 0 {
				lineCount = 1
			} else {
				lineCount = app.rope.LineCount()
			}
			for top := 0; top < app.cursor.y; top++ {
				drawText(s, xmin, top, xmax, top, LightStyle, fmt.Sprintf("%*v", pad, app.cursor.y-top))
			}
			drawText(s, xmin, app.cursor.y, xmax, app.cursor.y, DefaultStyle, fmt.Sprintf("%*v", pad, app.cursor.y))
			for bottom := app.cursor.y + 1; bottom < lineCount; bottom++ {
				drawText(s, xmin, bottom, xmax, bottom, LightStyle, fmt.Sprintf("%*v", pad, bottom-app.cursor.y))
			}
		}
	}
}
func (app *Application) bufferBox(dims layout.Dimensions) {
	s := app.screen
	xmin, ymin, xmax, ymax := dims.Origin.X, dims.Origin.Y, dims.Origin.X+dims.Width, dims.Origin.Y+dims.Height

	if app.currentSink == bufferArea {
		drawRunes(s, xmin, ymin, xmax, ymax, DefaultStyle, app.rope.Runes())

		app.bufferArea = BufferArea{xmin, xmax, ymin, ymax}
		app.clampBufferAreaCursor()
	}
}

func (app *Application) clampBufferAreaCursor() {
	cursor := app.cursor

	// move cursor to end of previous line
	if cursor.x < app.bufferArea.minX && cursor.y > app.bufferArea.minY {
		cursor.y--
		cursor.x = app.bufferArea.minX + app.rope.LastCharInRow(cursor.y) + 1
	}

	// keep cursor in left and right bounds
	cursor.x = max(cursor.x, app.bufferArea.minX)
	cursor.x = min(cursor.x, app.bufferArea.maxX)

	// keep cursor in top and bottom bounds
	cursor.y = max(cursor.y, app.bufferArea.minY)
	cursor.y = min(cursor.y, app.bufferArea.maxY)
}

func (app *Application) statusLineBox(dims layout.Dimensions) {
	s := app.screen
	xmin, ymin, xmax, ymax := dims.Origin.X, dims.Origin.Y, dims.Origin.X+dims.Width, dims.Origin.Y+dims.Height
	// app.log.Printf("Drawing status line box at (%v, %v) to (%v, %v)", xmin, ymin, xmax, ymax)
	drawBox(s, xmin, ymin, xmax-1, ymax-1, DefaultStyle)
	drawText(s, xmax-7, ymin+1, xmax-1, ymax-1, DefaultStyle, "Normal")

	prefix := "Cmd: "
	offset := len(prefix) + 1
	if app.currentSink == commandArea {
		// Cursor needs to consider 'Cmd: ' prefx
		drawText(s, xmin+1, ymin+1, xmax-1, ymax-1, DefaultStyle, prefix+app.currentCommand)
	}
	app.commandArea = CommandArea{Origin{xmin + offset, ymin + 1}, Origin{xmax - 1, ymax - 7}}
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
	window := &Window{Box{width, height}}
	cursor := &Cursor{Origin{x: 0, y: 0}, Origin{0, 0}}
	cursorArea := BufferArea{0, window.width - 1, 0, window.height - 1}
	log := NewLogger()
	config := config.NewConfig(log)
	config.Init()
	defer config.Cleanup()
	commands := commands.NewCommands(log)
	application := &Application{
		file:        "",
		cursor:      cursor,
		config:      config,
		commands:    commands,
		rope:        BRope.NewRopeString(""),
		currentSink: bufferArea,
		window:      window,
		bufferArea:  cursorArea,
		screen:      s,
		log:         log,
		isAlive:     true,
	}
	application.currentInputSink = application.handleInputBufferArea

	commands.Register("help", application.helpCmd)
	commands.Register("quit", application.quitCmd)
	commands.Register("write", application.writeCmd)
	commands.Register("read", application.readCmd)
	commands.Register("hsplit", application.hsplitCmd)
	commands.Register("files", application.filesCmd)

	flag.Parse()
	file := flag.Arg(0)

	if file == "" {
		rope := BRope.NewRopeString("")
		application.rope = rope
		log.Print("Started program without any files. Created new rope.")
	} else {
		application.file = file
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

	layout := Column(
		FlexItemBox(EmptyBox, Max(Rel(1)), Row(
			FlexItemBox(application.lineNumberBox, Exact(Abs(3)), nil),
			FlexItemBox(application.bufferBox, Max(Rel(1)), nil),
		)),
		FlexItemBox(application.statusLineBox, Exact(Abs(3)), nil),
	)

	// Event loop
	for application.isAlive {
		window.update(s.Size())
		s.Clear()
		layout.StartLayouting(window.width, window.height)

		s.ShowCursor(application.cursor.x, application.cursor.y)

		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		application.currentInputSink(ev)
	}
}

func (app *Application) helpCmd() {
	app.log.Println("Sadly there is no help yet.")
}

func (app *Application) quitCmd() {
	app.isAlive = false
}

func (app *Application) writeCmd() {
	err := Files.Write(app.file, app.rope)
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