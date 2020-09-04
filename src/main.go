package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	token = ""
)

var (
	updates = make(chan Updates)
	offset  int
)

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
		log.Println(err.Error())
		return err
	}
	defer body.Close()

	return nil
}

func processingUpdates() {
	for {
		select {
		case newUpdates := <-updates:
			for _, update := range newUpdates.Result {
				err := sendMessage(update.Message.Chat.ID, update.Message.Text)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}
}

func getUpdates() {
	for {
		body, err := doRequestToAPI("getUpdates", &url.Values{"offset": {strconv.Itoa(offset)}, "timeout": {"30"}})
		if err != nil {
			log.Println(err.Error())
			continue
		}
		defer body.Close()

		var newUpdates Updates
		err = json.NewDecoder(body).Decode(&newUpdates)
		if err != nil {
			log.Println(err.Error())
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
