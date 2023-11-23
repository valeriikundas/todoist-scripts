package main

import (
	"log"

	"github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	nextActionsTasksLimitPerProject := 3
	projectsWithTooManyTasks, projectsWithZeroTasks := todoist_utils.GetProjectsWithTooManyAndZeroTasks(nextActionsTasksLimitPerProject)
	todoist_utils.PrintOutput(projectsWithTooManyTasks, projectsWithZeroTasks)
}
