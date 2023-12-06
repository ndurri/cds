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
)

var movSubmitter string = os.Getenv("MOV_SUBMITTER")
var decSubmitter string = os.Getenv("DEC_SUBMITTER")
var tokenBucket string = os.Getenv("TOKEN_BUCKET")
var tokenPrefix string = os.Getenv("TOKEN_PREFIX")
var notifyTopic string = os.Getenv("NOTIFY_TOPIC")

var tokenCache = map[string]*oauth.Token{}

func getSubmitter(docType string) string {
	if docType == "Movement" {
		return movSubmitter
	} else {
		return decSubmitter
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
	content, err := s3.Get(tokenBucket, tokenPrefix + id)
	if err != nil {
		return nil, err
	}
	return oauth.TokenFromJSON(string(content))
}

func saveToken(id string, token *oauth.Token) error {
	if err := s3.PutAsJSON(tokenBucket, tokenPrefix + id, token); err != nil {
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
	if err := sns.PublishJSON(notifyTopic, request); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Finished.")
}

func main() {
	lambda.Start(lambdaMain)
}