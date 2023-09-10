package main

import (
	"os"
	"fmt"
	"strings"
	"gosubmit/s3"
	"gosubmit/oauth"
	"gosubmit/api"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
)

type Request struct {
	UUID string
	From string
	Subject string
	Command string
	Submitter string
	DocType string
	ResponseUUID string
	ResponseStatus int
	ResponseBody string
}

var movSubmitter string
var decSubmitter string

var tokenCache = map[string]*oauth.Token{}

func getSubmitter(docType string) string {
	if docType == "Movement" {
		return movSubmitter
	} else {
		return decSubmitter
	}
}

func getValidToken(submitter string) (*oauth.Token, error) {
	cacheToken, prs := tokenCache[submitter]
	if prs && !cacheToken.Expired() {
		return cacheToken, nil
	}
	fmt.Printf("Requesting token for submitter %s\n", submitter)
	var newToken *oauth.Token
	var err error
	if !prs {
		newToken, err = oauth.GetToken(submitter)
		if err != nil {
			return nil, err
		}
	}
	newToken, err = newToken.Refresh()
	if err != nil {
		return nil, err
	}
	tokenCache[submitter] = newToken
	if err := newToken.Save(); err != nil {
		fmt.Printf("WARNING: Unable to save new token. %v\n", err)
	}
	fmt.Println("Success.")
	return newToken, nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func lambdaMain(event events.S3Event) {
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.Key
	uuid := strings.Split(key, "/")[1]
	fmt.Printf("Processing request %s\n", uuid)
	request := Request{}
	if err := s3.LoadJSON(bucket, "requests/" + uuid, &request); err != nil {
		fmt.Println(err)
		return
	}
	submitter := getSubmitter(request.DocType)
	token, err := getValidToken(submitter)
	checkError(err)
	api, err := api.GetAPI(request.DocType)
	checkError(err)
	payload, err := s3.Load(bucket, "payloads/" + uuid)
	checkError(err)
	resp, err := api.Call(token.AccessToken, payload)
	checkError(err)
	fmt.Printf("API returned %v.\n", resp.StatusCode)
	if !resp.Ok() {
		fmt.Println(resp.Body)
	}
	request.ResponseUUID = resp.ConversationId
	request.ResponseStatus = resp.StatusCode
	request.ResponseBody = resp.Body
	err = s3.SaveJSON(bucket, "requests/" + uuid, request)
	checkError(err)
	if resp.Ok() {
		err = s3.SaveJSON(bucket, "responses/" + request.ResponseUUID, request)
	} else {
		err = s3.SaveJSON(bucket, "failed/" + uuid, request)
	}
	checkError(err)
	fmt.Println("Finished.")
}

func main() {
	movSubmitter = os.Getenv("MOV_SUBMITTER")
	decSubmitter = os.Getenv("DEC_SUBMITTER")
	oauth.TokenURL = os.Getenv("TOKEN_URL")
	oauth.ClientId = os.Getenv("CLIENT_ID")
	oauth.ClientSecret = os.Getenv("CLIENT_SECRET")
	oauth.TokenBucket = os.Getenv("TOKEN_BUCKET")
	lambda.Start(lambdaMain)
}