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
	"strings"
)

var (
	TokenHost = os.Getenv("TOKEN_HOST")
	ClientId = os.Getenv("CLIENT_ID")
	ClientSecret = os.Getenv("CLIENT_SECRET")
	RedirectURI = os.Getenv("REDIRECT_URI")
	TokenBucket = os.Getenv("TOKEN_BUCKET")
	SessionBucket = os.Getenv("SESSION_BUCKET")
)

var UUIDRegex = regexp.MustCompile("^[0-9a-f]{8}\\b-[0-9a-f]{4}\\b-[0-9a-f]{4}\\b-[0-9a-f]{4}\\b-[0-9a-f]{12}$")

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
}

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
	code, prs := event.QueryStringParameters["code"]
	if !prs || code == "" {
		fmt.Println("Code not provided in auth-redirect.")
		fmt.Println(event)
		return
	}
	state, prs := event.QueryStringParameters["state"]
	if !prs || state == "" || !UUIDRegex.MatchString(state) {
		fmt.Println("State not provided or malformed in auth-redirect.")
		fmt.Println(event)
		return
	}
	bucket, prefix := splitPrefix(SessionBucket)
	fmt.Printf("Retrieving session from %s:%s.", bucket, prefix + state)
	session, err := s3.Get(bucket, prefix + state)
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
	bucket, prefix = splitPrefix(TokenBucket)
	fmt.Printf("Saving token to %s:%s.", bucket, prefix + session)
	if err := s3.Put(bucket, prefix + session, token); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Println("Finished.")
}

func main() {
	lambda.Start(lambdaMain)
}