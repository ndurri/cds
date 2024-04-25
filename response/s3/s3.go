package s3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"strings"
)

var client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	client = s3.NewFromConfig(cfg)	
}

func Put(bucket string, key string, body string) error {
	r := strings.NewReader(body)
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key: &key,
		Body: r,
	}
	_, err := client.PutObject(context.TODO(), params)
	return err
}