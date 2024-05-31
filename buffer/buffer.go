package buffer

import (
	"bufio"
	"io"
	"log"
	BRope "main/brope"
	"os"
)

type file string
type Buffer struct {
  Open map[string]BRope.Rope

  log *log.Logger
}

func NewBuffer(log *log.Logger) *Buffer {
  return &Buffer{
    Open: make(map[string]BRope.Rope),
    log: log,
  }
}

func (b *Buffer) OpenFile(file string) error {
  rope, err := Read(file)

  if err != nil {
    return err  
  }
  
  b.Open[file] = rope

  return nil
}

func Read(path string) (BRope.Rope, error) {
	file, err := os.Open(path)
	if err != nil {
		return BRope.EmptyRope(), err
	}
	defer file.Close()

	var dest BRope.Rope
	src := bufio.NewReader(file)
	_, err = io.Copy(&dest, src)

	if err != nil {
		return BRope.EmptyRope(), err
	}

	return dest, nil
}

func Write(path string, buffer io.Reader) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, buffer)

	if err != nil {
		return err
	}

	return nil
}
