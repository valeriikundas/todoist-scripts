package todoist_utils

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

func PrintOutput(aboveLimitProjects []ResultUnit, zeroTaskProjects []ResultUnit) {
	for _, p := range aboveLimitProjects {
		fmt.Printf("project \"%s\" has %d @next_action tasks, max allowed is %d. ",
			p.ProjectName, p.TasksCount, p.Limit)
		fmt.Printf("please review and fix at %s\n", p.URL)
	}

	for _, p := range zeroTaskProjects {
		fmt.Printf("does not have @next_action tasks: projectName=%s.", p.ProjectName)
		fmt.Printf("please review and fix at %s\n", p.URL)
	}
}

type ResultUnit struct {
	ProjectName string `json:"projectName"`
	TasksCount  int    `json:"tasksCount"`
	URL         string `json:"url"`
	Limit       int    `json:"limit,omitempty"`
	Description string `json:"description"`
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
				ProjectName: projectName,
				TasksCount:  len(projectTasks),
				URL:         url,
				Limit:       nextActionsTasksLimitPerProject,
				Description: "Project has more active tasks that allowed",
			})
		}
		if len(projectTasks) == 0 {
			url := getTasksURL(projectName, nil)
			projectsWithZeroTasks = append(projectsWithZeroTasks, ResultUnit{
				ProjectName: projectName,
				TasksCount:  0,
				URL:         url,
				Limit:       nextActionsTasksLimitPerProject,
				Description: "Project has no active tasks",
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

func ReadApiTokenFromDotenv() string {
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

type Task struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"project_id"`
	Content   string     `json:"content"`
	Labels    []string   `json:"labels"`
	CreatedAt TimeParser `json:"created_at"`
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
