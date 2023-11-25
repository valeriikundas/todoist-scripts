package main

import (
	"log"
	"time"

	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	srcProjectName := "Inbox"
	dstProjectName := "inbox_archive"
	oldThreshold := time.Hour * 24 * 3
	dryRun := false
	apiToken := todoist.ReadApiTokenFromDotenv()

	todoist.MoveOlderTasks(srcProjectName, dstProjectName, oldThreshold, dryRun, apiToken)
}
