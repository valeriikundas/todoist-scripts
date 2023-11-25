package main

import (
	"log"

	"github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	nextActionsTasksLimitPerProject := 3
	apiToken := todoist_utils.ReadApiTokenFromDotenv()
	todoist := todoist_utils.NewTodoist(apiToken)
	projectsWithTooManyTasks, projectsWithZeroTasks := todoist.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject)
	todoist_utils.PrintOutput(projectsWithTooManyTasks, projectsWithZeroTasks)

	// TODO: send notification in telegram
}
