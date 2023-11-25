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

type Todoist struct {
	apiToken string
}

func NewTodoist(apiToken string) *Todoist {
	return &Todoist{apiToken: apiToken}
}

func (t *Todoist) GetProjectsWithTooManyAndZeroTasks(limit int) (projectsWithTooManyTasks []ResultUnit, projectsWithZeroTasks []ResultUnit) {
	// TODO: maybe `get all sections` will be more useful
	projects := t.getProjectList()
	tasks := t.getTasks()
	nextActionTasks := mapTasksToProjectAndFilterByLabel(projects, tasks)
	projectsWithTooManyTasks, projectsWithZeroTasks = filterProjects(nextActionTasks, limit)
	return projectsWithTooManyTasks, projectsWithZeroTasks
}

func (t *Todoist) MoveOlderTasks(srcProjectName, dstProjectName string, oldThreshold time.Duration, dryRun bool) []Task {
	projects := t.getProjectList()

	srcProject, ok := findProjectByName(projects, srcProjectName)
	if !ok {
		log.Fatalf("did not find `%s` project\n", srcProjectName)
	}

	tasks := t.getProjectTasks(*srcProject)
	oldTasks := filterOldTasks(tasks, oldThreshold)

	dstProject, ok := findProjectByName(projects, dstProjectName)
	if !ok {
		log.Fatalf("did not find `%s` project\n", srcProjectName)
	}

	t.moveTasks(oldTasks, dstProject.ID, dryRun)
	return oldTasks
}

func (t *Todoist) getProjectList() []Project {
	url := "https://api.todoist.com/rest/v2/projects"

	b := t.doTodoistRequest(url)

	var projects []Project
	err := json.Unmarshal(b, &projects)
	if err != nil {
		log.Fatal(err)
	}

	return projects
}

func (t *Todoist) getTasks() []Task {
	url := "https://api.todoist.com/rest/v2/tasks"
	b := t.doTodoistRequest(url)

	var tasks []Task
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

func (t *Todoist) getProjectTasks(project Project) []Task {
	projectID := project.ID
	url := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks?project_id=%s", projectID)
	b := t.doTodoistRequest(url)

	var tasks []Task
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

func (t *Todoist) moveTasks(tasks []Task, projectID string, dryRun bool) {
	// FIXME: rewrite with single todoist sync api request
	for _, task := range tasks {
		t.moveTask(task.ID, projectID, dryRun)
	}
}

func (t *Todoist) doTodoistRequest(url string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	headerKey, headerValue := "Authorization", fmt.Sprintf("Bearer %s", t.apiToken)
	req.Header.Add(headerKey, headerValue)

	// FIXME: construct client once, do requests then
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func (t *Todoist) doTodoistPostRequest(method string, url string, body io.Reader) []byte {
	headerKey, headerValue := "Authorization", fmt.Sprintf("Bearer %s", t.apiToken)

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

func (t *Todoist) moveTask(task_id string, project_id string, dryRun bool) {
	logMessage := fmt.Sprintf("moving task_id=%s to project_id=%s\n", task_id, project_id)
	if dryRun {
		log.Printf("dry run: %v", logMessage)
		return
	}
	log.Println(logMessage)

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

	resp := t.doTodoistPostRequest(http.MethodPost, "https://api.todoist.com/sync/v9/sync", bodyReader)

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
