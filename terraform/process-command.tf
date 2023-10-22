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

resource "aws_iam_role_policy_attachment" "processCommand-commandswaiting-read" {
	role       = aws_iam_role.processCommand.name
	policy_arn = aws_iam_policy.commandswaiting-read.arn
}

resource "aws_iam_role_policy_attachment" "processCommand-payloadwaitingdoctype-write" {
  role       = aws_iam_role.processCommand.name
  policy_arn = aws_iam_policy.payloadwaitingdoctype-write.arn
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
  memory_size = 128
  architectures = ["arm64"]

	source_code_hash = data.archive_file.processCommand.output_base64sha256

	environment {
		variables = {
      BUCKET = aws_s3_bucket.appdata.bucket
      SEND_QUEUE = "https://sqs.eu-west-2.amazonaws.com/605391140887/payloadWaitingDoctype"
    }
	}
}

resource "aws_cloudwatch_log_group" "processCommand" {
  name              = "/aws/lambda/${aws_lambda_function.processCommand.function_name}"
  retention_in_days = 7
}