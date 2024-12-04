package regexpscanner_test

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	rs "github.com/tonymet/regexpscanner"
)

// use ProcessTokens when a simple callback-based stream tokenizer is needed
func ExampleProcessTokens() {
	rs.ProcessTokens(
		strings.NewReader("<html><body><p>Welcome to My Website</p></body></html>"),
		regexp.MustCompile(`</?[a-z]+>`),
		func(text string) {
			fmt.Println(text)
		})
	// Output:
	// <html>
	// <body>
	// <p>
	// </p>
	// </body>
	// </html>
}

// use MakeSplitter to create a "splitter" for scanner.Split()
func ExampleMakeSplitter() {
	splitter := rs.MakeSplitter(regexp.MustCompile(`</?[a-z]+>`))
	scanner := bufio.NewScanner(strings.NewReader("<html><body><p>Welcome to My Website</p></body></html>"))
	// be sure to call Split()
	scanner.Split(splitter)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	// Output:
	// <html>
	// <body>
	// <p>
	// </p>
	// </body>
	// </html>
}

// use MakeScanner to create a scanner that will tokenize using the regex
func ExampleMakeScanner() {
	scanner := rs.MakeScanner(strings.NewReader("<html><body><p>Welcome to My Website</p></body></html>"),
		regexp.MustCompile(`</?[a-z]+>`),
	)
	// scanner has Split function defined using the regexp passed to MakeScanner
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	// Output:
	// <html>
	// <body>
	// <p>
	// </p>
	// </body>
	// </html>
}
