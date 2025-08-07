package main

import (
	"fmt"
	"log"
)

func main() {
	s := "qwe\\"
	result, err := Unpack(s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
