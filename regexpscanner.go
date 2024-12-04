// ©️ 2024 Anthony Metzidis
/*
regexpscanner -- stream-based scanner that scans io.Reader and extracts tokens
matching a regular expression.
*/
package regexpscanner

import (
	"bufio"
	"io"
	"regexp"
)

// MakeSplitter(re) creates a splitter to be passed to scanners.Split()
// the re will be used to tokenize input passed to the scanner
// splitters can be wrapped with more complicated splitters for further processing
// see bufio.Scanner for example splitter-wrappers
func MakeSplitter(re *regexp.Regexp) func([]byte, bool) (int, []byte, error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		loc := re.FindIndex(data)
		if loc == nil {
			// try again
			if !atEOF {
				return 0, nil, nil
			}
			return 0, nil, bufio.ErrFinalToken
		}
		return loc[1], data[loc[0]:loc[1]], nil
	}
}

// MakeScanner creates a scanner you can call scanner.Scan() and scanner.Text() with.
// Calling scanner.Scan() && scanner.Text() will return the latest token matching the regex in the stream.
func MakeScanner(in io.Reader, re *regexp.Regexp) *bufio.Scanner {
	scanner := bufio.NewScanner(in)
	scanner.Split(MakeSplitter(re))
	return scanner
}

// ProcessTokens calls handler(string) for each matching token from the Scanner.
func ProcessTokens(in io.Reader, re *regexp.Regexp, handler func(string)) {
	scanner := MakeScanner(in, re)
	for scanner.Scan() {
		handler(scanner.Text())
	}
}
