package main

import (
	"log"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"strings"
	"github.com/google/uuid"
	"goparser/mail"
	"goparser/s3"
	"goparser/xmlfactory"
	"goparser/cds"
	"encoding/xml"
)

type Request struct {
	UUID string
	From string
	Subject string
	Command string
	DocType string
}

func handleXML(message *mail.Message, bucket string) error {
	log.Println("Found XML attachment. Sending direct.")
	var e xmlfactory.Envelope
	if err := xml.Unmarshal([]byte(message.XMLContent), &e); err != nil {
		return err
	}
	docType := e.Content.(cds.Request).DocType()
	log.Printf("Detected: %s.\n", docType)
	req := Request{
		UUID: uuid.New().String(),
		From: message.From,
		Subject: message.Subject,
		DocType: string(docType),
	}
	if err := s3.SaveJSON(bucket, "requests/" + req.UUID, req); err != nil {
		return err
	}
	if err := s3.Save(bucket, "payloads/" + req.UUID, message.XMLContent); err != nil {
		return err
	}
	return nil
}

func handleCommand(message *mail.Message, bucket string) error {
	log.Println("Found command. Sending to command processor.")
	lines := strings.Split(strings.ReplaceAll(message.TextContent, "\r\n", "\n"), "\n")
	req := Request{
		UUID: uuid.New().String(),
		From: message.From,
		Subject: message.Subject,
		Command: lines[0],
	}
	if err := s3.SaveJSON(bucket, "commands/" + req.UUID, req); err != nil {
		return err
	}
	return nil
}

func handleLambda(event events.S3Event) {
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.Key
	log.Printf("Processing %s\n", key)
	content, err := s3.Load(bucket, key)
	if err != nil {
		log.Println(err)
		return
	}
	message, err := mail.Parse(strings.NewReader(content))
	if err != nil {
		log.Println(err)
		return
	}
	if message.TextContent == "" && message.XMLContent == "" {
		log.Printf("ERROR: Unable to find any content.")
		return
	}
	if message.XMLContent != "" {
		err = handleXML(message, bucket)
	} else {
		err = handleCommand(message, bucket)
	}
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Success.")
	}
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