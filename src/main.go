package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	token = "1381993359:AAFFtO_qc_Pz0pMgqHbShTYYVIlUT12xlVc"
)

var (
	updates = make(chan Updates)
	offset  int
)

func checkIfCommand(entities []MessageEntity) bool {
	for _, entity := range entities {
		if entity.Type == "bot_command" {
			return true
		}
	}
	return false
}

func doRequestToAPI(method string, values *url.Values) (io.ReadCloser, error) {
	resp, err := http.PostForm(
		fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method),
		*values,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func sendMessage(chatID int, text string) error {
	body, err := doRequestToAPI("sendMessage", &url.Values{"chat_id": {strconv.Itoa(chatID)}, "text": {text}})
	if err != nil {
		return err
	}
	defer body.Close()

	return nil
}

func processingUpdates() {
	for {
		newUpdates := <-updates
		for _, update := range newUpdates.Result {
			if checkIfCommand(update.Message.Entities) {
				if strings.Contains(update.Message.Text, "/start") {
					err := sendMessage(update.Message.Chat.ID, "Hello! I can send anime wallpaper")
					if err != nil {
						log.Println(err)
					}
				} else if strings.Contains(update.Message.Text, "/getpic") {
					err := sendMessage(update.Message.Chat.ID, "Sorry:( not implemented")
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}

func getUpdates() {
	for {
		body, err := doRequestToAPI("getUpdates", &url.Values{"offset": {strconv.Itoa(offset)}, "timeout": {"30"}})
		if err != nil {
			log.Println(err)
			continue
		}
		defer body.Close()

		var newUpdates Updates
		err = json.NewDecoder(body).Decode(&newUpdates)
		if err != nil {
			log.Println(err)
			continue
		}

		updatesCount := len(newUpdates.Result)
		if updatesCount > 0 {
			offset = newUpdates.Result[updatesCount-1].UpdateID + 1
		}

		updates <- newUpdates
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	go getUpdates()
	processingUpdates()
}
