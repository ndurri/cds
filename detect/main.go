package main

import (
	"os"
	"fmt"
	"detect/cds"
	"detect/xmlfactory"
	"detect/request"
	"detect/s3"
	"detect/sns"
	"encoding/xml"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
)

var NotifyTopic string = os.Getenv("NEXT_TOPIC")

func detect(payload string) (string, error) {
	var e xmlfactory.Envelope
	if err := xml.Unmarshal([]byte(payload), &e); err != nil {
		return "", err
	}
	docType := e.Content.(cds.Request).DocType()
	return string(docType), nil
}

func handleLambda(event events.SNSEvent) {
	fmt.Printf("Received notification on topic %s\n", event.Records[0].SNS.TopicArn)
	request, err := request.FromJSON(event.Records[0].SNS.Message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Processing detect request for %s:%s\n", request.Bucket, request.Key)
	payload, err := s3.Get(request.Bucket, request.Key)
	if err != nil {
		fmt.Println(err)
		return
	}
	doctype, err := detect(payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	request.DocType = doctype
	fmt.Printf("Document type is %s\n", doctype)
	err = sns.PublishJSON(NotifyTopic, request)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Success.")
}

func init() {
	xmlfactory.Register("inventoryLinkingConsolidationRequest", (*cds.Consolidation)(nil))
	xmlfactory.Register("inventoryLinkingMovementRequest", (*cds.Movement)(nil))
	xmlfactory.Register("inventoryLinkingQueryRequest", (*cds.Query)(nil))
	xmlfactory.Register("MetaData", (*cds.MetaData)(nil))	
}

func main() {
	lambda.Start(handleLambda)
}