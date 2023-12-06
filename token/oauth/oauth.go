package oauth

import (
	"io"
	"time"
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	"errors"
	"os"
)

var TokenURL string = os.Getenv("OAUTH_TOKEN_URL")
var ClientId string = os.Getenv("OAUTH_CLIENT_ID")
var ClientSecret string = os.Getenv("OAUTH_CLIENT_SECRET")

type Token struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn int `json:"expires_in"`
	Created time.Time `json:"created"`
}

func NewToken() *Token {
	return &Token{}
}

func TokenFromJSON(content string) (*Token, error) {
	token := NewToken()
	if err := json.Unmarshal([]byte(content), token); err != nil {
		return nil, err
	}

	return token, nil
}

func (t *Token) Expired() bool {
	expires := t.Created.Add(time.Second * time.Duration(t.ExpiresIn))
	return time.Now().After(expires)
}

func (t *Token) Refresh() (*Token, error) {
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
	newToken, err := TokenFromJSON(string(content))
	if err != nil {
		return nil, err
	}
	newToken.Created = time.Now()

	return newToken, nil
}
