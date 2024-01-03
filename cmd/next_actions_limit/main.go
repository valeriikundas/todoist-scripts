package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/valeriikundas/todoist-scripts/telegram"
	"github.com/valeriikundas/todoist-scripts/todoist"
	"log"
	"os"
	"strconv"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

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

	todoistClient := todoist.NewClient(todoistApiToken)
	nextActionsTasksLimitPerProject := 3

	file, err := os.Open("../config.json")
	must(err)

	decoder := json.NewDecoder(file)
	var config struct {
		ExcludeFromZeroProjectsList []string
	}
	err = decoder.Decode(&config)
	must(err)

	projectsWithTooManyTasks, projectsWithZeroTasks := todoistClient.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject, config.ExcludeFromZeroProjectsList)
	log.Printf("projectsWithTooManyTasks=%+v projectsWithZeroTasks=%+v", projectsWithTooManyTasks, projectsWithZeroTasks)

	message := todoistClient.PrettyOutput(projectsWithTooManyTasks, projectsWithZeroTasks)

	tg := telegram.NewTelegram(telegramApiToken)
	err = tg.Send(chatID, message, telegram.ParseModeMarkdownV2)
	if err != nil {
		log.Fatalf("error sending message, %v", err)
	}
}
