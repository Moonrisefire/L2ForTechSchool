package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/beevik/ntp"
)

func main() {
	currentTime, err := ntp.Time("pool.ntp.org")
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("Ошибка получения времени с NTP-сервера: %v", err)
		os.Exit(1)
	}

	fmt.Println("Точное время (через NTP):", currentTime.Format(time.RFC1123))
}
