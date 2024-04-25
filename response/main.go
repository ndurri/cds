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
var NotifyTopic string = os.Getenv("NEXT_TOPIC")

var ResponseCreated = &events.APIGatewayProxyResponse{StatusCode: 201}

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
}

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
			fmt.Println(key)
			return value
		}
	}
	return ""
}

func handler(event *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error){
	fmt.Printf("Received response from %s\n", event.RequestContext.Identity.SourceIP)
	convoId := getHeader(event.Headers, "x-conversation-id")
	if convoId == "" {
		fmt.Println("ConversationId not found in headers.")
		fmt.Println(event.Headers)
		return ResponseCreated, nil
	}
	guid := uuid.New().String()
	fmt.Printf("ConversationId is %s\n", convoId)
	fmt.Printf("GUID is %s\n", guid)
	body, err := decodeBody(event)
	if err != nil {
		fmt.Println(err)
		return ResponseCreated, nil
	}
	bucket, prefix := splitPrefix(ResponseBucket)
	fmt.Printf("Saving response for %s to %s:%s\n", convoId, bucket, prefix + guid)
	if err := s3.Put(bucket, prefix + guid, body); err != nil {
		fmt.Println(err)
		return ResponseCreated, nil
	}
	res := response.Message{
		RequestId: convoId,
		Bucket: bucket,
		Key: prefix + guid,
	}
	if err := sns.PublishJSON(NotifyTopic, &res); err != nil {
		fmt.Println(err)
		return ResponseCreated, nil
	}
	fmt.Println("Finished.")
	return ResponseCreated, nil
}

func main() {
	lambda.Start(handler)
}