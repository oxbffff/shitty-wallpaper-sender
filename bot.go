package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	startMessage = `
Hello! I can send random anime wallpaper
Use <code>/pic</code> to get random wallpaper
Or use <code>/sub</code> to get random wallpaper every 30 minutes
	`
	subMessage = `
You have subscribed to the newsletter of photos. Use <code>/unsub</code> to undo
	`
)

var (
	updatesCh   = make(chan Updates)
	subscribers = make(map[int](chan struct{}))

	token  string
	offset int
)

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

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env not found")
	}

	token = os.Getenv("TOKEN")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	f, err := os.OpenFile(os.Getenv("LOG_FILE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
}

func main() {
	go getUpdates()

	processingUpdates()
}
