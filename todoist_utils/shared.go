package todoist_utils

import (
	"time"
)

type Command struct {
	Type string `json:"type"`
	Args X      `json:"args"`
	Uuid string `json:"uuid"`
}

type X struct {
	Id        string `json:"id"`
	ProjectID string `json:"project_id"`
}

func filterOldTasks(tasks []Task, duration time.Duration) []Task {
	filteredTasks := make([]Task, 0, len(tasks))

	for _, t := range tasks {
		oldTaskThreshold := time.Now().Add(-duration)
		if t.CreatedAt.Compare(oldTaskThreshold) == -1 {
			filteredTasks = append(filteredTasks, t)
		}
	}
	return filteredTasks
}

func findProjectByName(projects []Project, projectName string) (*Project, bool) {
	for _, p := range projects {
		if p.Name == projectName {
			return &p, true
		}
	}
	return nil, false
}
