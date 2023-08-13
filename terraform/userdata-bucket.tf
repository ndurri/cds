resource "aws_s3_bucket" "userdata" {
  force_destroy = true
}

resource "aws_iam_policy" "userdata-read" {
	name = "userdata-read"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["s3:getObject"]
	        Effect   = "Allow"
	        Resource = "${aws_s3_bucket.userdata.arn}/*"
	      },
	    ]
	})
}

resource "aws_iam_policy" "userdata-write" {
	name = "userdata-write"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["s3:putObject", "s3:deleteObject"]
	        Effect   = "Allow"
	        Resource = "${aws_s3_bucket.userdata.arn}/*"
	      },
	    ]
	})
}