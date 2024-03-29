package todoist

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const PriorityThreshold = 3

type Client struct {
	apiToken string
}

func NewClient(apiToken string) *Client {
	return &Client{apiToken: apiToken}
}

func (t *Client) GetProjectsWithTooManyAndZeroTasks(limit int, excludeFromZeroProjectsList []string) (
	projectsWithTooManyTasks []IncorrectProjectSchema,
	projectsWithZeroTasks []IncorrectProjectSchema,
) {
	// TODO: maybe `get all sections` will be more useful
	projects := t.getProjectList()
	// fixme: getTasks() fetches too many tasks, filter in some way, can fetch by label straight away or fetch by project
	tasks := t.getTasks()
	nextActionTasks := t.mapTasksToProjectAndFilterByLabel(projects, tasks)

	projectsWithTooManyTasks = t.filterProjects(nextActionTasks, limit)

	projectsWithZeroTasks = make([]IncorrectProjectSchema, 0, 100)
	for _, project := range projects {
		if slices.Contains(excludeFromZeroProjectsList, project.Name) {
			continue
		}

		_, ok := nextActionTasks[project.Name]
		if !ok {
			projectsWithZeroTasks = append(projectsWithZeroTasks, IncorrectProjectSchema{
				ProjectName: project.Name,
				TasksCount:  0,
				URL:         project.Url,
				Limit:       0,
				Description: "no active tasks on this project",
			})
		}
	}

	return projectsWithTooManyTasks, projectsWithZeroTasks
}

type IncorrectProjectSchema struct {
	ProjectName string `json:"projectName"`
	TasksCount  int    `json:"tasksCount"`
	URL         string `json:"url"`
	Limit       int    `json:"limit,omitempty"`
	Description string `json:"description"`
}

// MoveInactiveTasks moves tasks from one project to another that are older and have low priority
func (t *Client) MoveInactiveTasks(srcProjectName, dstProjectName string, oldThreshold time.Duration, dryRun bool) []Task {
	projects := t.getProjectList()

	srcProject, ok := t.findProjectByName(projects, srcProjectName)
	if !ok {
		// fixme: return errors instead of log.Fatal in all project
		log.Fatalf("did not find `%s` project\n", srcProjectName)
	}

	tasks := t.getProjectTasks(*srcProject)
	filteredTasks := t.filterTasksByCreationTime(tasks, oldThreshold)
	filteredTasks = t.filterByPriority(filteredTasks, PriorityThreshold)

	dstProject, ok := t.findProjectByName(projects, dstProjectName)
	if !ok {
		log.Fatalf("did not find `%s` project\n", srcProjectName)
	}

	t.moveTasks(filteredTasks, dstProject.ID, dryRun)
	return filteredTasks
}

func (c *Client) filterByPriority(tasks []Task, priority int) []Task {
	filteredTasks := make([]Task, 0)
	for _, t := range tasks {
		if t.Priority < priority {
			filteredTasks = append(filteredTasks, t)
		}
	}
	return filteredTasks
}

func (t *Client) getProjectList() []Project {
	projectsUrl := "https://api.todoist.com/rest/v2/projects"

	b := t.doTodoistRequest(projectsUrl)

	var projects []Project
	err := json.Unmarshal(b, &projects)
	if err != nil {
		log.Fatal(err)
	}

	return projects
}

