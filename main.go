package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type Options struct {
	After      int
	Before     int
	Context    int
	CountOnly  bool
	IgnoreCase bool
	Invert     bool
	Fixed      bool
	LineNum    bool
	Pattern    string
	Filename   string
}

func main() {
	after := flag.Int("A", 0, "Показать N строк после совпадения")
	before := flag.Int("B", 0, "Показать N строк до совпадения")
	context := flag.Int("C", 0, "Показать N строк вокруг совпадения (эквивалентно -A N -B N)")
	countOnly := flag.Bool("c", false, "Показать только количество совпадений")
	ignoreCase := flag.Bool("i", false, "Игнорировать регистр")
	invert := flag.Bool("v", false, "Инвертировать фильтр")
	fixed := flag.Bool("F", false, "Шаблон — фиксированная строка (не регулярное выражение)")
	lineNum := flag.Bool("n", false, "Выводить номера строк")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalln("Укажите шаблон поиска.")
	}

	pattern := flag.Arg(0)
	filename := ""
	if flag.NArg() > 1 {
		filename = flag.Arg(1)
	}

	opts := Options{
		After:      *after,
		Before:     *before,
		Context:    *context,
		CountOnly:  *countOnly,
		IgnoreCase: *ignoreCase,
		Invert:     *invert,
		Fixed:      *fixed,
		LineNum:    *lineNum,
		Pattern:    pattern,
		Filename:   filename,
	}

	if err := Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Run(opts Options) error {
	if opts.Context > 0 {
		opts.After = opts.Context
		opts.Before = opts.Context
	}

	lines, err := readLines(opts.Filename)
	if err != nil {
		return err
	}

	matcher, err := compileMatcher(opts)
	if err != nil {
		return err
	}

	type matchInfo struct {
		index int
	}

	var matches []matchInfo
	for i, line := range lines {
		match := matcher(line)
		if opts.Invert {
			match = !match
		}
		if match {
			matches = append(matches, matchInfo{index: i})
		}
	}

	if opts.CountOnly {
		fmt.Println(len(matches))
		return nil
	}

	printed := make(map[int]bool)
	for _, match := range matches {
		start := max(0, match.index-opts.Before)
		end := min(len(lines)-1, match.index+opts.After)
		for i := start; i <= end; i++ {
			if printed[i] {
				continue
			}
			printed[i] = true
			if opts.LineNum {
				fmt.Printf("%d:%s\n", i+1, lines[i])
			} else {
				fmt.Println(lines[i])
			}
		}
	}
	return nil
}

func readLines(filename string) ([]string, error) {
	var scanner *bufio.Scanner
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func compileMatcher(opts Options) (func(string) bool, error) {
	pattern := opts.Pattern
	if opts.IgnoreCase {
		pattern = strings.ToLower(pattern)
	}

	if opts.Fixed {
		return func(line string) bool {
			if opts.IgnoreCase {
				line = strings.ToLower(line)
			}
			return strings.Contains(line, pattern)
		}, nil
	}

	re, err := regexp.Compile(getRegexpPattern(pattern, opts.IgnoreCase))
	if err != nil {
		return nil, err
	}
	return func(line string) bool {
		return re.MatchString(line)
	}, nil
}

func getRegexpPattern(p string, ignoreCase bool) string {
	if ignoreCase {
		return "(?i)" + p
	}
	return p
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
