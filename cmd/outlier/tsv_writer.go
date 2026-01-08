package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

// TSVWriter struct to manage file and writer state
type TSVWriter struct {
	file   *os.File
	writer *csv.Writer
	mutex  sync.Mutex
}

// Open method to create and open the TSV file
func (t *TSVWriter) Open(fileName string, headers []string) error {
	var err error
	// Create or open the TSV file
	t.file, err = os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}

	// Create a TSV writer
	t.writer = csv.NewWriter(t.file)
	t.writer.Comma = '\t' // Set the delimiter to tab

	// Write the header row if provided
	if len(headers) > 0 {
		if err := t.writer.Write(headers); err != nil {
			return fmt.Errorf("error writing header: %v", err)
		}
		t.writer.Flush()
		if err := t.writer.Error(); err != nil {
			return fmt.Errorf("error flushing writer after header: %v", err)
		}
	}

	return nil
}

// AddRow method to add a row to the TSV file
func (t *TSVWriter) AddRow(row []string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.writer == nil {
		return fmt.Errorf("TSVWriter not opened")
	}
	if err := t.writer.Write(row); err != nil {
		return fmt.Errorf("error writing row: %v", err)
	}
	t.writer.Flush()
	if err := t.writer.Error(); err != nil {
		return fmt.Errorf("error flushing writer after row: %v", err)
	}
	return nil
}

// Close method to close the TSV file
func (t *TSVWriter) Close() error {
	if t.writer != nil {
		t.writer.Flush()
		if err := t.writer.Error(); err != nil {
			return fmt.Errorf("error flushing writer on close: %v", err)
		}
	}
	if t.file != nil {
		if err := t.file.Close(); err != nil {
			return fmt.Errorf("error closing file: %v", err)
		}
	}
	return nil
}
