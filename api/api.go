package api

import (
	"github.com/valeriikundas/todoist-scripts/telegram"
	"github.com/valeriikundas/todoist-scripts/todoist_utils"
	"strconv"
)

type IncorrectResponse struct {
	TooMany []todoist_utils.IncorrectProjectSchema `json:"TooMany"`
	Zero    []todoist_utils.IncorrectProjectSchema `json:"Zero"`
}

func GetIncorrectProjects(todoistApiToken string, telegramApiToken string, telegramUserIDString string) (*IncorrectResponse, error) {
	todoist := todoist_utils.NewTodoist(todoistApiToken)
	tooMany, zero := todoist.GetProjectsWithTooManyAndZeroTasks(3)
	combined := IncorrectResponse{
		TooMany: tooMany,
		Zero:    zero,
	}

	tg := telegram.NewTelegram(telegramApiToken)
	message := todoist.PrettyOutput(tooMany, zero)
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
