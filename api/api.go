package api

import (
	"errors"
	"github.com/valeriikundas/todoist-scripts/telegram"
	"github.com/valeriikundas/todoist-scripts/todoist"
	"github.com/valeriikundas/todoist-scripts/toggl"
	"log"
	"strconv"
	"time"
)

func SendReportAboutIncorrectProjectsToTelegram(todoistApiToken string, telegramApiToken string, telegramUserIDString string) (*IncorrectResponse, error) {
	todoistClient := todoist.NewClient(todoistApiToken)
	tooMany, zero := todoistClient.GetProjectsWithTooManyAndZeroTasks(3)
	combined := IncorrectResponse{
		TooMany: tooMany,
		Zero:    zero,
	}

	tg := telegram.NewTelegram(telegramApiToken)
	message := todoistClient.PrettyOutput(tooMany, zero)
	telegramUserID, err := strconv.Atoi(telegramUserIDString)
	if err != nil {
		return nil, err
	}

	err = tg.Send(telegramUserID, message)
	if err != nil {
		return nil, err
	}

	return &combined, nil
}

type IncorrectResponse struct {
	TooMany []todoist.IncorrectProjectSchema `json:"TooMany"`
	Zero    []todoist.IncorrectProjectSchema `json:"Zero"`
}

func ArchiveOlderInboxTasks(todoistApiToken string) (*MoveOlderTasksResponse, error) {
	todoist := todoist.NewClient(todoistApiToken)
	dstProjectName, oldThreshold, dryRun := "inbox_archive", time.Hour*24*3, false
	tasks := todoist.MoveOlderTasks("Inbox", dstProjectName, oldThreshold, dryRun)
	return &MoveOlderTasksResponse{
		Tasks: tasks,
	}, nil
}

type MoveOlderTasksResponse struct {
	Tasks []todoist.Task `json:"tasks"`
}

func AssertRunningTogglEntry(togglApiToken string, togglWorkspaceID string, telegramApiToken string, telegramUserIDString string) (*AssertToggleEntryResponse, error) {
	telegramUserID, err := strconv.Atoi(telegramUserIDString)
	if err != nil {
		return nil, err
	}

	isEmpty, timeEntry, err := toggl.AskForTogglEntryIfEmpty(togglApiToken, telegramApiToken, telegramUserID)
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

	togglClient := toggl.NewToggl(togglApiToken)

	err = togglClient.StartTimeEntry(timeEntry, togglWorkspaceID)
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
