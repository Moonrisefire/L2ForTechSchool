package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type Fields struct {
	set   map[int]struct{}
	order []int
}

func ParseFields(spec string) (*Fields, error) {
	result := &Fields{set: make(map[int]struct{})}
	parts := strings.Split(spec, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("некорректный диапазон: %s", part)
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || start <= 0 || end < start {
				return nil, fmt.Errorf("некорректный диапазон: %s", part)
			}
			for i := start; i <= end; i++ {
				if _, exists := result.set[i]; !exists {
					result.set[i] = struct{}{}
					result.order = append(result.order, i)
				}
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil || num <= 0 {
				return nil, fmt.Errorf("некорректное поле: %s", part)
			}
			if _, exists := result.set[num]; !exists {
				result.set[num] = struct{}{}
				result.order = append(result.order, num)
			}
		}
	}
	return result, nil
}

func Run(r io.Reader, delimiter string, separated bool, fields *Fields) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if separated && !strings.Contains(line, delimiter) {
			continue
		}
		cols := strings.Split(line, delimiter)
		var output []string
		for _, idx := range fields.order {
			if idx-1 < len(cols) {
				output = append(output, cols[idx-1])
			}
		}
		if len(output) > 0 {
			fmt.Println(strings.Join(output, delimiter))
		}
	}
	return scanner.Err()
}

func main() {
	fieldsFlag := flag.String("f", "", "Номера полей для вывода, например: 1,2,4-6")
	delimiterFlag := flag.String("d", "\t", "Разделитель (по умолчанию табуляция)")
	separatedFlag := flag.Bool("s", false, "Игнорировать строки без разделителя")

	flag.Parse()

	if *fieldsFlag == "" {
		log.Fatalln("Необходимо указать флаг -f")
	}

	if len([]rune(*delimiterFlag)) != 1 {
		log.Fatalln("Разделитель должен быть одним символом")
	}

	fields, err := ParseFields(*fieldsFlag)
	if err != nil {
		log.Fatalf("Ошибка разбора полей: %v\n", err)
	}

	if err := Run(os.Stdin, *delimiterFlag, *separatedFlag, fields); err != nil {
		log.Fatalf("Ошибка выполнения: %v\n", err)
	}
}
