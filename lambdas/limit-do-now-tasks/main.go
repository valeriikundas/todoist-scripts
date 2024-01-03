package main

import (
	"github.com/pkg/errors"
	"github.com/valeriikundas/todoist-scripts/api"
	lambdacommon "github.com/valeriikundas/todoist-scripts/lambdas"
	"os"
	"strings"
)

func f(secrets *lambdacommon.Secrets) (*api.IncorrectResponse, error) {
	excludeFromZeroProjectsString, ok := os.LookupEnv("ExcludeFromZeroProjectsList")
	if !ok {
		return nil, errors.New("ExcludeFromZeroProjectsList environment variable is not set")
	}
	excludeFromZeroProjectsList := strings.Split(excludeFromZeroProjectsString, ";")

	return api.SendReportAboutIncorrectProjectsToTelegram(
		secrets.TodoistApiToken,
		secrets.TelegramApiToken,
		secrets.TelegramUserID,
		excludeFromZeroProjectsList,
	)
}

func main() {
	lambdacommon.Run(f)
}
