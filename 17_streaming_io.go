package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func readGzip(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer gz.Close()

	scanner := bufio.NewScanner(gz)

	// [default is 64kb, set if the line is longer than that]
	// Set initial buffer + max allowed token size (e.g., 1MB).
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		_ = scanner.Text()
	}

	// scanner.Err() will be nil if it just hit EOF.
	// It only returns errors like "bufio.ErrTooLong".
	if err := scanner.Err(); err != nil {
		log.Printf("Scan error: %v", err)
	}
}

func readHugeWithReader(path string) {
	f, _ := os.Open(path)
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Process the very last bit of data if the file didn't end with \n
				if len(line) > 0 {
					fmt.Print(line)
				}
				break // Exit loop gracefully
			}
			log.Fatal(err) // Actual error (disk failure, etc.)
		}
		fmt.Print(line)
	}
}

func readInChunks(path string) {
	f, _ := os.Open(path)
	defer f.Close()

	buf := make([]byte, 4096) // 4KB buffer
	for {
		n, err := f.Read(buf)
		if n > 0 {
			// ALWAYS process the 'n' bytes read before checking the error
			//process(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break // Successful end
			}
			log.Fatal(err) // Real error
		}
	}
}

func pipeData(srcPath, dstPath string) {
	src, _ := os.Open(srcPath)
	defer src.Close()

	dst, _ := os.Create(dstPath)
	defer dst.Close()

	// io.Copy returns (bytesWritten, error).
	// It returns nil (not EOF) on success because EOF is the expected exit.
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Fatalf("Copy failed: %v", err)
	}
}

// FindLogErrors doesn't care if the data comes from a file or a string.
// It just needs an io.Reader.
func findLogErrors(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ERROR") {
			fmt.Println("Found:", line)
		}
	}
}

func checkLogErrors() {
	// Instead of opening a real file, we create a "virtual" file from a string
	mockFileData := `INFO: starting up
ERROR: connection failed
INFO: retrying...
ERROR: auth timeout`

	// This is the "Adapter"
	reader := strings.NewReader(mockFileData)

	// Now we pass our string-reader into the function
	findLogErrors(reader)
}

// TODO: add more tests
