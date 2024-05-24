package main

import (
	"oauth/aws"
	"strings"
	"github.com/google/uuid"
	URL "net/url"
	"os"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"io"
	"encoding/json"
)

var (
	SessionBucket = mustHave("SESSION_BUCKET")
	SessionPrefix = mustHave("SESSION_PREFIX")
	AuthURL = mustHave("AUTH_URL")
	ClientId = mustHave("CLIENT_ID")
	ClientSecret = mustHave("CLIENT_SECRET")
	Scope = mustHave("SCOPE")
	RedirectURI = mustHave("REDIRECT_URI")
	TokenHost = mustHave("TOKEN_HOST")
	TokenBucket = mustHave("TOKEN_BUCKET")
	TokenPrefix = mustHave("TOKEN_PREFIX")
	TestUserHost = mustHave("TESTUSER_HOST")
)

var UUIDRegex = regexp.MustCompile("^[0-9a-f]{8}\\b-[0-9a-f]{4}\\b-[0-9a-f]{4}\\b-[0-9a-f]{4}\\b-[0-9a-f]{12}$")

func mustHave(varname string) string {
	var value string
	if value = os.Getenv(varname); value == "" {
		panic(errors.New(varname + " does not exist."))
	}
	return value
}

type Token struct {
	AccessToken string `json:"access_token"`
}

type TestUserSpec struct {
	ServiceNames []string `json:"serviceNames"`
	EORI string `json:"eoriNumber"`
}

func getAccessToken() (*string, error) {
	tokenJSON, err := tokenForParams(URL.Values{
		"client_id":     {ClientId},
		"client_secret": {ClientSecret},
		"grant_type":    {"client_credentials"},
	})
	if err != nil {
		return nil, err
	}
	token := Token{}
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, err
	}
	return &token.AccessToken, nil
}

func tokenForCode(code string) ([]byte, error) {
	return tokenForParams(URL.Values{
		"client_id":     {ClientId},
		"client_secret": {ClientSecret},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {RedirectURI},
		"code":          {code},
	})
}

func tokenForParams(params URL.Values) ([]byte, error) {
	res, err := http.PostForm(TokenHost, params)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err		
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		errm := fmt.Sprintf("Call to %s failed with status %d.\nServer returned %s.", TokenHost, res.StatusCode, content)
		return nil, errors.New(errm)
	}
	return content, nil
}

func processTestUser(url string, headers map[string]string, params map[string]string, body string) (int, map[string]string, string) {
	submitter := params["submitter"]
	if submitter == "" {
		fmt.Println(errors.New("submitter not provided"))
		return 400, nil, ""
	}
	spec := TestUserSpec{ServiceNames: []string{"customs-services"}, EORI: submitter}
	content, err := json.Marshal(spec)
	if err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	token, err := getAccessToken()
	if err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}	
	reader := strings.NewReader(string(content))
	req, err := http.NewRequest(http.MethodPost, TestUserHost, reader)
	if err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	req.Header.Add("authorization", "Bearer " + *token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	defer res.Body.Close()
	content, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	if res.StatusCode != 201 {
		fmt.Printf("%s returned %d.", TestUserHost, res.StatusCode)
		fmt.Println(string(content))
		return 500, nil, ""		
	}
	fmt.Println("oauth: Finished.")
	return 201, map[string]string{"Content-Type": "application/json"}, string(content)
}

func processRedirect(url string, headers map[string]string, params map[string]string, body string) (int, map[string]string, string) {
	var (code, state string; session []byte; err error)
	if code = params["code"]; code == "" {
		fmt.Println("Code not provided in auth-redirect.")
		fmt.Println(params)
		return 400, nil, "Code not provided"
	}
	if state = params["state"]; state == "" || !UUIDRegex.MatchString(state) {
		fmt.Println("State not provided or malformed in auth-redirect.")
		fmt.Println(params)
		return 400, nil, "State not provided or malformed"
	}
	fmt.Printf("Retrieving session from %s:%s.", SessionBucket, SessionPrefix + state)
	if session, err = aws.S3.Get(SessionBucket, SessionPrefix + state); err != nil {
		fmt.Println(err)
		return 400, nil, "state parameter invalid."
	}	
	fmt.Printf("Exchanging code %s for token", code)
	token, err := tokenForCode(code)
	if err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	fmt.Printf("Saving token to %s:%s.", TokenBucket, TokenPrefix + string(session))
	if err = aws.S3.Put(TokenBucket, TokenPrefix + string(session), token); err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	fmt.Println("oauth: Finished.")
	return 200, nil, "Thankyou for authorising. Requests can now be submitted using " + string(session) + "."
}

func processAuthorize(url string, headers map[string]string, params map[string]string, body string) (int, map[string]string, string) {
	submitter := params["submitter"]
	if submitter == "" {
		fmt.Println(errors.New("submitter not provided"))
		return 400, nil, ""
	}
	sessionId := uuid.New().String()
	fmt.Printf("Saving session to %s:%s", SessionBucket, SessionPrefix + sessionId)
	if err := aws.S3.Put(SessionBucket, SessionPrefix + sessionId, []byte(submitter)); err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	location, err := URL.Parse(AuthURL)
	if err != nil {
		fmt.Println(err)
		return 500, nil, ""
	}
	urlparams := URL.Values{
		"response_type": {"code"},
		"client_id": {ClientId},
		"scope": {Scope},
		"redirect_uri": {RedirectURI},
		"state": {sessionId},
	}
	location.RawQuery = urlparams.Encode()
	fmt.Println("oauth: Finished.")
	return 302, map[string]string{"location": location.String()}, ""
}

func processMessage(url string, headers map[string]string, params map[string]string, body string)(int, map[string]string, string) {
	if strings.HasPrefix(url, "/oauth/authorize") {
		return processAuthorize(url, headers, params, body)
	} else if strings.HasPrefix(url, "/oauth/auth-redirect") {
		return processRedirect(url, headers, params, body)		
	} else if strings.HasPrefix(url, "/oauth/testuser") {
		return processTestUser(url, headers, params, body)		
	}
	return 404, nil, ""
}

func main() {
	if err := aws.Config(); err != nil {
		panic(err)
	}
	aws.Start.APIv2(processMessage)
}