package toggl

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"encore.dev/storage/cache"
	"github.com/valeriikundas/todoist-scripts/telegram"
)

func AskForTogglEntryIfEmpty(togglApiToken string, telegramApiToken string, telegramUserID int, TelegramGetUpdatesOffset *cache.IntKeyspace[int]) (isEmpty bool, textTimeEntry string, err error) {
	toggl := NewToggl(togglApiToken)
	timeEntry, err := toggl.getCurrentTimeEntry()
	if err != nil {
		return false, "", err
	}

	if timeEntry == nil {
		log.Print("No Toggl time entry found")

		entry, err := AskForTogglEntryInTelegram(telegramApiToken, telegramUserID, TelegramGetUpdatesOffset)
		if err != nil {
			return true, "", err
		}
		return true, entry, nil
	}

	return false, "", nil
}

func NotifyIfNoRunningTogglEntry(togglApiToken string, telegramApiToken string, telegramUserID int) error {
	toggl := NewToggl(togglApiToken)
	timeEntry, err := toggl.getCurrentTimeEntry()
	if err != nil {
		return err
	}
	log.Printf("%#v", timeEntry)

	if timeEntry == nil {
		tg := telegram.NewTelegram(telegramApiToken)
		err := tg.Send(telegramUserID, "No Toggl time entry found")
		if err != nil {
			return err
		}
	}

	return nil
}

type Toggl struct {
	apiToken string
}

func NewToggl(apiToken string) Toggl {
	return Toggl{
		apiToken: apiToken,
	}
}

func (t Toggl) getCurrentTimeEntry() (*TimeEntry, error) {
	url := "https://api.track.toggl.com/api/v9/me/time_entries/current"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// req.Header.Add("content-type", "application/json")
	req.SetBasicAuth(t.apiToken, "api_token")

	c := http.Client{}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	log.Print(resp.StatusCode)
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var v TimeEntry
	err = json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}

	if v.ID == 0 {
		return nil, err
	}

	return &v, nil
}

type TimeEntry struct {
	ID          int64   `json:"id"`
	WID         int64   `json:"wid"`
	PID         int64   `json:"pid"`
	Billable    bool    `json:"billable"`
	Start       string  `json:"start"`
	Stop        string  `json:"stop"`
	Duration    int64   `json:"duration"`
	Description string  `json:"description"`
	Tags        []int64 `json:"tags"`
	IsLocked    bool    `json:"is_locked"`
	Project     struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Billable  bool   `json:"billable"`
		IsPrivate bool   `json:"is_private"`
		Color     string `json:"color"`
	} `json:"project"`
	Task struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"task"`
	Workspace struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"workspace"`
}
