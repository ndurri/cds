package sqs

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"encoding/json"
	"log"
)

var client *sqs.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}

	client = sqs.NewFromConfig(cfg)	
}

func SendJSON(queue string, body interface{}) error {
	content, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return Send(queue, string(content))
}

func Send(queue string, body string) error {
	params := &sqs.SendMessageInput{
		QueueUrl: &queue,
		MessageBody: &body,
	}
	_, err := client.SendMessage(context.TODO(), params)
	return err
}