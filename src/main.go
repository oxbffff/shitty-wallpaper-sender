package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	token = ""
)

var (
	updatesCh = make(chan Updates)
	offset  int
)

func getPhotoURL() (string, error) {
	resp, err := http.Get(
		fmt.Sprintf(
			"https://wall.alphacoders.com/by_category.php?id=3&filter=HD&page=%d",
			rand.Intn(100),
		),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("<img(.*)src=\"(https://images.*)\" alt")
	links := re.FindAllStringSubmatch(string(body), -1)
	if links == nil {
		return "", errors.New("no links found")
	}

	re = regexp.MustCompile("thumb.*-")
	return re.ReplaceAllString(links[rand.Intn(len(links)-2)+1][2], ""), nil
}

func getPrettyJSON(body io.ReadCloser) (string, error) {
	parsedBody, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, parsedBody, "", "\t")
	if err != nil {
		return "", err
	}

	return string(prettyJSON.Bytes()), nil
}

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

func processingUpdates() {
	for {
		newUpdates := <-updatesCh

		for _, update := range newUpdates.Result {
			if checkIfCommand(update.Message.Entities) {
				if strings.Contains(update.Message.Text, "/start") {
					err := sendMessage(update.Message.Chat.ID, "Hello! I can send anime wallpaper")
					if err != nil {
						log.Println(err)
					}
				} else if strings.Contains(update.Message.Text, "/getpic") {
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

		updatesCh <- newUpdates
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	go getUpdates()
	processingUpdates()
}
