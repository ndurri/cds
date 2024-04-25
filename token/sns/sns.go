package sns

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"encoding/json"
	"fmt"
	"os"
)

var client *sns.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("Failed to load SDK configuration: %v", err)
		os.Exit(1)
	}

	client = sns.NewFromConfig(cfg)	
}

func PublishJSON(topic string, obj interface{}) error {
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return Publish(topic, string(content))
}

func Publish(topic string, body string) error {
	params := &sns.PublishInput{
		TopicArn: &topic,
		Message: &body,
	}
	_, err := client.Publish(context.TODO(), params)
	return err
}
