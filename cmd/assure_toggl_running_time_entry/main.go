package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/valeriikundas/todoist-scripts/toggl"
)

func main() {
	log.SetFlags(log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file, %v", err)
	}

	togglApiToken := os.Getenv("TOGGL_API_TOKEN")
	telegramApiToken := os.Getenv("TELEGRAM_API_TOKEN")

	telegramUserIDStr := os.Getenv("TELEGRAM_USER_ID")
	telegramUserID, err := strconv.Atoi(telegramUserIDStr)
	if err != nil {
		log.Fatalf("error converting telegram user id to int, %v", err)
	}

	err = toggl.NotifyIfNoRunningTogglEntry(togglApiToken, telegramApiToken, telegramUserID)
	if err != nil {
		log.Fatalf("error notifying if no running toggl entry, %v", err)
	}
}
