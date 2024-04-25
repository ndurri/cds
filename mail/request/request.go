package request

import (
	"github.com/google/uuid"
)

type Request struct {
	UUID string
	From string
	Subject string
	Bucket string
	Key string
}

func NewRequest() *Request {
	return &Request{
		UUID: uuid.New().String(),
	}
}