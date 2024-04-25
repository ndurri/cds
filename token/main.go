package main

import (
	"os"
	"fmt"
	"token/s3"
	"token/oauth"
	"token/request"
	"token/sns"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"strings"
)

var MovSubmitter string = os.Getenv("MOV_SUBMITTER")
var DecSubmitter string = os.Getenv("DEC_SUBMITTER")
var TokenBucket string = os.Getenv("TOKEN_BUCKET")
var NotifyTopic string = os.Getenv("NEXT_TOPIC")

var tokenCache = map[string]*oauth.Token{}

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
}

func getSubmitter(docType string) string {
	if docType == "Movement" {
		return MovSubmitter
	} else {
		return DecSubmitter
	}
}

func getTokenFromCache(id string) *oauth.Token {
	token, prs := tokenCache[id]
	if !prs {
		return nil
	} else {
		return token
	}
}

func loadToken(id string) (*oauth.Token, error) {
	token := getTokenFromCache(id)
	if token != nil {
		fmt.Println("Token found in cache.")
		return token, nil
	}
	fmt.Println("Token not in cache. Loading.")
	bucket, prefix := splitPrefix(TokenBucket)
	content, err := s3.Get(bucket, prefix + id)
	if err != nil {
		return nil, err
	}
	return oauth.TokenFromJSON(content)
}

func saveToken(id string, token *oauth.Token) error {
	bucket, prefix := splitPrefix(TokenBucket)
	if err := s3.PutAsJSON(bucket, prefix + id, token); err != nil {
		return err
	}
	return nil
}

func getValidToken(submitter string) (*oauth.Token, error) {
	token, err := loadToken(submitter)
	if err != nil {
		return nil, err
	}
	if !token.Expired() {
		fmt.Println("Non-expired token found.")
		return token, nil
	}
	fmt.Println("Token is expired. Refreshing.")
	token, err = token.Refresh()
	if err != nil {
		return nil, err
	}
	fmt.Println("Saving new token.")
	tokenCache[submitter] = token
	if err := saveToken(submitter, token); err != nil {
		return nil, err		
	}
	return token, nil
}

func lambdaMain(event events.SNSEvent) {
	fmt.Printf("Received notification on topic %s\n", event.Records[0].SNS.TopicArn)
	request, err := request.FromJSON(event.Records[0].SNS.Message)
	if err != nil {
		fmt.Println(err)
		return
	}
	if request.DocType == "" {
		fmt.Println("ERROR: Request does not contain a doctype.")
		return		
	}
	fmt.Printf("Processing token request for %s:%s\n", request.Bucket, request.Key)
	submitter := getSubmitter(request.DocType)
	fmt.Printf("Document type is %s. Using submitter %s\n", request.DocType, submitter)
	fmt.Printf("Requesting token for submitter %s\n", submitter)
	token, err := getValidToken(submitter)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Token found. Adding to request.")
	request.AccessToken = token.AccessToken
	if err := sns.PublishJSON(NotifyTopic, request); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Finished.")
}

func main() {
	lambda.Start(lambdaMain)
}