package main

import (
	"os"
	"fmt"
	"submit/aws"
	"submit/request"
	"submit/api"
	"errors"
	"strings"
)

var APIHost = os.Getenv("API_HOST")
var RequestBucket = os.Getenv("REQUEST_BUCKET")
var NotifyTopic = os.Getenv("NEXT_TOPIC")

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
}

func validateRequest(request *request.Request) error {
	if request.Bucket == "" || request.Key == "" {
		return errors.New("ERROR: Request Bucket or Key not provided.")
	} else if request.DocType == "" {
		return errors.New("ERROR: Request DocType not provided.")		
	} else if request.AccessToken == "" {
		return errors.New("ERROR: Request AccessKey not provided.")		
	}
	return nil
}

func processMessage(message string) {
	fmt.Println("submit: Received notification")
	request, err := request.FromJSON(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := validateRequest(request); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Printf("submit: Processing request for %s:%s\n", request.Bucket, request.Key)
	api, err := api.GetAPI(request.DocType)
	if err != nil {
		fmt.Println(err)
		return
	}
	payload, err := aws.S3.Get(request.Bucket, request.Key)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Calling %v...\n", api.Endpoint)
	resp, err := api.Call(request.AccessToken, string(payload))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("API returned %v.\n", resp.StatusCode)
	if !resp.Ok() {
		fmt.Println(resp.Body)
	}
	request.Id = resp.ConversationId
	request.Status = resp.StatusCode
	request.Body = resp.Body
	content, err := request.ToJSON()
	if err != nil {
		fmt.Println(err)
		return		
	}
	bucket, prefix := splitPrefix(RequestBucket)
	if err := aws.S3.Put(bucket, prefix + request.Id, []byte(*content)); err != nil {
		fmt.Println(err)
		return		
	}
	if err := aws.SNS.Put(NotifyTopic, *content); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Println("submit: Finished.")
}

func main() {
	api.APIHost = APIHost
	if err := aws.Config(); err != nil {
		panic(err)
	}
	aws.Lambda.StartSNS(processMessage)
}