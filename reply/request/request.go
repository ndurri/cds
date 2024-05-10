package request

import (
	"encoding/json"
)

type Message struct {
	From string
	Subject string
	Bucket string
	Key string
	DocType string
	AccessToken string
	Id string
	Status int
	Body string
}

func FromJSON(content string) (*Message, error) {
	message := &Message{}
	if err := json.Unmarshal([]byte(content), message); err != nil {
		return nil, err
	}
	return message, nil
}
