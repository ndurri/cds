package request

import (
	"encoding/json"
)

type Request struct {
	UUID string
	From string
	Subject string
	Bucket string
	Key string
	DocType string
	AccessToken string
}

func FromJSON(content string) (*Request, error) {
	var req Request
	if err := json.Unmarshal([]byte(content), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (r Request) ToJSON() (*string, error) {
	content, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	retstr := string(content)
	return &retstr, nil
}