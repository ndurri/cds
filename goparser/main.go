package main

import (
	"os"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"strings"
	"github.com/google/uuid"
	"goparser/mail"
	"goparser/s3"
	"goparser/sqs"
)

var commandQueue string
var payloadQueue string

type Request struct {
	UUID string
	From string
	Subject string
	Command string
}

func handleXML(message *mail.Message, bucket string) error {
	fmt.Println("Found XML attachment. Sending direct.")
	req := Request{
		UUID: uuid.New().String(),
		From: message.From,
		Subject: message.Subject,
	}
	if err := s3.Save(bucket, "payloads/" + req.UUID, message.XMLContent); err != nil {
		return err
	}
	if err := sqs.SendJSON(payloadQueue, req); err != nil {
		return err
	}
	return nil
}

func handleCommand(message *mail.Message, bucket string) error {
	fmt.Println("Found command. Sending to command processor.")
	lines := strings.Split(strings.ReplaceAll(message.TextContent, "\r\n", "\n"), "\n")
	req := Request{
		UUID: uuid.New().String(),
		From: message.From,
		Subject: message.Subject,
		Command: lines[0],
	}
	if err := sqs.SendJSON(commandQueue, req); err != nil {
		return err
	}
	return nil
}

func handleLambda(event events.S3Event) {
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.Key
	fmt.Printf("Processing %s\n", key)
	content, err := s3.Load(bucket, key)
	if err != nil {
		fmt.Println(err)
		return
	}
	message, err := mail.Parse(strings.NewReader(content))
	if err != nil {
		fmt.Println(err)
		return
	}
	if message.TextContent == "" && message.XMLContent == "" {
		fmt.Printf("ERROR: Unable to find any content.")
		return
	}
	if message.XMLContent != "" {
		err = handleXML(message, bucket)
	} else {
		err = handleCommand(message, bucket)
	}
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success.")
	}
}

func main() {
	commandQueue = os.Getenv("COMMAND_QUEUE")
	payloadQueue = os.Getenv("PAYLOAD_QUEUE")
	lambda.Start(handleLambda)
}