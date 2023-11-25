package main

import (
	"fmt"
	"log"

	"github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	nextActionsTasksLimitPerProject := 3
	apiToken := todoist_utils.ReadApiTokenFromDotenv()
	todoist := todoist_utils.NewTodoist(apiToken)
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
