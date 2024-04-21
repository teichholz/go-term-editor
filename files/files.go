package Files

import (
	"bufio"
	"io"
	"main/rope"
	"os"
)

func Read(path string) (*rope.Rope, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := bufio.NewReader(file)
	writer := rope.Writer()
	_, err = io.Copy(writer, reader)

	if (err != nil) {
		return nil, err
	}

	return writer.Rope(), nil
}

func Write(path string, buffer io.Reader) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
    if err != nil {
        return err
    }
    defer file.Close()

	_, err = io.Copy(file, buffer)

	if (err != nil) {
		return err
	}

	return nil
}