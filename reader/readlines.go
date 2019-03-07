package reader

import (
	"bufio"
	"os"
)

// ReadLines reads a file into a slice of strings.
func ReadLines(path string) ([]string, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string

	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines, nil
}
