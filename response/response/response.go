package response

import (
	"encoding/json"
)

type Message struct {
	RequestId string
	Bucket string
	Key string
}

func (m Message) ToJSON() (*string, error) {
	content, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	retstr := string(content)
	return &retstr, nil
}

