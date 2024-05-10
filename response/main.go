package main

import (
	"github.com/google/uuid"
	"response/aws"
	"response/response"
	"fmt"
	"os"
	"strings"

)

var ResponseBucket string = os.Getenv("RESPONSE_BUCKET")
var NotifyTopic string = os.Getenv("NEXT_TOPIC")

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
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

func processMessage(headers map[string]string, body string) int {
	fmt.Println("response: Received response.")
	convoId := getHeader(headers, "x-conversation-id")
	if convoId == "" {
		fmt.Println("ConversationId not found in headers.")
		fmt.Println(headers)
		return 201
	}
	guid := uuid.New().String()
	fmt.Printf("ConversationId is %s\n", convoId)
	fmt.Printf("GUID is %s\n", guid)
	bucket, prefix := splitPrefix(ResponseBucket)
	fmt.Printf("Saving response for %s to %s:%s\n", convoId, bucket, prefix + guid)
	if err := aws.S3.Put(bucket, prefix + guid, []byte(body)); err != nil {
		fmt.Println(err)
		return 201
	}
	res := response.Message{
		RequestId: convoId,
		Bucket: bucket,
		Key: prefix + guid,
	}
	content, err := res.ToJSON()
	if err != nil {
		fmt.Println(err)
		return 201	
	}
	if err := aws.SNS.Put(NotifyTopic, *content); err != nil {
		fmt.Println(err)
		return 201
	}
	fmt.Println("response: Finished.")
	return 201
}

func main() {
	if err := aws.Config(); err != nil {
		panic(err)
	}
	aws.Start.API(processMessage)
}