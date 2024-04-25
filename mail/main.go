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
	"mail/sns"
	"strings"
)

var PayloadBucket = os.Getenv("PAYLOAD_BUCKET")
var NotifyTopic = os.Getenv("NEXT_TOPIC")

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
}

func handleXML(req *request.Request, xml string) error {
	bucket, prefix := splitPrefix(PayloadBucket)
	if err := s3.Put(bucket, prefix + req.UUID, xml); err != nil {
		return err
	}
	req.Bucket = bucket
	req.Key = prefix + req.UUID
	if err := sns.PublishJSON(NotifyTopic, req); err != nil {
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

func handleLambda(event events.S3Event) {
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.URLDecodedKey
	fmt.Printf("Received S3 notification from %s:%s\n", bucket, key)
	rawMail, err := s3.Get(bucket, key)
	if err != nil {
		fmt.Println(err)
		fmt.Println(event)
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