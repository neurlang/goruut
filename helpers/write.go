package helpers

import (
	"errors"
	"io"
)

var errNotFullyWritten = errors.New("all bytes not written")

// Write writes data to the specified writer.
func Write(w io.Writer, data []byte) error {
	cnt, err := w.Write(data)
	if err != nil {
		return err
	}
	if cnt != len(data) {
		return errNotFullyWritten
	}
	return nil
}
