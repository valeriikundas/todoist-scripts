package api

import (
	"context"
	"errors"
	"github.com/valeriikundas/todoist-scripts/telegram"
	"log"
	"strconv"
	"time"

	"encore.dev/cron"
	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
	"github.com/valeriikundas/todoist-scripts/toggl"
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
func (s *Service) GetIncorrectProjectsEndpoint(ctx context.Context) (*IncorrectResponse, error) {
	todoistApiToken := secrets.TodoistApiToken
	telegramApiToken := secrets.TelegramApiToken
	return GetIncorrectProjects(todoistApiToken, telegramApiToken)
}

type IncorrectResponse struct {
	TooMany []todoist.IncorrectProjectSchema `json:"TooMany"`
	Zero    []todoist.IncorrectProjectSchema `json:"Zero"`
}

//encore:api private method=POST path=/tasks/archive-older
func (s *Service) ArchiveOlderTasksEndpoint(ctx context.Context) (*MoveOlderTasksResponse, error) {
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

//encore:api private method=POST path=/toggl/assertRunningEntry
func (s *Service) AssertRunningTogglEntryEndpoint(ctx context.Context) (*AssertToggleEntryResponse, error) {
	telegramUserID, err := strconv.Atoi(secrets.TelegramUserID)
	if err != nil {
		return nil, err
	}

	isEmpty, timeEntry, err := toggl.AskForTogglEntryIfEmpty(secrets.TogglApiToken, secrets.TelegramApiToken, telegramUserID)
	if err != nil {
		if errors.Is(err, &toggl.TelegramTimeoutError{}) {
			return &AssertToggleEntryResponse{
				Reason: ReasonTimeout,
			}, nil
		}

		return nil, err
	}
	log.Printf("Toggl: isEmpty=%v timeEntry=%v", isEmpty, timeEntry)
	if !isEmpty {
		log.Printf("timeEntry is not empty, skipping")
		return &AssertToggleEntryResponse{
			Reason:    ReasonRunning,
			TimeEntry: "",
		}, nil
	}

	togglClient := toggl.NewToggl(secrets.TogglApiToken)
	err = togglClient.StartTimeEntry(timeEntry, secrets.TogglWorkspaceID)
	if err != nil {
		return nil, err
	}

	log.Printf("Toggl: started time entry %v", timeEntry)
	return &AssertToggleEntryResponse{
		Reason:    ReasonUserStarted,
		TimeEntry: timeEntry,
	}, nil
}

type AssertToggleEntryResponse struct {
	Reason    Reason `json:"reason"`
	TimeEntry string `json:"time_entry"`
}

type Reason string

const (
	ReasonRunning     Reason = "running"
	ReasonUserStarted Reason = "user-started"
	ReasonTimeout     Reason = "timeout"
)

// todo: will be moved

func GetIncorrectProjects(todoistApiToken string, telegramApiToken string) (*IncorrectResponse, error) {
	todoist := todoist.NewTodoist(todoistApiToken)
	tooMany, zero := todoist.GetProjectsWithTooManyAndZeroTasks(3)
	combined := IncorrectResponse{
		TooMany: tooMany,
		Zero:    zero,
	}

	tg := telegram.NewTelegram(telegramApiToken)
	message := todoist.PrettyOutput(tooMany, zero)
	telegramUserID, err := strconv.Atoi(secrets.TelegramUserID)
	if err != nil {
		return nil, err
	}

	err = tg.Send(telegramUserID, message)
	if err != nil {
		return nil, err
	}

	return &combined, nil
}
