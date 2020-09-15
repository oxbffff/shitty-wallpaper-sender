package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	startMessage = `
Hello! I can send random anime wallpapers
Use <code>/pic</code> to get random wallpaper
Or use <code>/sub</code> to get random wallpaper every 30 minutes
	`
	subMessage = `
You subscribed to send photos. Use <code>/unsub</code> to undo
	`
)

var (
	updatesCh   = make(chan Updates)
	token       = os.Getenv("TOKENWBOT")
	offset      int
	subscribers = make(map[int](chan struct{}))
)

func sendMessage(chatID int, text, parseMode string) error {
	body, err := doRequestToAPI(
		"sendMessage",
		url.Values{
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

func sendPhoto(chatID int, by func() (string, error)) error {
	data, err := by()
	if err != nil {
		return err
	}

	body, err := doRequestToAPI(
		"sendPhoto",
		url.Values{
			"chat_id":    {strconv.Itoa(chatID)},
			"photo":      {data},
			"caption":    {fmt.Sprintf("<a href=\"%s\">original</a>", data)},
			"parse_mode": {"HTML"},
		},
	)
	if err != nil {
		return err
	}
	defer body.Close()

	parsedBody, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	var tgResp map[string]interface{}

	if err = json.Unmarshal(parsedBody, &tgResp); err != nil {
		return err
	}

	if tgResp["ok"].(bool) != true {
		prettyJSON, err := getPrettyJSON(parsedBody)
		if err != nil {
			return err
		}

		log.Println(prettyJSON)

		err = sendMessage(chatID, "<code>"+prettyJSON+"</code>", "HTML")
		if err != nil {
			return err
		}
	}

	return nil
}

func getUpdates() {
	for {
		body, err := doRequestToAPI(
			"getUpdates",
			url.Values{
				"offset":  {strconv.Itoa(offset)},
				"timeout": {"30"},
			},
		)
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
						err := sendPhoto(update.Message.Chat.ID, getPhotoByURL)
						if err != nil {
							log.Println(err)
						}
					}()
				} else if strings.Contains(update.Message.Text, "/sub") {
					if _, ok := subscribers[update.Message.Chat.ID]; ok {
						continue
					}

					err := sendMessage(update.Message.Chat.ID, subMessage, "HTML")
					if err != nil {
						log.Println(err)
						continue
					}

					subscribers[update.Message.Chat.ID] = make(chan struct{})

					go func(done <-chan struct{}) {
						ticker := time.NewTicker(30 * time.Minute)

						for {
							select {
							case <-ticker.C:
								err = sendPhoto(update.Message.Chat.ID, getPhotoByURL)
								if err != nil {
									log.Println(err)
								}
							case <-done:
								ticker.Stop()
								return
							}
						}
					}(subscribers[update.Message.Chat.ID])
				} else if strings.Contains(update.Message.Text, "/unsub") {
					if _, ok := subscribers[update.Message.Chat.ID]; !ok {
						err := sendMessage(update.Message.Chat.ID, "Subscribe first", "")
						if err != nil {
							log.Println(err)
						}
						continue
					}

					close(subscribers[update.Message.Chat.ID])
					delete(subscribers, update.Message.Chat.ID)

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
	f, err := os.OpenFile("./info.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v\n", err)
	}
	defer f.Close()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(f)

	log.Println("TOKEN: " + token)

	go getUpdates()
	processingUpdates()
}
