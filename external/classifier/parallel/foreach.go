package parallel

import "sync"

// ForEach executes a for loop with a limited number of concurrent goroutines.
// Each goroutine processes one integer, from 0 to length.
func ForEach(length, limit int, body func(i int)) {
	if limit <= 0 {
		limit = 1 // Default to 1 if limit is zero or negative
	}
	if length <= 0 {
		return // No iterations to perform
	}

	sem := make(chan struct{}, limit) // Semaphore with buffer size 'limit'
	var wg sync.WaitGroup
	wg.Add(length)

	for i := 0; i < length; i++ {
		i := i            // Capture loop variable
		sem <- struct{}{} // Acquire semaphore
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore after function exits

			body(i)
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish
}
