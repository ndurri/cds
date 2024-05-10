package main

import (
	"reply/aws"
	"reply/request"
	"reply/response"
	"reply/mail"
	"net/smtp"
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

func getRequest(bucket, key string) (*request.Message, error) {
	content, err := aws.S3.Get(bucket, key)
	if err != nil {
		return nil, err
	}
	return request.FromJSON(string(content))
}

func processMessage(message string) {
	fmt.Println("reply: Received notification.")
	response, err := response.FromJSON(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	bucket, prefix := splitPrefix(RequestBucket)
	fmt.Printf("Retrieving request from %s:%s\n", bucket, prefix + response.RequestId)
	request, err := getRequest(bucket, prefix + response.RequestId)
	if err != nil {
		fmt.Println(err)
		return		
	}
	if request.From == "" {
		fmt.Println("Originator not found in request. Unable to send reply.")
		fmt.Println(request)
		return
	}
	fmt.Printf("Retrieving response from %s:%s\n", response.Bucket, response.Key)
	payload, err := aws.S3.Get(response.Bucket, response.Key)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := sendMail(request.From, request.Subject, string(payload)); err != nil {
		fmt.Println(err)
		return		
	}
	fmt.Println("reply: Finished.")
}

func main() {
	if err := aws.Config(); err != nil {
		panic(err)
	}
	aws.Start.SNS(processMessage)
}