resource "aws_lambda_function" "reply" {
  function_name = "reply"

  filename = "../reply.zip"

  runtime = "nodejs18.x"
  handler = "app.handler"
  memory_size = 256

  source_code_hash = filebase64sha256("../reply.zip")

  role = aws_iam_role.reply.arn

  environment {
    variables = {
      SMTP_HOST = var.SMTP_HOST
      SMTP_PORT = var.SMTP_PORT
      SMTP_USER = var.SMTP_USER
      SMTP_PASSWORD = var.SMTP_PASSWORD
      SENDER = var.MAIL_RECIPIENT
    }
  }
}

resource "aws_iam_role" "reply" {
  name = "reply"

  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "reply-basic" {
  role       = aws_iam_role.reply.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "reply-appdata-read" {
  role       = aws_iam_role.reply.name
  policy_arn = aws_iam_policy.appdata-read.arn
}

resource "aws_lambda_permission" "reply-s3" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.reply.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.appdata.arn
}

resource "aws_cloudwatch_log_group" "reply" {
  name              = "/aws/lambda/${aws_lambda_function.reply.function_name}"
  retention_in_days = 7
}