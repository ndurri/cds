package s3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"os"
	"fmt"
)

var client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("Failed to load SDK configuration: %v\n", err)
		os.Exit(1)
	}

	client = s3.NewFromConfig(cfg)	
}

func Get(bucket string, key string) (string, error) {
	params := &s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	}
	res, err := client.GetObject(context.TODO(), params)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
