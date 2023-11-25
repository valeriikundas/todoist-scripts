package todoist_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func MoveOlderTasks(srcProjectName, dstProjectName string, oldThreshold time.Duration, dryRun bool) []Task {
	projects := GetProjectList()

	srcProject, ok := findProjectByName(projects, srcProjectName)
	if !ok {
		log.Fatalf("did not find `%s` project\n", srcProjectName)
	}

	tasks := GetProjectTasks(*srcProject)
	oldTasks := filterOldTasks(tasks, oldThreshold)

	dstProject, ok := findProjectByName(projects, dstProjectName)
	if !ok {
		log.Fatalf("did not find `%s` project\n", srcProjectName)
	}

	moveTasks(oldTasks, dstProject.ID, dryRun)
	return oldTasks
}

func moveTasks(tasks []Task, projectID string, dryRun bool) {
	// FIXME: rewrite with single todoist sync api request
	for _, t := range tasks {
		moveTask(t.ID, projectID, dryRun)
	}
}

type Command struct {
	Type string `json:"type"`
	Args X      `json:"args"`
	Uuid string `json:"uuid"`
}

type X struct {
	Id        string `json:"id"`
	ProjectID string `json:"project_id"`
}

func moveTask(task_id string, project_id string, dryRun bool) {
	if dryRun {
		log.Printf("moving task_id=%s to project_id=%s\n", task_id, project_id)
		return
	}

	commands := []Command{
		{
			Type: "item_move",
			Args: X{
				Id:        task_id,
				ProjectID: project_id,
			},
			Uuid: uuid.New().String(),
		},
	}

	b, err := json.Marshal(commands)
	if err != nil {
		log.Fatal(err)
	}
	bodyReader := strings.NewReader(fmt.Sprintf("commands=%+s", string(b)))

	resp := DoTodoistPostRequest(http.MethodPost, "https://api.todoist.com/sync/v9/sync", bodyReader)

	var v map[string]interface{}
	err = json.Unmarshal(resp, &v)
	if err != nil {
		log.Fatal(err)
	}

	status, ok := v["sync_status"]
	if !ok {
		log.Fatal("todoist request failed")
	}

	log.Print(status)
}

func DoTodoistPostRequest(method string, url string, body io.Reader) []byte {
	token := GetApiToken()
	headerKey, headerValue := "Authorization", fmt.Sprintf("Bearer %s", token)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add(headerKey, headerValue)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// FIXME: construct client once, do requests then
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("todoist request failed, url=%s, code=%d, body=%v\n", url, resp.StatusCode, string(body))
	}

	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return resultBytes
}

func filterOldTasks(tasks []Task, duration time.Duration) []Task {
	filteredTasks := make([]Task, len(tasks))
	k := 0

	for _, t := range tasks {
		oldTaskThreshold := time.Now().Add(-duration)
		if t.CreatedAt.Compare(oldTaskThreshold) == -1 {
			filteredTasks[k] = t
			k += 1
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
