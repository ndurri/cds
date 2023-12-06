package main

import (
	"os"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"mail/request"
	"mail/mail"
	"mail/command"
	"mail/s3"
	"mail/ses"
	"mail/sns"
)

var payloadBucket = os.Getenv("PAYLOAD_BUCKET")
var payloadPrefix = os.Getenv("PAYLOAD_PREFIX")
var notifyTopic = os.Getenv("NOTIFY_TOPIC")

func handleXML(req *request.Request, xml string) error {
	if err := s3.Put(payloadBucket, payloadPrefix + req.UUID, xml); err != nil {
		return err
	}
	req.Bucket = payloadBucket
	req.Key = payloadPrefix + req.UUID
	if err := sns.PublishJSON(notifyTopic, req); err != nil {
		return err
	}
	return nil
}

func handleCommand(req *request.Request, text string) error {
	xml, err := command.Parse(text)
	if err != nil {
		return err
	}
	return handleXML(req, xml)
}

func handleLambda(event events.SNSEvent) {
	message, err := ses.FromJSON(event.Records[0].SNS.Message)
	if err != nil {
		fmt.Println(err)
		return		
	}
	topic := message.Topic()
	bucket := message.Bucket()
	key := message.Key()
	if bucket == "" || key == "" {
		fmt.Println("Bucket and/or key not provided.")
		fmt.Println(event)
		return
	}
	fmt.Printf("Received SES notification for %s:%s on topic %s\n", bucket, key, topic)
	rawMail, err := s3.Get(bucket, key)
	if err != nil {
		fmt.Println(err)
		return
	}
	parsed, err := mail.Parse(rawMail)
	if err != nil {
		fmt.Println(err)
		return
	}
	if parsed.TextContent == "" && parsed.XMLContent == "" {
		fmt.Printf("ERROR: Unable to find any content.")
		return
	}
	req := request.NewRequest()
	req.From = parsed.From
	req.Subject = parsed.Subject
	if parsed.XMLContent != "" {
		fmt.Println("Found XML attachment. Sending direct.")
		err = handleXML(req, parsed.XMLContent)
	} else {
		fmt.Println("Found command. Sending to command processor.")
		err = handleCommand(req, parsed.TextContent)
	}
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success.")
	}
}

func main() {
	lambda.Start(handleLambda)
}