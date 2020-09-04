package main

type Updates struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	Message  `json:"message"`
	UpdateID int `json:"update_id"`
}

type Message struct {
	Chat      `json:"chat"`
	MessageID int             `json:"message_id"`
	Text      string          `json:"text"`
	Entities  []MessageEntity `json:"entities"`
}

type Chat struct {
	ID int `json:"id"`
}

type MessageEntity struct {
	Type string `json:"type"`
}
