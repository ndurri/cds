package oauth

import (
	"io"
	"time"
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	"gosubmit/s3"
	"errors"
)

var TokenURL string
var ClientId string
var ClientSecret string
var TokenBucket string

type Token struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn int `json:"expires_in"`
	expires time.Time
	isNew bool
	id string
}

func GetToken(id string) (*Token, error) {
	token := Token{}
	created, err := s3.LoadToken(TokenBucket, "tokens/" + id, &token)
	if err != nil {
		return nil, err
	}
	token.expires = created.Add(time.Second * time.Duration(token.ExpiresIn))
	token.isNew = false
	token.id = id

	return &token, nil
}

func (t *Token) Expired() bool {
	return time.Now().After(t.expires)
}

func (t *Token) Refresh() (*Token, error) {
	if !t.Expired() {
		return t, nil
	}
	params := url.Values{
		"client_id":     {ClientId},
		"client_secret": {ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.RefreshToken},
	}
	res, err := http.PostForm(TokenURL, params)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("Token Refresh failed (%v)\n", res.StatusCode))
	}
	newToken := Token{}
	if err := json.Unmarshal([]byte(content), &newToken); err != nil {
		return nil, err
	}
	newToken.expires = time.Now().Add(time.Second * time.Duration(newToken.ExpiresIn))
	newToken.isNew = true
	newToken.id = t.id

	return &newToken, nil
}

func (t *Token) Save() error {
	if !t.isNew {
		return nil
	}
	content, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return s3.SaveJSON(TokenBucket, "tokens/" + t.id, string(content))
}