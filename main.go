package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/text/encoding/charmap"
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

	decoder := charmap.CodePage437.NewDecoder()
	reader := decoder.Reader(file)
	bufferedReader := bufio.NewReader(reader)

	const rateLimit = 7000 // bytes per second
	const chunkSize = 1024 // bytes
	delay := time.Second / (rateLimit / chunkSize)

	for {
		buf := make([]byte, chunkSize)
		n, err := bufferedReader.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
			time.Sleep(delay)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			break
		}
	}
}
