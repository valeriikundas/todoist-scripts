package toggl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"encore.dev/storage/cache"
)

const telegramBotAPI = "https://api.telegram.org/bot"

// fixme: generalize this function
func AskForTogglEntryInTelegram(telegramApiToken string, telegramUserID int, TelegramGetUpdatesOffset *cache.IntKeyspace[int]) (string, error) {
	updateURL := telegramBotAPI + telegramApiToken + "/getUpdates"

	query := "no running Toggl entry. please fill in:"
	sendMessage(telegramApiToken, telegramUserID, query)

	offsetValue, err := TelegramGetUpdatesOffset.Get(context.TODO(), 0)
	if err != nil {
		return "", err
	}

	// fixme: rewrite with telegram webhook
	updates, err := getUpdates(updateURL, int(offsetValue))
	if err != nil {
		return "", err
	}

	_, err = TelegramGetUpdatesOffset.Increment(context.TODO(), 0, 1)
	if err != nil {
		return "", err
	}

	for i, update := range updates {
		log.Printf("%d %#v", i, update)

		replyText := fmt.Sprintf("Ok. Recorded: '%s' ", update.Message.Text)
		sendMessage(telegramApiToken, update.Message.Chat.ID, replyText)

		TelegramGetUpdatesOffset.Set(context.TODO(), 0, int64(update.ID+1))

		return update.Message.Text, nil
	}

	return "", errors.New("no running Toggl entry and no Telegram reply")
}

func getUpdates(url string, offset int) ([]Update, error) {
	fullUrl := url + "?offset=" + strconv.Itoa(offset) + "&allowed_updates=message&timeout=120"
	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response GetUpdatesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func sendMessage(token string, chatID int, text string) {
	apiURL := telegramBotAPI + token + "/sendMessage"
	message := SendMessage{
		ChatID: chatID,
		Text:   text,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(messageBytes))
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}
	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)
}

type GetUpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	ID      int     `json:"update_id"`
	Message Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ID int `json:"id"`
}

type SendMessage struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}
