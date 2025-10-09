package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	chatId := os.Getenv("TELEGRAM_CHATID")

	if token == "" || chatId == "" {
		log.Fatal("TELEGRAM_TOKEN or TELEGRAM_CHATID is not set")
	}

	fmt.Println("Token:", token)
	fmt.Println("ChatId:", chatId)
}
