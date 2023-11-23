package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// TODO: maybe `get all sections` will be more useful
	projects := todoistListProjects()

	tasks := getTasks()

	nextActionTasks := map[string][]GetTaskSchema{}
	for _, task := range tasks {
		projectName := getProjectNameByProjectID(task.ProjectID, projects)
		if projectName == nil {
			// FIXME: look into it
			// log.Printf("nil projectName for projectid=%v taskName=%v\n", task.ProjectID, task.Content)
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

	nextActionsTasksLimitPerProject := 3

	for projectName, localTasks := range nextActionTasks {
		if len(localTasks) > nextActionsTasksLimitPerProject {
			fmt.Printf("has more @next_action tasks than allowed: projectName=%v tasks=%v\n", projectName, len(localTasks))
			continue
		}
		if len(localTasks) == 0 {
			fmt.Printf("does not have @next_action tasks: projectName=%v\n", projectName)
			continue
		}
	}
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
	headerKey, headerValue := "Authorization", fmt.Sprintf("Bearer %v", token)
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
	url := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks?project_id=%v", projectID)
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
