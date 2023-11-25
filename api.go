package api

import (
	"context"
	"time"

	"encore.dev/cron"
	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
)

// FIXME: move from .env here
// var secrets struct {
// 	TodoistApiToken string
// }

// Send Telegram message with projects that has too many and zero active tasks.
var _ = cron.NewJob("incorrect-projects-notifier", cron.JobConfig{
	Title:    "Send Telegram message with projects that has too many and zero active tasks",
	Schedule: "0 8 * * *",
	Endpoint: GetIncorrectProjectsEndpoint,
})

// Move tasks from `Inbox` project to `inbox_archive` if they are older than 3 days.
var _ = cron.NewJob("older-tasks-archivator", cron.JobConfig{
	Title:    "Move tasks from `Inbox` project to `inbox_archive` if they are older than 3 days",
	Schedule: "0 8 * * *",
	Endpoint: ArchiveOlderTasksEndpoint,
})

// ==================================================================

// Next steps
//
// 1. Deploy your application to the cloud
//
//     git add -A .
//     git commit -m 'Commit message'
//     git push encore
//
// 2. To continue exploring Encore, check out one of these topics:
//
//    Building a Slack bot:  https://encore.dev/docs/tutorials/slack-bot
//    Building a REST API:   https://encore.dev/docs/tutorials/rest-api
//    Using SQL databases:   https://encore.dev/docs/develop/databases
//    Authenticating users:  https://encore.dev/docs/develop/auth

//encore:api public path=/projects/incorrect
func GetIncorrectProjectsEndpoint(ctx context.Context) (*IncorrectResponse, error) {
	tooMany, zero := todoist.GetProjectsWithTooManyAndZeroTasks(3)
	combined := IncorrectResponse{
		TooMany: tooMany,
		Zero:    zero,
	}
	return &combined, nil
}

type IncorrectResponse struct {
	TooMany []todoist.ResultUnit `json:"TooMany"`
	Zero    []todoist.ResultUnit `json:"Zero"`
}

//encore:api public method=POST path=/tasks/archive-older
func ArchiveOlderTasksEndpoint(ctx context.Context) (*MoveOlderTasksResponse, error) {
	srcProjectName, dstProjectName, oldThreshold, dryRun := "leisure", "inbox_archive", time.Hour*24*3, true
	tasks := todoist.MoveOlderTasks(srcProjectName, dstProjectName, oldThreshold, dryRun)
	return &MoveOlderTasksResponse{
		Tasks: tasks,
	}, nil
}

type MoveOlderTasksResponse struct {
	Tasks []todoist.Task `json:"tasks"`
}

// FIXME: refactor log.Fatal to returning errors to callers in whole project
