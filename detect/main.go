package main

import (
	"os"
	"fmt"
	"detect/aws"
	"detect/cds"
	"detect/xmlfactory"
	"detect/request"
	"encoding/xml"
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

func processMessage(message string) {
	fmt.Println("Received notification")
	request, err := request.FromJSON(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Processing detect request for %s:%s\n", request.Bucket, request.Key)
	payload, err := aws.S3.Get(request.Bucket, request.Key)
	if err != nil {
		fmt.Println(err)
		return
	}
	doctype, err := detect(string(payload))
	if err != nil {
		fmt.Println(err)
		return
	}
	request.DocType = doctype
	fmt.Printf("Document type is %s\n", doctype)
	content, err := request.ToJSON()
	if err != nil {
		fmt.Println(err)
		return
	}	
	if err = aws.SNS.Put(NotifyTopic, *content); err != nil {
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
	aws.Config()
	aws.Lambda.StartSNS(processMessage)
}