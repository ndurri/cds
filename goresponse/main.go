package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"goresponse/s3"
	"fmt"
	"os"
	"encoding/base64"
	"strings"
)

var Bucket string

var ErrorResponse = &events.APIGatewayProxyResponse{StatusCode: 500,}
var OKResponse = &events.APIGatewayProxyResponse{StatusCode: 200,}

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

func getHeader(headers map[string]string, name string) string {
	name = strings.ToLower(name)
	for key, value := range headers {
		if strings.ToLower(key) == name {
			return value
		}
	}
	return ""
}

func handler(event *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	convoId := getHeader(event.Headers, "x-conversation-id")
	guid := uuid.New().String()
	fmt.Printf("Processing %s/%s", convoId, guid)
	body, err := decodeBody(event)
	if err != nil {
		fmt.Println(err)
		return ErrorResponse, err
	}
	if err := s3.Save(Bucket, "payload-in/" + convoId + "/" + guid, body); err != nil {
		fmt.Println(err)
		return ErrorResponse, err
	}
	return OKResponse, nil
}

func main() {
	Bucket = os.Getenv("BUCKET")
	lambda.Start(handler)
}