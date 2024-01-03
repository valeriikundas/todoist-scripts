package main

import (
	"github.com/pkg/errors"
	"github.com/valeriikundas/todoist-scripts/api"
	"github.com/valeriikundas/todoist-scripts/lambdas"
	"os"
	"strings"
)

func f(secrets *lambda_common.Secrets) (*api.IncorrectResponse, error) {
	excludeFromZeroProjectsString, ok := os.LookupEnv("ExcludeFromZeroProjectsList")
	if !ok {
		return nil, errors.New("ExcludeFromZeroProjectsList environment variable is not set")
	}
	excludeFromZeroProjectsList := strings.Split(excludeFromZeroProjectsString, ";")

	resp, err := api.SendReportAboutIncorrectProjectsToTelegram(
		secrets.TodoistApiToken,
		secrets.TelegramApiToken,
		secrets.TelegramUserID,
		excludeFromZeroProjectsList,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get incorrect projects")
	}

	return resp, nil
}

func main() {
	lambda_common.Run(f)
}
