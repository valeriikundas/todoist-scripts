package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valeriikundas/todoist-scripts/todoist_utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	projects := todoist_utils.TodoistListProjects()

	inboxProject, ok := findProjectByName(projects, "Inbox")
	if !ok {
		log.Fatalln("did not find `inbox` project")
	}

	oldTasks := todoist_utils.GetProjectTasks(*inboxProject)

	// oldTasks := filterOldTasks(tasks, time.Hour*24*3)

	inboxArchiveProjectID, ok := findProjectByName(projects, "inbox_archive")
	if !ok {
		log.Fatalln("did not find `inbox_archive` project")
	}
	log.Printf("inboxArchiveProjectID=%v\n", inboxArchiveProjectID.ID)

	moveTasks(oldTasks, inboxArchiveProjectID.ID)
}

func moveTasks(tasks []todoist_utils.Task, projectID string) {
	// TODO: move task solution https://github.com/Doist/todoist-api-python/issues/8#issuecomment-1344860782

	for _, t := range tasks {
		log.Printf("will move %v", t.Content)
		moveTaskSyncApi(t.ID, projectID)
		break
	}
}

type R struct {
	Commands []Command `json:"commands"`
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

func moveTaskSyncApi(task_id string, project_id string) {
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

	resp := DoTodoistRequestV2("https://api.todoist.com/sync/v9/sync", bodyReader)
	log.Print(string(resp))
}

func DoTodoistRequestV2(url string, body io.Reader) []byte {
	token := todoist_utils.GetApiToken()
	headerKey, headerValue := "Authorization", fmt.Sprintf("Bearer %s", token)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add(headerKey, headerValue)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Add("X-Request-Id", "$(uuidgen)")

	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("REQUEST:\n%s\n", b)

	// FIXME: construct client once, do requests then
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("resp.StatusCode=%v\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		bd, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("code=%d body=%v\n", resp.StatusCode, string(bd))
	}

	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return resultBytes
}

func filterOldTasks(tasks []todoist_utils.Task, duration time.Duration) []todoist_utils.Task {
	result := make([]todoist_utils.Task, len(tasks))
	k := 0

	for _, t := range tasks {
		threeDaysAgo := time.Now().Add(-3 * 24 * time.Hour)
		if t.CreatedAt.Compare(threeDaysAgo) == -1 {
			result[k] = t
			k += 1
		}
	}
	return result
}

func findProjectByName(projects []todoist_utils.Project, projectName string) (*todoist_utils.Project, bool) {
	for _, p := range projects {
		if p.Name == projectName {
			return &p, true
		}
	}
	return nil, false
}
