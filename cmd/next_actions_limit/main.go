package main

import (
	"log"

	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	nextActionsTasksLimitPerProject := 3
	apiToken := todoist.ReadApiTokenFromDotenv()
	projectsWithTooManyTasks, projectsWithZeroTasks := todoist.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject, apiToken)
	todoist.PrintOutput(projectsWithTooManyTasks, projectsWithZeroTasks)

	// TODO: send notification in telegram
}
