package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// TSVWriter struct to manage file and writer state
type TSVWriter struct {
	file   *os.File
	writer *bufio.Writer
	mut    sync.Mutex
}

// Open method to create and open the TSV file
func (t *TSVWriter) Open(fileName string) error {
	var err error
	// Create or open the TSV file
	t.file, err = os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}

	// Create a TSV writer
	t.writer = bufio.NewWriter(t.file)

	if t.writer == nil {
		return fmt.Errorf("no writer")
	}

	return nil
}

// AddRow method to add a row to the TSV file
func (t *TSVWriter) AddRow(record []string) error {
	t.mut.Lock()
	defer t.mut.Unlock()

	if t.writer == nil {
		return nil
	}

	line := strings.Join(record, "\t")
	_, err := t.writer.WriteString(line + "\n")
	if err != nil {
		return fmt.Errorf("error writing row: %v", err)
	}
	err = t.writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing writer after row: %v", err)
	}
	return nil
}

// Close method to close the TSV file
func (t *TSVWriter) Close() error {
	if t.writer != nil {
		err := t.writer.Flush()
		if err != nil {
			return fmt.Errorf("error flushing writer at end: %v", err)
		}
	}
	if t.file != nil {
		if err := t.file.Close(); err != nil {
			return fmt.Errorf("error closing file: %v", err)
		}
	}
	return nil
}
