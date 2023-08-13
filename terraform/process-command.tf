resource "aws_iam_role" "processCommand" {
	name = "processCommand"
    assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "processCommand-basic" {
  role       = aws_iam_role.processCommand.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "processCommand-appdata-read" {
	role       = aws_iam_role.processCommand.name
	policy_arn = aws_iam_policy.appdata-read.arn
}

resource "aws_iam_role_policy_attachment" "processCommand-appdata-write" {
	role       = aws_iam_role.processCommand.name
	policy_arn = aws_iam_policy.appdata-write.arn
}

data "archive_file" "processCommand" {
  type        = "zip"
  source_file = "../gocommand/bootstrap"
  output_path = "../gocommand.zip"
}

resource "aws_lambda_function" "processCommand" {
	filename      = data.archive_file.processCommand.output_path
	function_name = "processCommand"
	role          = aws_iam_role.processCommand.arn
	handler       = "bootstrap"
	runtime = "provided.al2"

	source_code_hash = data.archive_file.processCommand.output_base64sha256

	environment {
		variables = {
      DEC_SUBMITTER = "GB906263308468"
      MOV_SUBMITTER = "GB417869120000"
    }
	}
}

resource "aws_lambda_permission" "processCommand-s3" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.processCommand.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.appdata.arn
}

resource "aws_cloudwatch_log_group" "processCommand" {
  name              = "/aws/lambda/${aws_lambda_function.processCommand.function_name}"
  retention_in_days = 7
}