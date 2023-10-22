package main

import (
	"os"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"gocommand/command"
	"gocommand/s3"
	"gocommand/sqs"
	"encoding/json"
)

var bucket string
var sendQueue string

type Request struct {
	UUID string
	From string
	Subject string
	Command string
}

func handleLambda(event events.SQSEvent) {
	req := Request{}
	if err := json.Unmarshal([]byte(event.Records[0].Body), &req); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Processing request %s.\n", req.UUID)
	doc, err := command.Parse(req.Command)
	if err != nil {
		fmt.Println(err)
		return
	}
	if doc == "" {
		fmt.Println("Command not found.")
		return		
	}
	if err := s3.Save(bucket, "payloads/" + req.UUID, doc); err != nil {
		fmt.Println(err)
		return
	}
	err = sqs.Send(sendQueue, event.Records[0].Body)
	fmt.Println("Done.")
}

func main() {
	bucket = os.Getenv("BUCKET")
	sendQueue = os.Getenv("SEND_QUEUE")

	lambda.Start(handleLambda)
}