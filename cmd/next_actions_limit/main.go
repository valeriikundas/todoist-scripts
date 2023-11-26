package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file, %v", err)
	}

	todoistApiToken := os.Getenv("TODOIST_API_TOKEN")
	// telegramApiToken := os.Getenv("TELEGRAM_API_TOKEN")

	nextActionsTasksLimitPerProject := 3

	todoist := todoist_utils.NewTodoist(todoistApiToken)
	projectsWithTooManyTasks, projectsWithZeroTasks := todoist.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject)
	printOutput(projectsWithTooManyTasks, projectsWithZeroTasks)

	// TODO: send notification in telegram
}

func printOutput(aboveLimitProjects []todoist_utils.IncorrectProjectSchema, zeroTaskProjects []todoist_utils.IncorrectProjectSchema) {
	for _, p := range aboveLimitProjects {
		fmt.Printf("project \"%s\" has %d @next_action tasks, max allowed is %d. ",
			p.ProjectName, p.TasksCount, p.Limit)
		fmt.Printf("please review and fix at %s\n", p.URL)
	}

	for _, p := range zeroTaskProjects {
		fmt.Printf("does not have @next_action tasks: projectName=%s.", p.ProjectName)
		fmt.Printf("please review and fix at %s\n", p.URL)
	}
}
