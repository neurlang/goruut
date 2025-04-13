package main

import "os"
import "fmt"
import "github.com/neurlang/classifier/parallel"
import "bufio"
import "math/rand"
import "sync/atomic"

func load(filename string, top int) []string {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()
	var slice []string

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		slice = append(slice, line)
	}

	// Check for any scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	if top > 0 {
		rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
		if len(slice) > top {
			slice = slice[:top]
		}
	}
	return slice
}

func loop(slice []string, group int, do func(string)) {
	size := uint64(len(slice))
	var pos atomic.Uint64
	parallel.ForEach(len(slice), group, func(n int) {
		// Process each column
		column := slice[n]
		// Example: Print the columns
		do(column)
		pos.Add(1)
		progressbar(pos.Load(), size)
	})
}

func emptySpace(space int) string {
	emptySpace := ""
	for i := 0; i < space; i++ {
		emptySpace += " "
	}
	return emptySpace
}
func progressBar(progress, width int) string {
	progressBar := ""
	for i := 0; i < progress; i++ {
		progressBar += "="
	}
	return progressBar
}
func progressbar(pos, max uint64) {
	const progressBarWidth = 40
	if max > 0 {
		progress := int(pos*progressBarWidth/max)
		percent := int(pos*100/max)
		fmt.Printf("\r[%s%s] %d%% ", progressBar(progress, progressBarWidth), emptySpace(progressBarWidth-progress), percent)
	}
}
