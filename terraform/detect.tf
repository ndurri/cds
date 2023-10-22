resource "aws_iam_role" "detect" {
	name = "detect"
    assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "detect-basic" {
  role       = aws_iam_role.detect.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "detect-appdata-read" {
	role       = aws_iam_role.detect.name
	policy_arn = aws_iam_policy.appdata-read.arn
}

resource "aws_iam_role_policy_attachment" "detect-payloadwaitingdoctype-read" {
	role       = aws_iam_role.detect.name
	policy_arn = aws_iam_policy.payloadwaitingdoctype-read.arn
}

resource "aws_iam_role_policy_attachment" "detect-payloadwaitingsubmit-write" {
	role       = aws_iam_role.detect.name
	policy_arn = aws_iam_policy.payloadwaitingsubmit-write.arn
}

data "archive_file" "detect" {
  type        = "zip"
  source_file = "../godetect/bootstrap"
  output_path = "../godetect.zip"
}

resource "aws_lambda_function" "detect" {
	filename      = "../godetect.zip"
	function_name = "detect"
	role          = aws_iam_role.detect.arn
	handler       = "bootstrap"
	runtime = "provided.al2"
	memory_size = 128
	architectures = ["arm64"]

	source_code_hash = filebase64sha256("../godetect.zip")
  environment {
    variables = {
      BUCKET = aws_s3_bucket.appdata.bucket
      SEND_QUEUE = "https://sqs.eu-west-2.amazonaws.com/605391140887/payloadWaitingSubmit"
    }
  }
}

resource "aws_lambda_permission" "detect-s3" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.detect.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.appdata.arn
}

resource "aws_cloudwatch_log_group" "detect" {
  name              = "/aws/lambda/${aws_lambda_function.detect.function_name}"
  retention_in_days = 7
}