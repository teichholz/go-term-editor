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
	File string
	Rope BRope.Rope
}

type Buffers struct {
	Open map[string]*Buffer

	log *log.Logger
}

func NewBuffers(log *log.Logger) *Buffers {
	return &Buffers{
		Open: make(map[string]*Buffer),
		log:  log,
	}
}

// Opens a unique temporary buffer
func (b *Buffers) OpenTemp() (*Buffer, error) {
	temp, err := os.CreateTemp(os.TempDir(), "temp")
	if err != nil {
		return nil, err
	}

	buf := &Buffer{File: temp.Name(), Rope: BRope.EmptyRope()}
	b.Open[temp.Name()] = buf

	return buf, nil
}

func (b *Buffers) OpenFile(file string) (*Buffer, error) {
	rope, err := Read(file)

	if err != nil {
		return nil, err
	}

	buf := &Buffer{File: file, Rope: rope}
	b.Open[file] = buf

	return buf, nil
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

func (b *Buffers) WriteClose(file string) error {
	err := write(file, b.Open[file].Rope)
	if err != nil {
		return err
	}

	return b.Close(file)
}

func (b *Buffers) Write(file string) error {
	err := write(file, b.Open[file].Rope)
	if err != nil {
		return err
	}

	return nil
}

func (b *Buffers) Close(file string) error {
	delete(b.Open, file)
	return nil
}

func write(path string, buffer io.Reader) error {
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
