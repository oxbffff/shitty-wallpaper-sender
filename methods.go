package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
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
