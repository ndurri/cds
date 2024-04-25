package request

type Message struct {
	From string
	Subject string
	Bucket string
	Key string
	DocType string
	AccessToken string
	Id string
	Status int
	Body string
}
