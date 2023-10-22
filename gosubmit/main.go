package main

import (
	"os"
	"fmt"
	"gosubmit/s3"
	"gosubmit/oauth"
	"gosubmit/api"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
)

type Request struct {
	UUID string
	From string
	Subject string
	Command string
	DocType string
	ResponseUUID string
	ResponseStatus int
	ResponseBody string
}

var movSubmitter string
var decSubmitter string
var appBucket string

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

func lambdaMain(event events.SQSEvent) {
	request := Request{}
	if err := json.Unmarshal([]byte(event.Records[0].Body), &request); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Processing request %s\n", request.UUID)
	submitter := getSubmitter(request.DocType)
	token, err := getValidToken(submitter)
	checkError(err)
	api, err := api.GetAPI(request.DocType)
	checkError(err)
	payload, err := s3.Load(appBucket, "payloads/" + request.UUID)
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
	err = s3.SaveJSON(appBucket, "requests/" + request.UUID, request)
	checkError(err)
	if resp.Ok() {
		err = s3.SaveJSON(appBucket, "responses/" + request.ResponseUUID, request)
	} else {
		err = s3.SaveJSON(appBucket, "failed/" + request.UUID, request)
	}
	checkError(err)
	fmt.Println("Finished.")
}

func main() {
	movSubmitter = os.Getenv("MOV_SUBMITTER")
	decSubmitter = os.Getenv("DEC_SUBMITTER")
	appBucket = os.Getenv("APPDATA_BUCKET")
	oauth.ClientId = os.Getenv("CLIENT_ID")
	oauth.ClientSecret = os.Getenv("CLIENT_SECRET")
	oauth.TokenBucket = os.Getenv("TOKEN_BUCKET")
	api.APIHost = os.Getenv("API_HOST")
	oauth.TokenHost = api.APIHost
	lambda.Start(lambdaMain)
}