package telegram

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"slices"
	"strings"

	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const ParseModeMarkdownV2 = "MarkdownV2"

type SendError struct {
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

func (e SendError) Error() string {
	return fmt.Sprintf("telegram send error: code=%d description=%s", e.ErrorCode, e.Description)
}

type Telegram struct {
	apiToken string
}

func NewTelegram(apiToken string) Telegram {
	return Telegram{apiToken: apiToken}
}

func (t *Telegram) Send(chatID int, message string, parseMode string) error {
	sendUrl := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   fmt.Sprintf("bot%s/sendMessage", t.apiToken),
	}

	message = addBacklash(message)

	requestData := struct {
		ChatID    int    `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}{
		ChatID:    chatID,
		Text:      message,
		ParseMode: parseMode,
	}

	b, err := json.Marshal(requestData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal telegram message")
	}

	requestBody := bytes.NewReader(b)

	// fixme: use post request
	resp, err := http.Post(sendUrl.String(), "application/json", requestBody)
	if err != nil {
		return err
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var body struct {
		Ok          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code"`
		Description string `json:"description"`
	}
	err = json.Unmarshal(b, &body)
	if err != nil {
		return err
	}
	if body.ErrorCode != 0 {
		return SendError{
			ErrorCode:   body.ErrorCode,
			Description: body.Description,
		}
	}

	log.Printf("sent telegram message: %s", string(b))
	return nil
}

// addBacklash escapes special characters in the given string.
func addBacklash(s string) string {
	b := strings.Builder{}
	escapeChars := []rune{'_', '.'}
	for _, c := range s {
		if slices.Contains(escapeChars, c) {
			b.WriteRune('\\')
			b.WriteRune(c)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func (t *Telegram) Ask(chatID string, query string) error {
	// todo: send a telegram message expecting a simple text reply
	return errors.New("not implemented")
}