func (t *Client) getTasks() []Task {
	tasksUrl := "https://api.todoist.com/rest/v2/tasks"
	b := t.doTodoistRequest(tasksUrl)

	var tasks []Task
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

func (t *Client) getProjectTasks(project Project) []Task {
	projectID := project.ID
	projectTasksUrl := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks?project_id=%s", projectID)
	b := t.doTodoistRequest(projectTasksUrl)

	var tasks []Task
	err := json.Unmarshal(b, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

func (t *Client) moveTasks(tasks []Task, projectID string, dryRun bool) {
	// FIXME: rewrite with single todoist sync api request
	for _, task := range tasks {
		t.moveTask(task.ID, projectID, dryRun)
	}
}

func (t *Client) doTodoistRequest(url string) []byte {
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

func (t *Client) doTodoistPostRequest(method string, url string, body io.Reader) []byte {
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

func (t *Client) moveTask(taskID string, projectID string, dryRun bool) {
	logMessage := fmt.Sprintf("moving task_id=%s to project_id=%s\n", taskID, projectID)
	if dryRun {
		log.Printf("dry run: %v", logMessage)
		return
	}
	log.Println(logMessage)

	commands := []Command{
		{
			Type: "item_move",
			Args: MoveCommandArgs{
				Id:        taskID,
				ProjectID: projectID,
			},
			Uuid: uuid.New().String(),
		},
	}

	b, err := json.Marshal(commands)
	if err != nil {
		log.Fatal(err)
	}
	bodyReader := strings.NewReader(fmt.Sprintf("commands=%s", string(b)))

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

func (c *Client) filterTasksByCreationTime(tasks []Task, duration time.Duration) []Task {
	filteredTasks := make([]Task, 0, len(tasks))

	oldTaskThreshold := time.Now().Add(-duration)
	for _, t := range tasks {
		if t.CreatedAt.Compare(oldTaskThreshold) == -1 {
			filteredTasks = append(filteredTasks, t)
		}
	}
	return filteredTasks
}

func (t *Client) findProjectByName(projects []Project, projectName string) (*Project, bool) {
	for _, p := range projects {
		if p.Name == projectName {
			return &p, true
		}
	}
	return nil, false
}

func (t *Client) filterProjects(nextActionTasks map[string][]Task, nextActionsTasksLimitPerProject int) []IncorrectProjectSchema {
	projectsWithTooManyTasks := make([]IncorrectProjectSchema, 0, 10)

	for projectName, projectTasks := range nextActionTasks {
		if len(projectTasks) > nextActionsTasksLimitPerProject {
			filterLabel := "next_action"
			tasksUrl := t.getTasksURL(projectName, &filterLabel)
			projectsWithTooManyTasks = append(projectsWithTooManyTasks, IncorrectProjectSchema{
				ProjectName: projectName,
				TasksCount:  len(projectTasks),
				URL:         tasksUrl,
				Limit:       nextActionsTasksLimitPerProject,
				Description: "Project has more active tasks that allowed",
			})
		}
	}

	return projectsWithTooManyTasks
}

func (t *Client) mapTasksToProjectAndFilterByLabel(projects []Project, tasks []Task) map[string][]Task {
	// FIXME: split into 2 steps: 1. filter tasks by label 2. map tasks to project
	// FIXME: move tasks filter to API query
	// FIXME: rewrite to map[projectID]Task
	nextActionTasks := map[string][]Task{}
	for _, task := range tasks {
		projectName := t.getProjectNameByProjectID(task.ProjectID, projects)
		if projectName == nil {
			// FIXME: look into it
			// log.Printf("nil projectName for projectID=%s taskName=%s\n", task.ProjectID, task.Content)
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

func (t *Client) getProjectNameByProjectID(projectID string, projects []Project) *string {
	for _, p := range projects {
		if p.ID == projectID {
			return &p.Name
		}
	}
	return nil
}

func (t *Client) getTasksURL(projectName string, label *string) string {
	var query string
	if label == nil {
		query = fmt.Sprintf("#%s", projectName)
	} else {
		query = fmt.Sprintf("@%s&#%s", *label, projectName)
	}
	escapedQuery := url.QueryEscape(query)
	tasksUrl := fmt.Sprintf("https://todoist.com/app/search/%s", escapedQuery)
	return tasksUrl
}

func (t *Client) PrettyOutput(projectsWithTooManyTasks []IncorrectProjectSchema, projectsWithZeroTasks []IncorrectProjectSchema) string {
	builder := strings.Builder{}

	if len(projectsWithTooManyTasks) > 0 {
		builder.WriteString("projects with too many @next_action tasks:\n")
		for _, p := range projectsWithTooManyTasks {
			builder.WriteString(fmt.Sprintf("%d - [%s](%s)\n", p.TasksCount, p.ProjectName, p.URL))
		}
	}

	if len(projectsWithZeroTasks) > 0 {
		builder.WriteString("\n")
		builder.WriteString("projects without @next_action tasks:\n")
		for _, p := range projectsWithZeroTasks {
			builder.WriteString(fmt.Sprintf("[%s](%s)\n", p.ProjectName, p.URL))
		}
	}

	return builder.String()
}
