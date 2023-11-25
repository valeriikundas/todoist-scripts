package main

import (
	"log"

	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	nextActionsTasksLimitPerProject := 3
	projectsWithTooManyTasks, projectsWithZeroTasks := todoist.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject)
	todoist.PrintOutput(projectsWithTooManyTasks, projectsWithZeroTasks)

	// TODO: send notification in telegram
}
