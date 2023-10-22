resource "aws_iam_role" "processMail" {
  name               = "processMail"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "processMail-basic" {
  role       = aws_iam_role.processMail.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "processMail-appdata-read" {
  role       = aws_iam_role.processMail.name
  policy_arn = aws_iam_policy.appdata-read.arn
}

resource "aws_iam_role_policy_attachment" "processMail-appdata-write" {
  role       = aws_iam_role.processMail.name
  policy_arn = aws_iam_policy.appdata-write.arn
}

resource "aws_iam_role_policy_attachment" "processMail-commandswaiting-write" {
  role       = aws_iam_role.processMail.name
  policy_arn = aws_iam_policy.commandswaiting-write.arn
}

resource "aws_iam_role_policy_attachment" "detect-payloadwaitingdoctype-write" {
  role       = aws_iam_role.detect.name
  policy_arn = aws_iam_policy.payloadwaitingdoctype-write.arn
}

data "archive_file" "processMail" {
  type        = "zip"
  source_file = "../goparser/bootstrap"
  output_path = "../goparser.zip"
}

resource "aws_lambda_function" "processMail" {
  filename      = data.archive_file.processMail.output_path
  function_name = "processMail"
  role          = aws_iam_role.processMail.arn
  handler       = "bootstrap"
  runtime = "provided.al2"
  memory_size = 128
  architectures = ["arm64"]
  environment {
    variables = {
      COMMAND_QUEUE = "https://sqs.eu-west-2.amazonaws.com/605391140887/commandsWaiting"
      PAYLOAD_QUEUE = "https://sqs.eu-west-2.amazonaws.com/605391140887/payloadWaitingDoctype"
    }
  }
  source_code_hash = data.archive_file.processMail.output_base64sha256
}

resource "aws_lambda_permission" "processMail-s3" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.processMail.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.appdata.arn
}

resource "aws_cloudwatch_log_group" "processMail" {
  name              = "/aws/lambda/${aws_lambda_function.processMail.function_name}"
  retention_in_days = 7
}