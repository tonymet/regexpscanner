package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
)

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

func MakeScanner(in io.Reader, re *regexp.Regexp) *bufio.Scanner {
	scanner := bufio.NewScanner(in)
	scanner.Split(MakeSplitter(patternRE))
	return scanner
}

func Read(in io.Reader) {
	ProcessTokens(in, patternRE, func(text string) {
		fmt.Println(text)
	})
}

func ProcessTokens(in io.Reader, re *regexp.Regexp, handler func(string)) {
	scanner := MakeScanner(in, patternRE)
	for scanner.Scan() {
		handler(scanner.Text())
	}
}

var (
	pattern   string
	patternRE *regexp.Regexp
)

func init() {
	flag.StringVar(&pattern, "pattern", "\\w+", "regex pattern for testing")
}

func main() {
	flag.Parse()
	patternRE = regexp.MustCompile(pattern)
	Read(os.Stdin)
}
