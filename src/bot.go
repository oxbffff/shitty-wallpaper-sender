package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	updatesCh = make(chan Updates)
	token     = os.Getenv("TOKENWBOT")
	offset    int
)

func sendMessage(chatID int, text string) error {
	body, err := doRequestToAPI("sendMessage", &url.Values{"chat_id": {strconv.Itoa(chatID)}, "text": {text}})
	if err != nil {
		return err
	}
	body.Close()

	return nil
}

func sendPhoto(chatID int, photoURL string) error {
	body, err := doRequestToAPI("sendPhoto", &url.Values{"chat_id": {strconv.Itoa(chatID)}, "photo": {photoURL}})
	if err != nil {
		return err
	}
	defer body.Close()

	prettyJSON, err := getPrettyJSON(body)
	if err != nil {
		return err
	}

	err = sendMessage(chatID, prettyJSON)
	if err != nil {
		return err
	}

	return nil
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

		updatesCh <- newUpdates
	}
}

func processingUpdates() {
	for newUpdates := range updatesCh {
		for _, update := range newUpdates.Result {
			if checkIfCommand(update.Message.Entities) {
				if strings.Contains(update.Message.Text, "/start") {
					err := sendMessage(update.Message.Chat.ID, "Hello! I can send anime wallpaper")
					if err != nil {
						log.Println(err)
					}
				} else if strings.Contains(update.Message.Text, "/pic") {
					go func() {
						photoURL, err := getPhotoURL()
						if err != nil {
							log.Println(err)
						}
						err = sendPhoto(update.Message.Chat.ID, photoURL)
						if err != nil {
							log.Println(err)
						}
					}()
				} else if strings.Contains(update.Message.Text, "/sub") {
					err := sendMessage(update.Message.Chat.ID, "You subscribed to send photos. Use /unsub to undo")
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("TOKEN: " + token)

	go getUpdates()
	processingUpdates()
}
