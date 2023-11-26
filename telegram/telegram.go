package telegram

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Telegram struct {
	apiToken string
}

func NewTelegram(apiToken string) Telegram {
	return Telegram{apiToken: apiToken}
}

func (t *Telegram) Send(chatID string, message string) error {
	urlObj := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   fmt.Sprintf("bot%s/sendMessage", t.apiToken),
		RawQuery: url.Values{
			"chat_id": []string{chatID},
			"text":    []string{message},
		}.Encode(),
	}
	url := urlObj.String()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("sent telegram message: %s", string(b))
	return nil
}
