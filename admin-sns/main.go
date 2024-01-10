package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"fmt"
	"os"
	"admin-sns/sns"
)

var Response201 = &events.APIGatewayProxyResponse{StatusCode: 201,}
var Response400 = &events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request",}
var Response500 = &events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error",}

var topicARN = os.Getenv("TOPIC_ARN")
var sourceBucket = os.Getenv("SOURCE_BUCKET")
var keyPrefix = os.Getenv("KEY_PREFIX")

type SESMessage struct {
	Receipt struct {
		Action struct {
			TopicARN string `json:"topicArn"`
			BucketName string `json:"bucketName"`
			ObjectKey string `json:"objectKey"`
		} `json:"action"`
	} `json:"receipt"`
}

func lambdaMain(event *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Printf("Received request for %s from %s.\n", event.Path, event.RequestContext.Identity.SourceIP)
	key, prs := event.PathParameters["key"]
	if !prs {
		fmt.Println("ERROR: Key path parameter missing.")
		return Response400, nil
	}
	message := SESMessage{}
	message.Receipt.Action.TopicARN = topicARN
	message.Receipt.Action.BucketName = sourceBucket
	message.Receipt.Action.ObjectKey = keyPrefix + "/" + key
	fmt.Printf("Sending replay request for %s.\n", message.Receipt.Action.ObjectKey)
	if err := sns.PublishJSON(topicARN, message); err != nil {
		fmt.Println(err)
		return Response500, nil		
	}
	fmt.Println("Success.")
	return Response201, nil
}

func main() {
	lambda.Start(lambdaMain)
}