package main

import (
	"os"
	"fmt"
	"mail/aws"
	"mail/request"
	"mail/mail"
	"mail/command"
	"strings"
	"encoding/json"
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
	if err := aws.S3.Put(bucket, prefix + req.UUID, []byte(xml)); err != nil {
		return err
	}
	req.Bucket = bucket
	req.Key = prefix + req.UUID
	content, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if err := aws.SNS.Put(NotifyTopic, string(content)); err != nil {
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

func processRequest(bucket, key string) {
	fmt.Printf("Received S3 notification from %s:%s\n", bucket, key)
	rawMail, err := aws.S3.Get(bucket, key)
	if err != nil {
		fmt.Println(err)
		return
	}
	parsed, err := mail.Parse(string(rawMail))
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
	if err := aws.Config(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	aws.Lambda.StartS3(processRequest)
}