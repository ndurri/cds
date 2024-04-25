package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"fmt"
	"authorize/s3"
	"net/url"
	"os"
	"strings"
)

var (
	SessionBucket = os.Getenv("SESSION_BUCKET")
	AuthURL = os.Getenv("AUTH_URL")
	ClientId = os.Getenv("CLIENT_ID")
	Scope = os.Getenv("SCOPE")
	RedirectURI = os.Getenv("REDIRECT_URI")
)

func splitPrefix(path string) (string, string) {
	elems := strings.Split(path, ":")
	if len(elems) == 1 {
		return path, ""
	}
	return elems[0], elems[1]
}

var Response400 = &events.APIGatewayProxyResponse{StatusCode: 400, Body: "Submitter not provided",}
var Response500 = &events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error",}

func lambdaMain(event *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	submitter, prs := event.QueryStringParameters["submitter"]
	if !prs || submitter == "" {
		fmt.Printf("submitter not provided in query string %v.", event.QueryStringParameters)
		return Response400, nil
	}
	sessionId := uuid.New().String()
	bucket, prefix := splitPrefix(SessionBucket)
	fmt.Println("Saving session to %s:%s", bucket, prefix + sessionId)
	if err := s3.Put(bucket, prefix + sessionId, submitter); err != nil {
		fmt.Println(err)
		return Response500, nil
	}
	location, err := url.Parse(AuthURL)
	if err != nil {
		fmt.Println(err)
		return Response500, nil	
	}
	params := url.Values{
		"response_type": {"code"},
		"client_id": {ClientId},
		"scope": {Scope},
		"redirect_uri": {RedirectURI},
		"state": {sessionId},
	}
	location.RawQuery = params.Encode()
	return &events.APIGatewayProxyResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"location": location.String(),
		},
	}, nil
}

func main() {
	lambda.Start(lambdaMain)
}