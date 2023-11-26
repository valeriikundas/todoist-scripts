package api

import (
	"context"
	"time"

	"encore.dev/cron"
	"github.com/valeriikundas/todoist-scripts/telegram"
	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
)

var secrets struct {
	TodoistApiToken  string
	TelegramApiToken string
	TelegramUserID   string
}

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

//encore:api private method=GET path=/projects/incorrect
func GetIncorrectProjectsEndpoint(ctx context.Context) (*IncorrectResponse, error) {
	todoist := todoist.NewTodoist(secrets.TodoistApiToken)
	tooMany, zero := todoist.GetProjectsWithTooManyAndZeroTasks(3)
	combined := IncorrectResponse{
		TooMany: tooMany,
		Zero:    zero,
	}

	tg := telegram.NewTelegram(secrets.TelegramApiToken)
	message := todoist.PrettyOutput(tooMany, zero)
	err := tg.Send(secrets.TelegramUserID, message)
	if err != nil {
		return nil, err
	}

	return &combined, nil
}

type IncorrectResponse struct {
	TooMany []todoist.IncorrectProjectSchema `json:"TooMany"`
	Zero    []todoist.IncorrectProjectSchema `json:"Zero"`
}

//encore:api private method=POST path=/tasks/archive-older
func ArchiveOlderTasksEndpoint(ctx context.Context) (*MoveOlderTasksResponse, error) {
	todoist := todoist.NewTodoist(secrets.TodoistApiToken)
	srcProjectName, dstProjectName, oldThreshold, dryRun := "Inbox", "inbox_archive", time.Hour*24*3, false
	tasks := todoist.MoveOlderTasks(srcProjectName, dstProjectName, oldThreshold, dryRun)
	return &MoveOlderTasksResponse{
		Tasks: tasks,
	}, nil
}

type MoveOlderTasksResponse struct {
	Tasks []todoist.Task `json:"tasks"`
}

// FIXME: refactor log.Fatal to returning errors to callers in whole project
