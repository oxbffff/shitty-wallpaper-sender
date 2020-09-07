package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"time"
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

	re := regexp.MustCompile("<img(.*)src=\"(https://images.*)\"\\s+(.*)>")
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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
