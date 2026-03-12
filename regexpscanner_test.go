package regexpscanner_test

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

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

func BenchmarkMakeSplitter(b *testing.B) {
	// The testing framework will automatically adjust b.N until the benchmark
	// runs for a sufficient amount of time to get stable results.
	// All setup code should go *before* the loop.
	splitter := rs.MakeSplitter(regexp.MustCompile(`</?[a-z]+>`))
	for i := 0; i < b.N; i++ {
		scanner := bufio.NewScanner(strings.NewReader("<html><body><p>Welcome to My Website</p></body></html>"))
		// be sure to call Split()
		scanner.Split(splitter)
		for scanner.Scan() {
			_ = scanner.Text()
		}
	}
}

func TestSplitterBoundary(t *testing.T) {
	re := regexp.MustCompile(`a+`)

	// "aaaaaa" with a small buffer will test the boundary logic.
	input := "aaaaaa"
	reader := strings.NewReader(input)
	scanner := bufio.NewScanner(reader)

	// Start with a buffer of 3 (smaller than the final match)
	buf := make([]byte, 3)
	scanner.Buffer(buf, 1024)
	scanner.Split(rs.MakeSplitter(re))

	var matches []string
	for scanner.Scan() {
		matches = append(matches, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	if len(matches) != 1 || matches[0] != "aaaaaa" {
		t.Errorf("Expected 1 match 'aaaaaa', got %v", matches)
	}
}

func TestSplitterMixed(t *testing.T) {
	// Tests when a match is preceded by non-matching junk
	re := regexp.MustCompile(`a+`)
	input := "---aaaaaa"
	reader := strings.NewReader(input)
	scanner := bufio.NewScanner(reader)

	// Buffer of 6 (will contain "---aaa" and then advance past "---")
	buf := make([]byte, 6)
	scanner.Buffer(buf, 1024)
	scanner.Split(rs.MakeSplitter(re))

	var matches []string
	for scanner.Scan() {
		matches = append(matches, scanner.Text())
	}

	if len(matches) != 1 || matches[0] != "aaaaaa" {
		t.Errorf("Expected 1 match 'aaaaaa', got %v", matches)
	}
}

func BenchmarkScannerEfficiency(b *testing.B) {
	// 1. Prepare a "Simple Document"
	content := strings.Repeat("log: info message here\nlog: error something broke\n", 1000)
	re := regexp.MustCompile(`log: [[:alpha:]]+`)

	b.ResetTimer()
	b.ReportAllocs()                // Tracks memory overhead
	b.SetBytes(int64(len(content))) // Reports MB/s

	for i := 0; i < b.N; i++ {
		var callCount int
		// Wrap the splitter to count invocations
		rawSplitter := rs.MakeSplitter(re)
		countedSplitter := func(data []byte, atEOF bool) (int, []byte, error) {
			callCount++
			return rawSplitter(data, atEOF)
		}

		scanner := bufio.NewScanner(strings.NewReader(content))
		scanner.Split(countedSplitter)

		var matchCount int
		for scanner.Scan() {
			matchCount++
			_ = scanner.Bytes()
		}

		if i == 0 {
			b.Logf("Matches: %d, Splitter Invocations: %d (%.2f calls/match)",
				matchCount, callCount, float64(callCount)/float64(matchCount))
		}
	}
}

func BenchmarkScannerThroughput(b *testing.B) {
	content := strings.Repeat("a", 1024*1024) // 1MB of 'a's
	re := regexp.MustCompile(`a+`)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(content)))

	for i := 0; i < b.N; i++ {
		scanner := rs.MakeScanner(strings.NewReader(content), re)
		for scanner.Scan() {
			_ = scanner.Bytes()
		}
	}
}

func BenchmarkPrimitiveThroughput(b *testing.B) {
	content := []byte(strings.Repeat("a", 1024*1024)) // 1MB of 'a's
	pattern := []byte("a")

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(content)))

	for i := 0; i < b.N; i++ {
		scanner := bufio.NewScanner(bytes.NewReader(content))
		// Primitive split function using bytes.Index and greedy matching
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			start := bytes.Index(data, pattern)
			if start >= 0 {
				// Greedy match: find where the 'a's end
				end := start + 1
				for end < len(data) && data[end] == 'a' {
					end++
				}
				// If we hit the end of the buffer and more is coming, request more
				if !atEOF && end == len(data) {
					return start, nil, nil
				}
				return end, data[start:end], nil
			}
			if !atEOF {
				return 0, nil, nil
			}
			return len(data), nil, nil
		})
		for scanner.Scan() {
			_ = scanner.Bytes()
		}
	}
}
