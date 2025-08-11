package main

import (
	"fmt"
)

func main() {
	input := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	result := FindAnagramGroups(input)

	for key, group := range result {
		fmt.Printf("%q: %q\n", key, group)
	}
}
