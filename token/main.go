package main

import (
	"os"
	"fmt"
	"token/aws"
	"token/oauth"
	"token/request"
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
	content, err := aws.S3.Get(bucket, prefix + id)
	if err != nil {
		return nil, err
	}
	return oauth.TokenFromJSON(string(content))
}

func saveToken(id string, token *oauth.Token) error {
	bucket, prefix := splitPrefix(TokenBucket)
	content, err := token.ToJSON()
	if err != nil {
		return err
	}
	return aws.S3.Put(bucket, prefix + id, []byte(*content))
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

func processMessage(message string) {
	fmt.Println("token: Received notification.")
	request, err := request.FromJSON(message)
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
	content, err := request.ToJSON()
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := aws.SNS.Put(NotifyTopic, *content); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("token: Finished.")
}

func main() {
	if err := aws.Config(); err != nil {
		panic(err)
	}
	aws.Lambda.StartSNS(processMessage)
}