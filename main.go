package main

import (
	"flag"
	"fmt"
	"os"

	"moul.io/ansicat/ansicat"
)

const (
	defaultRateLimit = 7000 // bytes per second
	defaultChunkSize = 1024 // bytes
)

func main() {
	// Define command-line flags
	rateLimit := flag.Int("rate-limit", defaultRateLimit, "Rate limit in bytes per second (0 for no limit)")
	chunkSize := flag.Int("chunk-size", defaultChunkSize, "Chunk size in bytes")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: go run main.go [--rate-limit=<bytes per second>] [--chunk-size=<bytes>] <filename>")
		return
	}

	filename := flag.Arg(0)
	var file *os.File
	var err error

	if filename == "-" {
		// Read from stdin
		file = os.Stdin
	} else {
		// Open file
		file, err = os.Open(filename)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()
	}

	// Process the ANSI file
	err = ansicat.ProcessFile(file, os.Stdout, *rateLimit, *chunkSize)
	if err != nil {
		fmt.Printf("Error processing file: %v\n", err)
	}
}
