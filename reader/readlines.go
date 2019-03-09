package reader

import (
	"bufio"
	"io"
	"os"
)

// ReadLines reads a file into a slice of strings.
func ReadLines(path string) ([]string, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadLinesToStrings(f), nil
}

// ReadLinesToStrings converts input into separate lines.
func ReadLinesToStrings(r io.Reader) []string {
	var lines []string

	s := bufio.NewScanner(r)

	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines
}
