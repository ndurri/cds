package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"fmt"
	"authorize/s3"
	"net/url"
	"os"
)

var SessionBucket = os.Getenv("SESSION_BUCKET")
var SessionPrefix = os.Getenv("SESSION_PREFIX")
var AuthURL = os.Getenv("AUTH_URL")
var ClientId = os.Getenv("CLIENT_ID")
var Scope = os.Getenv("SCOPE")
var RedirectURI = os.Getenv("REDIRECT_URI")

var Response400 = &events.APIGatewayProxyResponse{StatusCode: 400, Body: "submitter not provided",}
var Response500 = &events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error",}

func lambdaMain(event *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	submitter := event.QueryStringParameters["submitter"]
	if submitter == "" {
		fmt.Printf("submitter not provided in query string %v.", event.QueryStringParameters)
		return Response400, nil
	}
	sessionId := uuid.New().String()
	fmt.Println("Saving session to %s:%s", SessionBucket, SessionPrefix + sessionId)
	if err := s3.Put(SessionBucket, SessionPrefix + sessionId, submitter); err != nil {
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