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

const userAgent = "Mozilla/5.0 (X11; Linux i686; rv:86.0) Gecko/20100101 Firefox/86.0"

func getPhotoByURL() (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"https://wall.alphacoders.com/by_category.php?id=3&filter=HD&page=%d",
			rand.Intn(100),
		),
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
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

func getPrettyJSON(parsedBody []byte) (string, error) {
	var prettyJSON bytes.Buffer

	err := json.Indent(&prettyJSON, parsedBody, "", "\t")
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

func doRequestToAPI(method string, values url.Values) (io.ReadCloser, error) {
	resp, err := http.PostForm(
		fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method),
		values,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
