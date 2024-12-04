// ©️ 2024 Anthony Metzids
// regexpscanner command.  Extract text from files
//
// install : go install github.com/tonymet/regexpscanner@latest
//
// usage: regexpscanner -pattern PATTERN < input.txt
package main

import (
	"flag"
	"fmt"
	rs "github.com/tonymet/regexpscanner"
	"io"
	"os"
	"regexp"
)

var (
	pattern   string
	patternRE *regexp.Regexp
)

func init() {
	flag.StringVar(&pattern, "pattern", "\\w+", "regex pattern for testing")
}

func Read(in io.Reader) {
	rs.ProcessTokens(in, patternRE, func(text string) {
		fmt.Println(text)
	})
}

func main() {
	flag.Parse()
	patternRE = regexp.MustCompile(pattern)
	Read(os.Stdin)
}
