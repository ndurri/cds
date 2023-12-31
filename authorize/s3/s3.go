package s3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"strings"
	"io"
	"os"
	"fmt"
	"encoding/json"
)

var client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("FATAL: Failed to load SDK configuration, %v", err)
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

func GetJSON(bucket string, key string, obj interface{}) error {
	content, err := Get(bucket, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(content), obj); err != nil {
		return err
	}
	return nil
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