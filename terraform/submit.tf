resource "aws_iam_role" "submit" {
	name = "submit"
    assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "submit-basic" {
  role       = aws_iam_role.submit.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "submit-appdata-read" {
	role       = aws_iam_role.submit.name
	policy_arn = aws_iam_policy.appdata-read.arn
}

resource "aws_iam_role_policy_attachment" "submit-appdata-write" {
	role       = aws_iam_role.submit.name
	policy_arn = aws_iam_policy.appdata-write.arn
}

resource "aws_iam_role_policy_attachment" "submit-userdata-read" {
	role       = aws_iam_role.submit.name
	policy_arn = aws_iam_policy.userdata-read.arn
}

resource "aws_iam_role_policy_attachment" "submit-userdata-write" {
	role       = aws_iam_role.submit.name
	policy_arn = aws_iam_policy.userdata-write.arn
}

resource "aws_lambda_function" "submit" {
	filename      = "../gosubmit.zip"
	function_name = "submit"
	role          = aws_iam_role.submit.arn
	handler       = "bootstrap"
	runtime = "provided.al2"
	memory_size = 128
	architectures = ["arm64"]

	source_code_hash = filebase64sha256("../gosubmit.zip")

	environment {
		variables = {
			TOKEN_URL = "https://test-api.service.hmrc.gov.uk/oauth/token"
			CLIENT_ID = var.CLIENT_ID
			CLIENT_SECRET = var.CLIENT_SECRET
			TOKEN_BUCKET = aws_s3_bucket.userdata.bucket
      DEC_SUBMITTER = "GB906263308468"
      MOV_SUBMITTER = "GB417869120000"
		}
	}
}

resource "aws_lambda_permission" "submit-s3" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.submit.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.appdata.arn
}

resource "aws_cloudwatch_log_group" "submit" {
  name              = "/aws/lambda/${aws_lambda_function.submit.function_name}"
  retention_in_days = 7
}