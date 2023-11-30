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
	"time"

	"encore.dev/storage/cache"
)

const telegramBotAPI = "https://api.telegram.org/bot"

// fixme: generalize this function
func AskForTogglEntryInTelegram(telegramApiToken string, telegramUserID int, TelegramUpdatesOffsetKeyspace *cache.IntKeyspace[int]) (string, error) {
	updateURL := telegramBotAPI + telegramApiToken + "/getUpdates"

	queryTime := time.Now().Unix()

	query := "no running Toggl entry. please fill in:"
	sendMessage(telegramApiToken, telegramUserID, query)

	offsetValue, err := TelegramUpdatesOffsetKeyspace.Get(context.TODO(), 0)
	if err != nil {
		if errors.Is(err, cache.Miss) {
			offsetValue = 0
		} else {
			return "", err
		}
	}

	// todo: #9 rewrite with telegram webhook

	replyChan := make(chan ReplyResult, 1)
	go telegramWaitForReply(updateURL, offsetValue, TelegramUpdatesOffsetKeyspace, queryTime, telegramApiToken, replyChan)

	select {
	case res := <-replyChan:
		return res.message, res.err
	case <-time.After(time.Minute):
		return "", &TelegramTimeoutError{}
	}
}

func telegramWaitForReply(updateURL string, offsetValue int64, TelegramUpdatesOffsetKeyspace *cache.IntKeyspace[int], queryTime int64, telegramApiToken string, result chan ReplyResult) {
	for receivedEntry := false; !receivedEntry; {
		updates, err := getUpdates(updateURL, int(offsetValue))
		if err != nil {
			result <- ReplyResult{"", err}
		}

		_, err = TelegramUpdatesOffsetKeyspace.Increment(context.TODO(), 0, 1)
		if err != nil {
			result <- ReplyResult{"", err}
		}

		for _, update := range updates {
			if update.Message.Date < int(queryTime) {
				log.Printf("message `%s` was sent before interested. skipping...", update.Message.Text)
				continue
			}

			replyText := fmt.Sprintf("Ok. Recorded: '%s' ", update.Message.Text)
			sendMessage(telegramApiToken, update.Message.Chat.ID, replyText)

			TelegramUpdatesOffsetKeyspace.Set(context.TODO(), 0, int64(update.ID+1))

			receivedEntry = true
			result <- ReplyResult{update.Message.Text, nil}
		}

		time.Sleep(time.Second)
	}
}

type ReplyResult struct {
	message string
	err     error
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
	Date int    `json:"date"`
}

type Chat struct {
	ID int `json:"id"`
}

type SendMessage struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}
