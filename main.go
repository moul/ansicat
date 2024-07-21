package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	rateLimit = 7000 // bytes per second
	chunkSize = 1024 // bytes
	sauceSize = 128  // SAUCE metadata size
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		return
	}

	filename := os.Args[1]
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
	data := make([]byte, fileSize)
	_, err = file.Read(data)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Check for SAUCE metadata
	sauceBytes := 0
	if fileSize > sauceSize {
		sauceHeader := data[len(data)-sauceSize : len(data)-sauceSize+5]
		if string(sauceHeader) == "SAUCE" {
			comments := int(data[len(data)-sauceSize+5+94])
			sauceBytes = 1 + 5 + comments*64 + 128
		}
	}

	// Adjust data to exclude SAUCE metadata if present
	if sauceBytes > 0 {
		data = data[:fileSize-int64(sauceBytes)]
	}

	// Convert CP437 to UTF-8
	decoder := charmap.CodePage437.NewDecoder()
	utf8Data, err := decoder.Bytes(data)
	if err != nil {
		fmt.Printf("Error converting data: %v\n", err)
		return
	}

	// Output the data with rate limiting
	reader := bufio.NewReader(bytes.NewReader(utf8Data))
	delay := time.Second / (rateLimit / chunkSize)

	for {
		buf := make([]byte, chunkSize)
		n, err := reader.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
			time.Sleep(delay)
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
