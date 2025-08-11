package main

import (
	"sort"
	"strings"
	"unicode/utf8"
)

func FindAnagramGroups(words []string) map[string][]string {

	anagramMap := make(map[string][]string)
	wordMap := make(map[string]string)

	for _, word := range words {
		wordLower := strings.ToLower(strings.TrimSpace(word))
		if utf8.RuneCountInString(wordLower) == 0 {
			continue
		}

		runes := []rune(wordLower)
		sort.Slice(runes, func(i, j int) bool {
			return runes[i] < runes[j]
		})
		sortedKey := string(runes)

		if _, exists := wordMap[sortedKey]; !exists {
			wordMap[sortedKey] = wordLower
		}
		anagramMap[sortedKey] = append(anagramMap[sortedKey], wordLower)
	}

	result := make(map[string][]string)

	for sortedKey, group := range anagramMap {
		if len(group) < 2 {
			continue
		}
		sort.Strings(group)
		firstWord := wordMap[sortedKey]
		result[firstWord] = group
	}

	return result
}
