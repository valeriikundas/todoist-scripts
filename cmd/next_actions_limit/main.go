package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/valeriikundas/todoist-scripts/telegram"
	"github.com/valeriikundas/todoist-scripts/todoist"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file, %v", err)
	}

	todoistApiToken := os.Getenv("TODOIST_API_TOKEN")
	telegramApiToken := os.Getenv("TELEGRAM_API_TOKEN")
	chatID, err := strconv.Atoi(os.Getenv("TELEGRAM_USER_ID"))
	if err != nil {
		log.Fatalf("error converting chatID to int, %v", err)
	}

	nextActionsTasksLimitPerProject := 3

	todoistClient := todoist.NewClient(todoistApiToken)
	projectsWithTooManyTasks, projectsWithZeroTasks := todoistClient.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject)
	log.Printf("projectsWithTooManyTasks=%+v projectsWithZeroTasks=%+v", projectsWithTooManyTasks, projectsWithZeroTasks)

	message := todoistClient.PrettyOutput(projectsWithTooManyTasks, projectsWithZeroTasks)

	tg := telegram.NewTelegram(telegramApiToken)
	err = tg.Send(chatID, message)
	if err != nil {
		log.Fatalf("error sending message, %v", err)
	}
}
