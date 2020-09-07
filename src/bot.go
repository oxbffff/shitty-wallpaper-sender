package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	startMessage = `
Hello! I can send random anime wallpapers
Use <code>/pic</code> to get random wallpaper
Or use <code>/sub</code> to get random wallpaper every hour
	`
	subMessage = `
You subscribed to send photos. Use <code>/unsub</code> to undo
	`
)

var (
	updatesCh = make(chan Updates)
	token     = os.Getenv("TOKENWBOT")
	offset    int
)

func sendMessage(chatID int, text string, parseMode string) error {
	body, err := doRequestToAPI(
		"sendMessage",
		&url.Values{
			"chat_id":    {strconv.Itoa(chatID)},
			"text":       {text},
			"parse_mode": {parseMode},
		},
	)
	if err != nil {
		return err
	}
	body.Close()

	return nil
}

func sendPhoto(chatID int, photoURL string) error {
	body, err := doRequestToAPI(
		"sendPhoto",
		&url.Values{
			"chat_id": {strconv.Itoa(chatID)},
			"photo":   {photoURL},
		},
	)
	if err != nil {
		return err
	}
	defer body.Close()

	prettyJSON, err := getPrettyJSON(body)
	if err != nil {
		return err
	}

	err = sendMessage(chatID, "<code>" + prettyJSON + "</code>", "HTML")
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
					err := sendMessage(update.Message.Chat.ID, startMessage, "HTML")
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
					err := sendMessage(update.Message.Chat.ID, subMessage, "HTML")
					if err != nil {
						log.Println(err)
					}
				} else if strings.Contains(update.Message.Text, "/unsub") {
					err := sendMessage(update.Message.Chat.ID, "Done", "")
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
