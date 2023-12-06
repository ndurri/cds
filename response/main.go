package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"response/s3"
	"response/sns"
	"response/response"
	"fmt"
	"os"
	"encoding/base64"
	"strings"

)

var ResponseBucket string = os.Getenv("RESPONSE_BUCKET")
var ResponsePrefix string = os.Getenv("RESPONSE_PREFIX")
var notifyTopic string = os.Getenv("NOTIFY_TOPIC")

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

func handler(event *events.APIGatewayProxyRequest) {
	sourceIP := event.RequestContext.Identity.SourceIP
	fmt.Printf("Received response from %s\n", sourceIP)
	convoId := getHeader(event.Headers, "x-conversation-id")
	if convoId == "" {
		fmt.Println("ConversationId not found in headers.")
		fmt.Println(event.Headers)
		return		
	}
	guid := uuid.New().String()
	fmt.Printf("ConversationId is %s\n", convoId)
	fmt.Printf("GUID is %s\n", guid)
	body, err := decodeBody(event)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := s3.Put(ResponseBucket, ResponsePrefix + guid, body); err != nil {
		fmt.Println(err)
		return
	}
	res := response.Message{
		RequestId: convoId,
		Bucket: ResponseBucket,
		Key: ResponsePrefix + guid,
	}
	if err := sns.PublishJSON(notifyTopic, &res); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Finished.")
}

func main() {
	lambda.Start(handler)
}