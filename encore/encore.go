package api

import (
	"context"
	"encore.dev/cron"
	api "github.com/valeriikundas/todoist-scripts/api"
	"log"
)

var secrets struct {
	TodoistApiToken string

	TelegramApiToken string
	TelegramUserID   string

	TogglApiToken    string
	TogglWorkspaceID string
}

//encore:service
type Service struct {
	// Add your dependencies here
}

func initService() (*Service, error) {
	log.SetFlags(log.Ltime | log.Lshortfile)

	return &Service{}, nil
}

// Send Telegram message with projects that has too many and zero active tasks.
var _ = cron.NewJob("incorrect-projects-notifier", cron.JobConfig{
	Title:    "Send Telegram message with projects that has too many and zero active tasks",
	Schedule: "0 6 * * *",
	Endpoint: GetIncorrectProjectsEndpoint,
})

// Move tasks from `Inbox` project to `inbox_archive` if they are older than 3 days.
var _ = cron.NewJob("older-tasks-archivator", cron.JobConfig{
	Title:    "Move tasks from `Inbox` project to `inbox_archive` if they are older than 3 days",
	Schedule: "0 6 * * *",
	Endpoint: ArchiveOlderTasksEndpoint,
})

// Ask for Toggl time entry if it is empty.
var _ = cron.NewJob("ask-for-toggl-entry", cron.JobConfig{
	Title:    "Ask for Toggl time entry through Telegram if it is empty. Save to Toggl",
	Schedule: "*/15 5-21 * * *", // Every 15 minutes from 5-21 UTC
	Endpoint: AssertRunningTogglEntryEndpoint,
})

//encore:api private method=GET path=/projects/incorrect
func (s *Service) GetIncorrectProjectsEndpoint(ctx context.Context) (*api.IncorrectResponse, error) {
	return api.GetIncorrectProjects(secrets.TodoistApiToken, secrets.TelegramApiToken, secrets.TelegramUserID)
}

//encore:api private method=POST path=/tasks/archive-older
func (s *Service) ArchiveOlderTasksEndpoint(ctx context.Context) (*api.MoveOlderTasksResponse, error) {
	return api.ArchiveOlderTasks(secrets.TodoistApiToken)
}

// FIXME: refactor log.Fatal to returning errors to callers in whole project

//encore:api private method=POST path=/toggl/assertRunningEntry
func (s *Service) AssertRunningTogglEntryEndpoint(ctx context.Context) (*api.AssertToggleEntryResponse, error) {
	return api.AssertRunningTogglEntry(
		secrets.TogglApiToken,
		secrets.TogglWorkspaceID,
		secrets.TelegramApiToken,
		secrets.TelegramUserID,
	)
}
