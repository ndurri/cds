# CDS
## A fully serverless Email Interface for the UK Customs Declarations Service

## Deployment Instructions

### Pre-requisites
1. An AWS Account
2. A custom domain
3. An outbound email provider with an SMTP gateway (e.g. brevo.com)
4. Terraform installed.
5. A go compiler.
6. A HMRC developer account and an application configured to access Customs APIs.

### Set up your AWS account
It is recommended that you create a fresh account specifically for this service to avoid name clashes and unexpected side-effects with other objects in your main account. Read the AWS documentation on Organisations for information on how to set up an organisation and new accounts.
1. In your preferrably new, dedicated account, create a terraform user.
2. Give the user admin rights within that account.
3. Create an API key ID and Secret for the user.
4. In your local home directory, create an ~/.aws/credentials file if one doesn't already exist.
5. Add a section for CDS (called cds) and add the terraform credentials.
6. For simplicity also add the aws region you will be building most of your objects in - most likely eu-west-2.

### Build Your Services
1. Clone this repo to your local environment.
2. Create a terraform.tfvars in your terraform directory. This must contain the following:  
	CLIENT_ID = *the client id associated with your HMRC application*  
	REDIRECT_URI = *the oauth redirect URL set up for your application* e.g. https://your.domain/oauth/redirect  
	CLIENT_SECRET = *the client secret associated with your HMRC application*  
	MAIL_DOMAIN = *the email domain your users will use to send emails to your application* e.g. your.domain  
	MAIL_RECIPIENT = *the email address your users will used to send emails to your application* e.g. cds@your.domain  
	SMTP_HOST = *SMTP host name of your outbound email provider*  
	SMTP_PORT = *SMTP port for your outbound email provider*  
	SMTP_USER = *username of your outbound email account*  
	SMTP_PASSWORD = *password for your outbound email account*  
3. Compile Lambdas where necessary.  
	3.1 gocommand  
		cd to the gocommand directory  
		go mod tidy  
		GOOS=linux GOARCH=amd64 go build -o bootstrap  
	3.2 goparser  
		cd to the goparser directory  
		go mod tidy  
		GOOS=linux GOARCH=amd64 go build -o bootstrap  
	3.3 reply  
		cd to the reply directory  
		zip ../reply.zip \*.mjs  
	3.4 submit  
		cd to the submit directory  
		zip ../submit.zip \*.mjs apis.json  
4. Set your AWS environment to point to your new AWS account - export AWS_PROFILE=cds
5. run terraform init
6. run terraform validate
7. run terraform plan
8. Review the output. Terraform should be planning to create some s3 resources, permissions, lambdas and an API gateway.
9. Assuming you're happy with the plan, run terraform apply
10. Once finished, all your main functionality should be built in AWS.

### Post Build Actions
For HMRC to use your callback URL for responses, you must provide an edge-optimised custom domain.
1. Log into the AWS console, switch to us-east-1.
2. Navigate to Certificate Manager. Request an HTTPS certificate for your.domain.
3. Switch back to your normal region and navigate to API Gateway. Create a custom domain for your.domain.
4. Attach the certificate you created in step 2.
5. Create an API mapping between the custom domain and your cds api.
6. Note your API gateway domain name (yyy.cloudfront.net)
7. Switch to eu-west-1 and navigate to the Simple Email Service.
8. Go to Verified Identities and follow the DKIM procedure for validating your domain.
9. Go to your domain provider and enter a CNAME record for your domain to point to the name from step 6. This will mean that all https traffic to your.domain will be sent to the AWS API.
10. Enter an MX record to point to the Amazon SES host available from the AWS documentation.
11. Enter the three CNAME entries required for DKIM from step 8.
12. You will need to wait for AWS to confirm your DKIM settings before you can send emails to your.domain. This can take up to an hour.
13. Log in to your HMRC developer account and update your redirect and callback URLs to https://your.domain/oauth/redirect and https://your.domain/hmrc respectively.
14. Create test users with EORIs GB906263308468 and GB417869120000.
15. Call https://your.domain/oauth/authorize?submitter=GB906263308468 and go through the authorisation process for both EORIs. All being well, this will create 2 oauth tokens in your appdata bucket that will be used when you submit declarations or movements by email.
16. Send a test email to cds@your.domain with subject "test" and body exp. You should be able to monitor progress in the cloudwatch logs for your lambdas.


