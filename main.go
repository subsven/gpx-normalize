package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath" // Added for output file path construction
	"sync"
)

func main() {
	flag.Parse() // Parse command-line flags

	files := flag.Args() // Get non-flag arguments (file paths)

	if len(files) == 0 {
		fmt.Println("Usage: gpx-normalizer <file1.gpx> [file2.gpx] ...")
		os.Exit(1)
	}

	var wg sync.WaitGroup

	log.Printf("Starting normalization for %d GPX file(s)...", len(files))

	for _, filePath := range files {
		wg.Add(1) // Increment the WaitGroup counter

		go func(file string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			log.Printf("Processing %s...", file)

			// Generate output filename
			dir := filepath.Dir(file)
			base := filepath.Base(file)
			outputFile := filepath.Join(dir, "normalized-"+base)

			err := normalizeGPX(file, outputFile) // Call refactored normalizeGPX
			if err != nil {
				log.Printf("Error normalizing %s to %s: %v", file, outputFile, err)
			} else {
				log.Printf("Successfully normalized %s to %s", file, outputFile)
			}
		}(filePath)
	}

	wg.Wait() // Wait for all goroutines to complete

	log.Println("All GPX files processed.")
}
