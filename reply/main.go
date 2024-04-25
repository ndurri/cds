package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"reply/s3"
	"reply/request"
	"reply/response"
	"reply/mail"
	"net/smtp"
	"encoding/json"
	"os"
	"fmt"
	"html"
	"strings"
)

var (
	SMTPUsername string = os.Getenv("SMTP_USERNAME")
	SMTPPassword string = os.Getenv("SMTP_PASSWORD")
	SMTPHost string = os.Getenv("SMTP_HOST")
	SMTPPort string = os.Getenv("SMTP_PORT")
	MailDomain string = os.Getenv("MAIL_DOMAIN")
	MailSender string = os.Getenv("MAIL_SENDER")
	RequestBucket string = os.Getenv("REQUEST_BUCKET")
)

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
}

func sendMail(to string, subject string, payload string) error {
	message := mail.NewMessage(MailDomain, MailSender, to)
	message.Subject = subject
	/*attachment := mail.Attachment{
		ContentType: "application/xml",
		Filename: "response.xml",
		Content: payload,
	}
	message.AddAttachment(attachment)*/
	message.TextContent = "Your response is attached"
	message.HTMLContent = html.EscapeString(payload)
	auth := smtp.PlainAuth("", SMTPUsername, SMTPPassword, SMTPHost)
	mime := message.Unmarshal()
	if err := smtp.SendMail(SMTPHost + ":" + SMTPPort, auth, MailSender, []string{to}, mime); err != nil {
		return err
	}
	return nil
}

func lambdaMain(event events.SNSEvent) {
	fmt.Printf("Received notification on topic %s\n", event.Records[0].SNS.TopicArn)
	response := response.Message{}
	if err := json.Unmarshal([]byte(event.Records[0].SNS.Message), &response); err != nil {
		fmt.Println(err)
		return
	}
	bucket, prefix := splitPrefix(RequestBucket)
	fmt.Printf("Retrieving request from %s:%s\n", bucket, prefix + response.RequestId)
	request := request.Message{}
	if err := s3.GetJSON(bucket, prefix + response.RequestId, &request); err != nil {
		fmt.Println(err)
		return
	}
	if request.From == "" {
		fmt.Println("Originator not found in request. Unable to send reply.")
		fmt.Println(request)
		return
	}
	fmt.Printf("Retrieving response from %s:%s\n", response.Bucket, response.Key)
	payload, err := s3.Get(response.Bucket, response.Key)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := sendMail(request.From, request.Subject, payload); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Println("Finished.")
}

func main() {
	lambda.Start(lambdaMain)
}