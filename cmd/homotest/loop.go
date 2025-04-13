package main

import "os"
import "fmt"
import "github.com/neurlang/classifier/parallel"
import "bufio"
import "strings"
import "math/rand"

func load(filename string, top int) [][2]string {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	var slice [][2]string

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")

		// Check if we have exactly two or three columns
		if len(columns) != 2 {
			fmt.Println("Line does not have exactly two columns:", line)
			continue
		}

		// Process each column
		column1 := columns[0]
		column2 := columns[1]

		slice = append(slice, [2]string{column1, column2})
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

func loop(slice [][2]string, group int, do func(string, string)) {
	parallel.ForEach(len(slice), group, func(n int) {
		// Process each column
		column1 := slice[n][0]
		column2 := slice[n][1]
		// Example: Print the columns
		do(column1, column2)
	})
}
