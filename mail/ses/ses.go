package ses

import (
	"encoding/json"
)

type Notification struct {
	Receipt struct {
		Action struct {
			TopicARN string `json:"topicArn"`
			BucketName string `json:"bucketName"`
			ObjectKey string `json:"objectKey"`
		} `json:"action"`
	} `json:"receipt"`
}

func FromJSON(content string) (*Notification, error) {
	notification := &Notification{}
	if err := json.Unmarshal([]byte(content), notification); err != nil {
		return nil, err
	}
	return notification, nil
}

func (n *Notification) Topic() string {
	return n.Receipt.Action.TopicARN
}

func (n *Notification) Bucket() string {
	return n.Receipt.Action.BucketName
}

func (n *Notification) Key() string {
	return n.Receipt.Action.ObjectKey
}