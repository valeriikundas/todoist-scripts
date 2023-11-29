package telegram

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Telegram struct {
	apiToken string
}

func NewTelegram(apiToken string) Telegram {
	return Telegram{apiToken: apiToken}
}

func (t *Telegram) Send(chatID int, message string) error {
	urlObj := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   fmt.Sprintf("bot%s/sendMessage", t.apiToken),
		RawQuery: url.Values{
			"chat_id": []string{strconv.Itoa(chatID)},
			"text":    []string{message},
		}.Encode(),
	}
	url := urlObj.String()

	// fixme: use post request
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

func (t *Telegram) Ask(chatID string, query string) error {
	// todo: send a telegram message expecting a simple text reply
	return errors.New("not implemented")
}
