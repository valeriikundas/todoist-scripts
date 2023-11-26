package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/valeriikundas/todoist-scripts/telegram"
	"github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file, %v", err)
	}

	todoistApiToken := os.Getenv("TODOIST_API_TOKEN")
	telegramApiToken := os.Getenv("TELEGRAM_API_TOKEN")
	chatID := os.Getenv("TELEGRAM_USER_ID")

	nextActionsTasksLimitPerProject := 3

	todoist := todoist_utils.NewTodoist(todoistApiToken)
	projectsWithTooManyTasks, projectsWithZeroTasks := todoist.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject)
	log.Printf("projectsWithTooManyTasks=%+v projectsWithZeroTasks=%+v", projectsWithTooManyTasks, projectsWithZeroTasks)

	message := todoist.PrettyOutput(projectsWithTooManyTasks, projectsWithZeroTasks)

	tg := telegram.NewTelegram(telegramApiToken)
	err = tg.Send(chatID, message)
	if err != nil {
		log.Fatalf("error sending message, %v", err)
	}
}
