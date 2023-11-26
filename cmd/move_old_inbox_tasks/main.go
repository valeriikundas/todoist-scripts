package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file, %v", err)
	}

	todoistApiToken := os.Getenv("TODOIST_API_TOKEN")

	srcProjectName := "Inbox"
	dstProjectName := "inbox_archive"
	oldThreshold := time.Hour * 24 * 3
	dryRun := false

	todoist := todoist.NewTodoist(todoistApiToken)
	todoist.MoveOlderTasks(srcProjectName, dstProjectName, oldThreshold, dryRun)
}
