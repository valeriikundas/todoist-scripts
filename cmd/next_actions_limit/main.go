package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	showProjectsWithTooManyNextActionTasks()
}

func showProjectsWithTooManyNextActionTasks() {
	// TODO: maybe `get all sections` will be more useful
	projects := todoistListProjects()
	tasks := getTasks()
	nextActionTasks := mapTasksToProjectAndFilterByLabel(projects, tasks)
	nextActionsTasksLimitPerProject := 3
	printOutput(nextActionTasks, nextActionsTasksLimitPerProject)
}

func printOutput(nextActionTasks map[string][]GetTaskSchema, nextActionsTasksLimitPerProject int) {
	// FIXME: split filtering and outputting
	for projectName, projectTasks := range nextActionTasks {
		if len(projectTasks) > nextActionsTasksLimitPerProject {
			filterLabel := "next_action"
			url := getTasksURL(projectName, &filterLabel)
			fmt.Printf("project \"%s\" has %d @next_action tasks, max allowed is %d. ", projectName, len(projectTasks), nextActionsTasksLimitPerProject)
			fmt.Printf("please review and fix at %s\n", url)
			continue
		}

		if len(projectTasks) == 0 {
			url := getTasksURL(projectName, nil)
			fmt.Printf("does not have @next_action tasks: projectName=%s.", projectName)
			fmt.Printf("please review and fix at %s\n", url)
			continue
		}
	}
}

func mapTasksToProjectAndFilterByLabel(projects []GetProjectSchema, tasks []GetTaskSchema) map[string][]GetTaskSchema {
	// FIXME: split into 2 steps: 1. filter tasks by label 2. map tasks to project
	// FIXME: move tasks filter to API query
	// FIXME: rewrite to map[projecID]Task
	nextActionTasks := map[string][]GetTaskSchema{}
	for _, task := range tasks {
		projectName := getProjectNameByProjectID(task.ProjectID, projects)
		if projectName == nil {
			// FIXME: look into it
			// log.Printf("nil projectName for projectid=%s taskName=%s\n", task.ProjectID, task.Content)
			continue
		}

		contains := false
		for _, l := range task.Labels {
			if l == "next_action" {
				contains = true
				break
			}
		}
		if contains {
			nextActionTasks[*projectName] = append(nextActionTasks[*projectName], task)
		}
	}
	return nextActionTasks
}

func getTasksURL(projectName string, label *string) string {
	var query string
	if label == nil {
		query = fmt.Sprintf("#%s", projectName)
	} else {
		query = fmt.Sprintf("@%s&#%s", *label, projectName)
	}
	escapedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://todoist.com/app/search/%s", escapedQuery)
	return url
}

func todoistListProjects() []GetProjectSchema {
	url := "https://api.todoist.com/rest/v2/projects"

	b := doTodoistRequest(url)

	var projects []GetProjectSchema
	err := json.Unmarshal(b, &projects)
	if err != nil {
		log.Fatal(err)
	}

	return projects
}

func doTodoistRequest(url string) []byte {
	token := getApiToken()
	headerKey, headerValue := "Authorization", fmt.Sprintf("Bearer %s", token)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add(headerKey, headerValue)
	if err != nil {
		log.Fatal(err)
	}

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

func getApiToken() string {
	// FIXME: rewrite with `github.com/joho/godotenv`
	file, err := os.Open(".env")
	if err != nil {
		log.Fatal(err)
	}
	b, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	text := string(b)

	split := strings.Split(text, "=")
	apiToken := split[1]
	return apiToken
}

type GetProjectSchema struct {
	ID   string
	Name string
	Url  string
}

func getProjectTasks(project GetProjectSchema) []GetTaskSchema {
	projectID := project.ID
	url := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks?project_id=%s", projectID)
	b := doTodoistRequest(url)

	var tasks []GetTaskSchema
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

func getTasks() []GetTaskSchema {
	url := "https://api.todoist.com/rest/v2/tasks"
	b := doTodoistRequest(url)

	var tasks []GetTaskSchema
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

type GetTaskSchema struct {
	ID          string
	ProjectID   string `json:"project_id"`
	Content     string
	Description string
	Labels      []string
}

func getProjectNameByProjectID(projectID string, projects []GetProjectSchema) *string {
	for _, p := range projects {
		if p.ID == projectID {
			return &p.Name
		}
	}
	return nil
}
