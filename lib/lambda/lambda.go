package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"encoding/base64"
	"fmt"
	"errors"
)

var InvalidEvent = errors.New("Invalid Event")

func validateS3Event(event events.S3Event) (*string, *string, error) {
	if len(event.Records) != 1 {
		return nil, nil, InvalidEvent
	}
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.URLDecodedKey
	if bucket == "" || key == "" {
		return nil, nil, InvalidEvent
	}
	return &bucket, &key, nil
}


func validateSNSEvent(event events.SNSEvent) (*string, error) {
	if len(event.Records) != 1 {
		return nil, InvalidEvent
	}
	message := event.Records[0].SNS.Message
	if message == "" {
		return nil, InvalidEvent
	}
	return &message, nil
}

func StartS3(fn func(string, string)) {
	lambda.Start(func(event events.S3Event) {
		var bucket, key *string
		var err error
		if bucket, key, err = validateS3Event(event); err != nil {
			fmt.Println("ERROR: S3 Notification is not valid.")
			fmt.Println(event)
			return
		}
		fn(*bucket, *key)
	})
}

func StartSNS(fn func(string)) {
	lambda.Start(func(event events.SNSEvent) {
		var message *string
		var err error
		if message, err = validateSNSEvent(event); err != nil {
			fmt.Println("ERROR: SNS Notification is not valid.")
			fmt.Println(event)
			return
		}
		fn(*message)
	})
}

func decodeBody(event *events.APIGatewayProxyRequest) (*string, error) {
	if !event.IsBase64Encoded {
		return &event.Body, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(event.Body)
	if err != nil {
		return nil, err
	}
	retstr := string(decoded)
	return &retstr, nil
}

func StartAPI(fn func(map[string]string, string)int) {
	lambda.Start(func(event *events.APIGatewayProxyRequest)(*events.APIGatewayProxyResponse, error) {
		body, err := decodeBody(event)
		if err != nil {
			fmt.Println(err)
			return &events.APIGatewayProxyResponse{StatusCode: 500}, nil
		}
		status := fn(event.Headers, *body)
		return &events.APIGatewayProxyResponse{StatusCode: status}, nil
	})
}

func Start(fn func()) {
	lambda.Start(fn)
}