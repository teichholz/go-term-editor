package application

import (
	"log"
	Rope "main/brope"
	"main/config"

	"github.com/gdamore/tcell"
)

type Cursor struct {
	x, y int
}

type Window struct {
	Width, Height int
}

type CursorArea struct {
	minX, maxX, minY, maxY int
}

type Application struct {
	file       *string
	rope       Rope.Rope
	cursor     *Cursor
	config     *config.Config
	BufferArea CursorArea
	window     *Window
	screen     tcell.Screen

	log *log.Logger
}