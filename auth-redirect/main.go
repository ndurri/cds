package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"io"
	"errors"
	"auth-redirect/s3"
	"regexp"
)

var TokenHost = os.Getenv("TOKEN_HOST")
var ClientId = os.Getenv("CLIENT_ID")
var ClientSecret = os.Getenv("CLIENT_SECRET")
var RedirectURI = os.Getenv("REDIRECT_URI")
var TokenBucket = os.Getenv("TOKEN_BUCKET")
var TokenPrefix = os.Getenv("TOKEN_PREFIX")
var SessionBucket = os.Getenv("SESSION_BUCKET")
var SessionPrefix = os.Getenv("SESSION_PREFIX")

var UUIDRegex = regexp.MustCompile("^[0-9a-f]{8}\\b-[0-9a-f]{4}\\b-[0-9a-f]{4}\\b-[0-9a-f]{4}\\b-[0-9a-f]{12}$")

func getToken(code string) (string, error) {
	params := url.Values{
		"client_id":     {ClientId},
		"client_secret": {ClientSecret},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {RedirectURI},
		"code":          {code},
	}
	resp, err := http.PostForm(TokenHost, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err		
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errm := fmt.Sprintf("Call to %s failed with status %d.\nServer returned %s.", TokenHost, resp.StatusCode, content)
		return "", errors.New(errm)
	}
	return string(content), nil
}

func lambdaMain(event *events.APIGatewayProxyRequest) {
	code := event.QueryStringParameters["code"]
	state := event.QueryStringParameters["state"]
	if code == "" {
		fmt.Println("Code not provided in auth-redirect.")
		fmt.Println(event)
		return
	}
	if state == "" || !UUIDRegex.MatchString(state) {
		fmt.Println("State not provided or malformed in auth-redirect.")
		fmt.Println(event)
		return
	}
	fmt.Printf("Retrieving session from %s:%s.", SessionBucket, SessionPrefix + state)
	session, err := s3.Get(SessionBucket, SessionPrefix + state)
	if err != nil {
		fmt.Println(err)
		return
	}	
	fmt.Printf("Exchanging code %s for token.", code)
	token, err := getToken(code)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Saving token to %s:%s.", TokenBucket, TokenPrefix + session)
	if err := s3.Put(TokenBucket, TokenPrefix + session, token); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Println("Finished.")
}

func main() {
	lambda.Start(lambdaMain)
}