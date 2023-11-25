package todoist_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func GetProjectsWithTooManyAndZeroTasks(limit int) (projectsWithTooManyTasks []ResultUnit, projectsWithZeroTasks []ResultUnit) {
	// TODO: maybe `get all sections` will be more useful
	projects := TodoistListProjects()
	tasks := getTasks()
	nextActionTasks := mapTasksToProjectAndFilterByLabel(projects, tasks)
	projectsWithTooManyTasks, projectsWithZeroTasks = filterProjects(nextActionTasks, limit)
	return projectsWithTooManyTasks, projectsWithZeroTasks
}

func PrintOutput(aboveLimitProjects []ResultUnit, zeroTaskProjects []ResultUnit) {
	for _, p := range aboveLimitProjects {
		fmt.Printf("project \"%s\" has %d @next_action tasks, max allowed is %d. ",
			p.projectName, p.tasksCount, p.limit)
		fmt.Printf("please review and fix at %s\n", p.url)
	}

	for _, p := range zeroTaskProjects {
		fmt.Printf("does not have @next_action tasks: projectName=%s.", p.projectName)
		fmt.Printf("please review and fix at %s\n", p.url)
	}
}

type ResultUnit struct {
	projectName string
	tasksCount  int
	url         string
	limit       int
}

type ProjectID string

func filterProjects(nextActionTasks map[string][]Task, nextActionsTasksLimitPerProject int) ([]ResultUnit, []ResultUnit) {
	// FIXME: split into 2 different filter functions
	projectsWithTooManyTasks := []ResultUnit{}
	projectsWithZeroTasks := []ResultUnit{}

	for projectName, projectTasks := range nextActionTasks {
		if len(projectTasks) > nextActionsTasksLimitPerProject {
			filterLabel := "next_action"
			url := getTasksURL(projectName, &filterLabel)
			projectsWithTooManyTasks = append(projectsWithTooManyTasks, ResultUnit{
				projectName: projectName,
				tasksCount:  len(projectTasks),
				url:         url,
				limit:       nextActionsTasksLimitPerProject,
			})
		}
		if len(projectTasks) == 0 {
			url := getTasksURL(projectName, nil)
			projectsWithZeroTasks = append(projectsWithZeroTasks, ResultUnit{
				projectName: projectName,
				tasksCount:  0,
				url:         url,
				limit:       nextActionsTasksLimitPerProject,
			})
		}
	}

	return projectsWithTooManyTasks, projectsWithZeroTasks
}

func mapTasksToProjectAndFilterByLabel(projects []Project, tasks []Task) map[string][]Task {
	// FIXME: split into 2 steps: 1. filter tasks by label 2. map tasks to project
	// FIXME: move tasks filter to API query
	// FIXME: rewrite to map[projecID]Task
	nextActionTasks := map[string][]Task{}
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

func TodoistListProjects() []Project {
	url := "https://api.todoist.com/rest/v2/projects"

	b := DoTodoistRequest(url)

	var projects []Project
	err := json.Unmarshal(b, &projects)
	if err != nil {
		log.Fatal(err)
	}

	return projects
}

func DoTodoistRequest(url string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	token := GetApiToken()
	headerKey, headerValue := "Authorization", fmt.Sprintf("Bearer %s", token)
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

func GetApiToken() string {
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

type Project struct {
	ID   string
	Name string
	Url  string
}

func GetProjectTasks(project Project) []Task {
	projectID := project.ID
	url := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks?project_id=%s", projectID)
	b := DoTodoistRequest(url)

	var tasks []Task
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

func getTasks() []Task {
	url := "https://api.todoist.com/rest/v2/tasks"
	b := DoTodoistRequest(url)

	var tasks []Task
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

type Task struct {
	ID          string
	ProjectID   string `json:"project_id"`
	Content     string
	Description string
	Labels      []string
	CreatedAt   TimeParser `json:"created_at"`
}

type TimeParser struct {
	time.Time
}

func (tp *TimeParser) UnmarshalJSON(b []byte) (err error) {
	time, err := time.Parse(`"2006-01-02T15:04:05.000000Z"`, string(b))
	if err != nil {
		return err
	}
	tp.Time = time
	return
}

func getProjectNameByProjectID(projectID string, projects []Project) *string {
	for _, p := range projects {
		if p.ID == projectID {
			return &p.Name
		}
	}
	return nil
}
