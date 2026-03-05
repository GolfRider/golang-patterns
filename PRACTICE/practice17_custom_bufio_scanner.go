package main

import (
	"bufio"
	"fmt"
	"strings"
)

func problem17() {
	// 1. Mock input: Data delimited by '#' instead of newlines
	input := "RECORD1#RECORD2_WITH_LONG_DATA#RECORD3"
	reader := strings.NewReader(input)

	scanner := bufio.NewScanner(reader)

	// 2. Custom Buffer Size
	// Default is 64KB. For Data Platforms, we often bump this to handle large individual records.
	const maxTokenSize = 1024 * 1024 // 1MB buffer limit
	buf := make([]byte, 64*1024)     // Initial 64KB allocation
	scanner.Buffer(buf, maxTokenSize)

	// 3. Custom Delimiter Logic (SplitFunc)
	// We define how the scanner finds the "end" of a token.
	/*split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// Search for our custom delimiter '#'
		if i := bytes.IndexByte(data, '#'); i >= 0 {
			// We found a delimiter! Return the data before it.
			return i + 1, data[0:i], nil
		}
		// If we're at EOF, return the remaining data
		if atEOF {
			return len(data), data, nil
		}
		// Request more data from the reader
		return 0, nil, nil
	}
	scanner.Split(split)*/

	// 4. Execution Loop
	for scanner.Scan() {
		fmt.Printf("Found Record: %s\n", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning: %v\n", err)
	}
}
