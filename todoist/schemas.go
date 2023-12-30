package todoist

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

type Command struct {
	Type string          `json:"type"`
	Args MoveCommandArgs `json:"args"`
	Uuid string          `json:"uuid"`
}

type MoveCommandArgs struct {
	Id        string `json:"id"`
	ProjectID string `json:"project_id"`
}
