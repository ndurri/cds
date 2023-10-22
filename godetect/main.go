package main

import (
	"os"
	"fmt"
	"godetect/cds"
	"godetect/xmlfactory"
	"godetect/s3"
	"godetect/sqs"
	"encoding/xml"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
)

var bucket string
var sendQueue string

type Request struct {
	UUID string
	From string
	Subject string
	Command string
	Submitter string
	DocType string
}

func detect(payload string) (string, error) {
	var e xmlfactory.Envelope
	if err := xml.Unmarshal([]byte(payload), &e); err != nil {
		return "", err
	}
	docType := e.Content.(cds.Request).DocType()
	fmt.Printf("Detected: %s.\n", docType)
	return string(docType), nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func lambdaMain(event events.SQSEvent) {
	request := Request{}
	if err := json.Unmarshal([]byte(event.Records[0].Body), &request); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Processing request %s\n", request.UUID)
	payload, err := s3.Load(bucket, "payloads/" + request.UUID)
	checkError(err)
	doctype, err := detect(payload)
	checkError(err)
	request.DocType = doctype
	err = sqs.SendJSON(sendQueue, request)
	fmt.Println("Finished.")
}

func init() {
	xmlfactory.Register("inventoryLinkingConsolidationRequest", (*cds.Consolidation)(nil))
	xmlfactory.Register("inventoryLinkingMovementRequest", (*cds.Movement)(nil))
	xmlfactory.Register("inventoryLinkingQueryRequest", (*cds.Query)(nil))
	xmlfactory.Register("MetaData", (*cds.MetaData)(nil))	
}

func main() {
	bucket = os.Getenv("BUCKET")
	sendQueue = os.Getenv("SEND_QUEUE")
	lambda.Start(lambdaMain)
}