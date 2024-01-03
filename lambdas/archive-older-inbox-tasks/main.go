package main

import (
	"github.com/valeriikundas/todoist-scripts/api"
	lambdacommon "github.com/valeriikundas/todoist-scripts/lambdas"
)

func f(secrets *lambdacommon.Secrets) (*api.MoveOlderTasksResponse, error) {
	return api.ArchiveOlderInboxTasks(secrets.TodoistApiToken)
}

func main() {
	lambdacommon.Run(f)
}
