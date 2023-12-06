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
)

var SMTPUsername string = os.Getenv("SMTP_USERNAME")
var SMTPPassword string = os.Getenv("SMTP_PASSWORD")
var SMTPHost string = os.Getenv("SMTP_HOST")
var SMTPPort string = os.Getenv("SMTP_PORT")
var MailDomain string = os.Getenv("MAIL_DOMAIN")
var MailSender string = os.Getenv("MAIL_SENDER")
var RequestBucket string = os.Getenv("REQUEST_BUCKET")
var RequestPrefix string = os.Getenv("REQUEST_PREFIX")

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
	fmt.Printf("Retrieving request from %s:%s\n", RequestBucket, RequestPrefix + response.RequestId)
	request := request.Message{}
	if err := s3.GetJSON(RequestBucket, RequestPrefix + response.RequestId, &request); err != nil {
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