package reader

import (
	"fmt"
	"io/ioutil"
	"log"
	"unicode/utf8"
)

// a ReadFile func that reads a file and decodes it into a slice of lines, where
// each line is a slice of runes. This is clearly less efficient than doing a
// smart tokenizer, but this is way simpler to grok. We're not going to be
// reading a zillion source files.

// Line is a slice of runes
type Line []rune

// InputFile is a slice of Lines
type InputFile []Line

const bom = 0xFEFF // byte order mark, only permitted as very first character

// ReadFile reads a file and returns its contents a
func ReadFile(path string) (InputFile, error) {

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot ReadFile: %s", err)
	}

	result := make([]Line, 0, 0)

	if len(content) > 0 {
		r, l := utf8.DecodeRune(content)
		if r == bom {
			content = content[l:]
		}
	}

	for len(content) > 0 {

		nextLine, consumed := scanLine(content)
		content = content[consumed:]

		result = append(result, nextLine)
	}

	return result, nil
}

// hacked from pkg/bytes.Runes()
func scanLine(content []byte) (Line, int) {

	var t Line

	consumed := 0
	for len(content) > 0 {

		r, c := utf8.DecodeRune(content)
		if r == utf8.RuneError || c == 0 {
			log.Fatalf("error decoding UTF8")
		}
		content = content[c:]
		consumed += c

		if r == '\r' {
			// check if "\r\n"
			if len(content) > 0 {
				nl, _ := utf8.DecodeRune(content)
				if nl == '\n' {
					return t, consumed + 1
				}
			}
			return t, consumed
		}

		if r == '\n' {
			return t, consumed
		}

		t = append(t, r)
	}

	return t, consumed
}
