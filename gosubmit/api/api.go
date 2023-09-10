package api

import (
	"log"
	"fmt"
	"io"
	"encoding/json"
	"net/http"
	"strings"
	_ "embed"
	"errors"
)

//go:embed apis.json
var apiJSON []byte

type API struct {
	Endpoint string `json:"endpoint"`
	Headers map[string]string `json:"headers"`
}

type APIResponse struct {
	StatusCode int
	ConversationId string
	Body string
}

var apis = map[string]*API{}

func init() {
	if err := json.Unmarshal(apiJSON, &apis); err != nil {
		log.Fatal(err)
	}
}

func GetAPI(doctype string) (*API, error) {
	api, prs := apis[doctype]
	if !prs {
		return nil, errors.New("API does not exist for doc type " + doctype)
	}
	return api, nil
}

func (api *API) Call(token string, body string) (*APIResponse, error) {
	reader := strings.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, api.Endpoint, reader)
	if err != nil {
		return nil, err
	}
	for name, value := range api.Headers {
		req.Header.Add(name, value)
	}
	req.Header.Add("authorization", "Bearer " + token)
	fmt.Printf("Calling %s with access token %s.\n", api.Endpoint, token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	conversationId := res.Header.Get("x-conversation-id")
	return &APIResponse{StatusCode: res.StatusCode, ConversationId: conversationId, Body: string(content),}, nil
}

func (r *APIResponse) Ok() bool {
	return r.StatusCode >= 200 && r.StatusCode <= 299
}