package api

import (
	"context"
	"log"
	"time"

	todoist "github.com/valeriikundas/todoist-scripts/todoist_utils"
)


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
