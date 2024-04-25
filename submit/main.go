package main

import (
	"os"
	"fmt"
	"submit/request"
	"submit/s3"
	"submit/sns"
	"submit/api"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
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
		return errors.New("Request Bucket or Key not provided.")
	} else if request.DocType == "" {
		return errors.New("Request DocType not provided.")		
	} else if request.AccessToken == "" {
		return errors.New("Request AccessKey not provided.")		
	}
	return nil
}

func lambdaMain(event events.SNSEvent) {
	fmt.Printf("Received notification on topic %s\n", event.Records[0].SNS.TopicArn)
	request, err := request.FromJSON(event.Records[0].SNS.Message)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := validateRequest(request); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Printf("Processing submit request for %s:%s\n", request.Bucket, request.Key)
	api, err := api.GetAPI(request.DocType)
	if err != nil {
		fmt.Println(err)
		return
	}
	payload, err := s3.Get(request.Bucket, request.Key)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Calling %v...\n", api.Endpoint)
	resp, err := api.Call(request.AccessToken, payload)
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
	bucket, prefix := splitPrefix(RequestBucket)
	if err := s3.PutJSON(bucket, prefix + request.Id, request); err != nil {
		fmt.Println(err)
		return		
	}
	if err := sns.PublishJSON(NotifyTopic, request); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Println("Finished.")
}

func main() {
	api.APIHost = APIHost
	lambda.Start(lambdaMain)
}