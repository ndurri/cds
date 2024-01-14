package main

import (
	"os"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"request/request"
	"request/s3"
	"request/sns"
	"encoding/base64"
)

var payloadBucket = os.Getenv("PAYLOAD_BUCKET")
var payloadPrefix = os.Getenv("PAYLOAD_PREFIX")
var notifyTopic = os.Getenv("NOTIFY_TOPIC")

var Response201 = &events.APIGatewayProxyResponse{StatusCode: 201,}
var Response400 = &events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request",}
var Response500 = &events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error",}

func decodeBody(event *events.APIGatewayProxyRequest) (string, error) {
	if !event.IsBase64Encoded {
		return event.Body, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(event.Body)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func handleLambda(event *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Printf("Received request for %s from %s.\n", event.Path, event.RequestContext.Identity.SourceIP)
	body, err := decodeBody(event)
	if err != nil {
		fmt.Println(err)
		return Response400, nil
	}
	req := request.NewRequest()
	if err := s3.Put(payloadBucket, payloadPrefix + req.UUID, body); err != nil {
		fmt.Println(err)
		return Response500, nil
	}
	req.Bucket = payloadBucket
	req.Key = payloadPrefix + req.UUID
	if err := sns.PublishJSON(notifyTopic, req); err != nil {
		fmt.Println(err)
		return Response500, nil
	}
	fmt.Println("Success.")
	return &events.APIGatewayProxyResponse{StatusCode: 201, Body: req.UUID}, nil
}

func main() {
	lambda.Start(handleLambda)
}