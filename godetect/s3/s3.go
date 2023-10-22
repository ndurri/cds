package s3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"strings"
	"encoding/json"
	"log"
	"time"
)

var client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}

	client = s3.NewFromConfig(cfg)	
}

func LoadJSON(bucket string, key string, result interface{}) error {
	content, err := Load(bucket, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(content), result)
}

func Load(bucket string, key string) (string, error) {
	content, _, err := LoadWithTime(bucket, key)
	return content, err
}

func LoadWithTime(bucket string, key string) (string, *time.Time, error) {
	params := &s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	}
	res, err := client.GetObject(context.TODO(), params)
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}
	return string(content), res.LastModified, nil
}

func SaveJSON(bucket string, key string, body interface{}) error {
	content, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return Save(bucket, key, string(content))
}

func Save(bucket string, key string, body string) error {
	r := strings.NewReader(body)
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key: &key,
		Body: r,
	}
	_, err := client.PutObject(context.TODO(), params)
	return err
}