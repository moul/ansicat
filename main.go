package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	defaultRateLimit = 7000 // bytes per second
	defaultChunkSize = 1024 // bytes
	sauceSize        = 128  // SAUCE metadata size
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
	var data []byte
	var err error

	if filename == "-" {
		// Read from stdin
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("Error reading from stdin: %v\n", err)
			return
		}
	} else {
		// Read from file
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		// Get file size
		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Printf("Error getting file info: %v\n", err)
			return
		}
		fileSize := fileInfo.Size()

		// Read the entire file into memory
		data = make([]byte, fileSize)
		_, err = file.Read(data)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}
	}

	// Check for SAUCE metadata
	sauceBytes := 0
	if len(data) > sauceSize {
		sauceHeader := data[len(data)-sauceSize : len(data)-sauceSize+5]
		if string(sauceHeader) == "SAUCE" {
			comments := int(data[len(data)-sauceSize+5+94])
			sauceBytes = 1 + 5 + comments*64 + 128
		}
	}

	// Adjust data to exclude SAUCE metadata if present
	if sauceBytes > 0 {
		data = data[:len(data)-sauceBytes]
	}

	// Convert CP437 to UTF-8
	decoder := charmap.CodePage437.NewDecoder()
	utf8Data, err := decoder.Bytes(data)
	if err != nil {
		fmt.Printf("Error converting data: %v\n", err)
		return
	}

	// Output the data with optional rate limiting
	reader := bufio.NewReader(bytes.NewReader(utf8Data))
	var delay time.Duration
	if *rateLimit > 0 && *rateLimit >= *chunkSize {
		delay = time.Second / (time.Duration(*rateLimit) / time.Duration(*chunkSize))
	} else {
		*rateLimit = 0 // Disable rate limiting if rate limit is smaller than chunk size
	}

	for {
		buf := make([]byte, *chunkSize)
		n, err := reader.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
			if *rateLimit > 0 {
				time.Sleep(delay)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading data: %v\n", err)
			break
		}
	}
}
