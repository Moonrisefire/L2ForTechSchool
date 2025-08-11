package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

var monthMap = map[string]int{
	"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
	"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
	"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
}

func main() {
	column := flag.Int("k", 1, "Сортировка по номеру столбца (разделитель — табуляция)")
	numeric := flag.Bool("n", false, "Числовая сортировка")
	reverse := flag.Bool("r", false, "Обратный порядок сортировки")
	unique := flag.Bool("u", false, "Выводить только уникальные строки")
	monthSort := flag.Bool("M", false, "Сортировка по названию месяца (Jan..Dec)")
	ignoreBlanks := flag.Bool("b", false, "Игнорировать хвостовые пробелы")
	check := flag.Bool("c", false, "Проверить, отсортированы ли данные")
	human := flag.Bool("h", false, "Сортировка по человекочитаемым числам (K, M, G)")
	flag.Parse()

	var reader io.Reader
	if flag.NArg() > 0 {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка открытия файла: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else {
		reader = os.Stdin
	}

	lines := readLines(reader, *ignoreBlanks)

	if *check {
		if isSorted(lines, *column, *numeric, *monthSort, *human, *reverse) {
			fmt.Println("Входные данные отсортированы")
		} else {
			fmt.Println("Входные данные не отсортированы")
			os.Exit(1)
		}
		return
	}

	sortLines(lines, *column, *numeric, *monthSort, *human, *reverse)

	if *unique {
		lines = uniqueLines(lines)
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

func readLines(r io.Reader, ignoreBlanks bool) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if ignoreBlanks {
			line = strings.TrimRight(line, " ")
		}
		lines = append(lines, line)
	}
	return lines
}

func getField(line string, col int) string {
	fields := strings.Split(line, "\t")
	if col <= 0 || col > len(fields) {
		return ""
	}
	return fields[col-1]
}

func parseHuman(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	last := s[len(s)-1]
	multiplier := 1.0
	switch last {
	case 'K', 'k':
		multiplier = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val * multiplier
}

func sortLines(lines []string, col int, numeric, monthSort, human, reverse bool) {
	sort.Slice(lines, func(i, j int) bool {
		a := getField(lines[i], col)
		b := getField(lines[j], col)

		var less bool
		switch {
		case human:
			less = parseHuman(a) < parseHuman(b)
		case numeric:
			af, _ := strconv.ParseFloat(a, 64)
			bf, _ := strconv.ParseFloat(b, 64)
			less = af < bf
		case monthSort:
			less = monthMap[a] < monthMap[b]
		default:
			less = a < b
		}

		if reverse {
			return !less
		}
		return less
	})
}

func uniqueLines(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	uniq := []string{lines[0]}
	for i := 1; i < len(lines); i++ {
		if lines[i] != lines[i-1] {
			uniq = append(uniq, lines[i])
		}
	}
	return uniq
}

func isSorted(lines []string, col int, numeric, monthSort, human, reverse bool) bool {
	for i := 1; i < len(lines); i++ {
		a := getField(lines[i-1], col)
		b := getField(lines[i], col)
		var less bool
		switch {
		case human:
			less = parseHuman(a) <= parseHuman(b)
		case numeric:
			af, _ := strconv.ParseFloat(a, 64)
			bf, _ := strconv.ParseFloat(b, 64)
			less = af <= bf
		case monthSort:
			less = monthMap[a] <= monthMap[b]
		default:
			less = a <= b
		}
		if reverse {
			if !less {
				return false
			}
		} else {
			if !less {
				return false
			}
		}
	}
	return true
}
