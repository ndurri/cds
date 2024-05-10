package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"io"
	"bytes"
)

var s3Client *s3.Client
var snsClient *sns.Client

func Config() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	s3Client = s3.NewFromConfig(cfg)
	snsClient = sns.NewFromConfig(cfg)
	return nil
}

func S3Get(bucket, key string) ([]byte, error) {
	params := &s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	}
	goo, err := s3Client.GetObject(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	defer goo.Body.Close()
	content, err := io.ReadAll(goo.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func S3Put(bucket, key string, content []byte) error {
	reader := bytes.NewReader(content)
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key: &key,
		Body: reader,
	}
	_, err := s3Client.PutObject(context.TODO(), params)
	if err != nil {
		return err
	}
	return nil
}

func SNSPut(topic, content string) error {
	params := &sns.PublishInput{
		TopicArn: &topic,
		Message: &content,
	}
	_, err := snsClient.Publish(context.TODO(), params)
	return err
}