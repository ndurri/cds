resource "aws_s3_bucket" "appdata" {
  force_destroy = true
}

resource "aws_s3_bucket_lifecycle_configuration" "appdata" {
	bucket = aws_s3_bucket.appdata.id

	rule {
		id = "expire30day"
		filter {}
		expiration {
			days = 30
		}
		status = "Enabled"
	}
}

resource "aws_iam_policy" "appdata-read" {
	name = "appdata-read"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["s3:getObject"]
	        Effect   = "Allow"
	        Resource = "${aws_s3_bucket.appdata.arn}/*"
	      },
	    ]
	})
}

resource "aws_iam_policy" "appdata-write" {
	name = "appdata-write"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["s3:putObject", "s3:deleteObject"]
	        Effect   = "Allow"
	        Resource = "${aws_s3_bucket.appdata.arn}/*"
	      },
	    ]
	})
}

resource "aws_s3_bucket_notification" "appdata_notifications" {
  bucket = aws_s3_bucket.appdata.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.processMail.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "mail-in/"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.processCommand.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "commands/"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.submit.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "payloads/"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.reply.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "payload-in/"
  }
}