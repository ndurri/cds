package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"gocommand/command"
	"gocommand/s3"
)

type Request struct {
	UUID string
	From string
	Subject string
	Command string
	DocType string
}

func handleLambda(event events.S3Event) {
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.Key
	fmt.Printf("Processing %s\n", key)
	req := &Request{}
	if err := s3.LoadJSON(bucket, key, req); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Processing request %s.\n", req.UUID)
	doc, doctype, err := command.Parse(req.Command)
	if err != nil {
		fmt.Println(err)
		return
	}
	if doc == "" {
		fmt.Println("Command not found.")
		return		
	}
	req.DocType = string(doctype)
	if err := s3.SaveJSON(bucket, "requests/" + req.UUID, &req); err != nil {
		fmt.Println(err)
		return
	}
	if err := s3.Save(bucket, "payloads/" + req.UUID, doc); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Done.")
}

func main() {
	lambda.Start(handleLambda)
}