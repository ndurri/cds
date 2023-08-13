package main

import (
	"os"
	"log"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"gocommand/cds"
	"gocommand/command"
	"gocommand/s3"
)

var decSubmitter = os.Getenv("DEC_SUBMITTER")
var movSubmitter = os.Getenv("MOV_SUBMITTER")

type Request struct {
	UUID string
	From string
	Subject string
	Command string
	Submitter string
	DocType   cds.DocType
}

func handleLambda(event events.S3Event) {
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.Key
	req := Request{}
	if err := s3.LoadJSON(bucket, key, &req); err != nil {
		log.Println(err)
		return
	}
	doc, doctype, err := command.Parse(req.Command)
	if err != nil {
		log.Println(err)
		return
	}
	if doc == "" {
		log.Println("Command not found.")
		return		
	}
	req.DocType = doctype
	if doctype == cds.MovementType {
		req.Submitter = movSubmitter
	} else {
		req.Submitter = decSubmitter
	}
	if err := s3.SaveJSON(bucket, "requests/" + req.UUID, req); err != nil {
		log.Println(err)
		return
	}
	if err := s3.Save(bucket, "payloads/" + req.UUID, doc); err != nil {
		log.Println(err)
		return
	}
}

func main() {
	lambda.Start(handleLambda)
}