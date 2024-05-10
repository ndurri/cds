package response

import (
	"encoding/json"
)

type Message struct {
	RequestId string
	Bucket string
	Key string
}

func FromJSON(content string) (*Message, error) {
	message := &Message{}
	if err := json.Unmarshal([]byte(content), message); err != nil {
		return nil, err
	}
	return message, nil
}

