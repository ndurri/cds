package request

import (
	"encoding/json"
)

type Request struct {
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

func FromJSON(content string) (*Request, error) {
	var req Request
	if err := json.Unmarshal([]byte(content), &req); err != nil {
		return nil, err
	}
	return &req, nil
}