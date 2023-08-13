data "aws_caller_identity" "current" {}

locals {
    account_id = data.aws_caller_identity.current.account_id
}

resource "aws_s3_bucket_policy" "ses_write" {
	bucket = aws_s3_bucket.appdata.id

	policy = jsonencode({
		Version = "2012-10-17"
	    Statement = [
			{
	        	Effect = "Allow"
	        	Principal = { "Service": "ses.amazonaws.com" }
	        	Action = "s3:PutObject"
	        	Resource = "${aws_s3_bucket.appdata.arn}/*"
	        	Condition = {
	          		StringEquals = {
	              		"AWS:SourceAccount": local.account_id
	          		}
	        	}
	      	}
	    ]
	})
}

resource "aws_ses_domain_identity" "cdsdec-uk" {
	provider = aws.ie
	domain = var.MAIL_DOMAIN
}
output "domain-verification-key" {
	value = aws_ses_domain_identity.cdsdec-uk.verification_token
}

resource "aws_ses_receipt_rule_set" "main" {
	provider = aws.ie
	rule_set_name = "cdsdec"
}

resource "aws_ses_receipt_rule" "store" {
	provider = aws.ie
	name          = "store"
	rule_set_name = "cdsdec"
	recipients    = [var.MAIL_RECIPIENT]
	enabled       = true
	scan_enabled  = true

	s3_action {
    	bucket_name = aws_s3_bucket.appdata.bucket
    	object_key_prefix = "mail-in"
    	position    = 1
	}

	depends_on = [aws_s3_bucket_policy.ses_write]
}

