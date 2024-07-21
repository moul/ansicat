package ansicat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const sauceSize = 128 // SAUCE metadata size

// ProcessFile processes the ANSI file, converts it from CP437 to UTF-8, and writes to the output stream.
func ProcessFile(input io.Reader, output io.Writer, rateLimit int, chunkSize int) error {
	// Read the entire file into memory
	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
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
		return fmt.Errorf("error converting data: %w", err)
	}

	// Output the data with optional rate limiting
	reader := bufio.NewReader(bytes.NewReader(utf8Data))
	var delay time.Duration
	if rateLimit > 0 && rateLimit >= chunkSize {
		delay = time.Second / (time.Duration(rateLimit) / time.Duration(chunkSize))
	} else {
		rateLimit = 0 // Disable rate limiting if rate limit is smaller than chunk size
	}

	for {
		buf := make([]byte, chunkSize)
		n, err := reader.Read(buf)
		if n > 0 {
			_, writeErr := output.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("error writing to output: %w", writeErr)
			}
			if rateLimit > 0 {
				time.Sleep(delay)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading data: %w", err)
		}
	}

	return nil
}
