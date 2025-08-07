package main

import (
	"errors"
	"strings"
	"unicode"
)

func Unpack(input string) (string, error) {
	var result strings.Builder
	runes := []rune(input)
	length := len(runes)

	if length == 0 {
		return "", nil
	}

	escaped := false
	for i := 0; i < length; i++ {
		current := runes[i]

		if !escaped && current == '\\' {
			escaped = true
			continue
		}

		if unicode.IsDigit(current) && !escaped {
			if i == 0 {
				return "", errors.New("Некорректная строка: начинается с цифры")
			}

			prev := runes[i-1]
			if unicode.IsDigit(prev) && !isEscaped(runes, i-1) {
				return "", errors.New("Некорректная строка: две подряд неэкранированные цифры")
			}

			count := int(current - '0')
			if count == 0 {
				s := []rune(result.String())
				if len(s) > 0 {
					result.Reset()
					result.WriteString(string(s[:len(s)-1]))
				}
			} else {
				s := []rune(result.String())
				if len(s) == 0 {
					return "", errors.New("Некорректная строка: цифра без предыдущего символа")
				}
				lastRune := s[len(s)-1]
				for j := 1; j < count; j++ {
					result.WriteRune(lastRune)
				}
			}
		} else {
			result.WriteRune(current)
			escaped = false
		}
	}

	if escaped {
		return "", errors.New("Некорректная строка: строка заканчивается escape-символом")
	}

	return result.String(), nil
}

func isEscaped(runes []rune, i int) bool {
	count := 0
	for j := i - 1; j >= 0 && runes[j] == '\\'; j-- {
		count++
	}
	return count%2 == 1
}
