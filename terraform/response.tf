data "archive_file" "response" {
  type        = "zip"
  source_file = "../response/app.mjs"
  output_path = "../response.zip"
}

resource "aws_lambda_function" "response" {
  function_name = "response"

  filename = data.archive_file.response.output_path

  runtime = "nodejs18.x"
  handler = "app.handler"
  memory_size = 256

  source_code_hash = data.archive_file.response.output_base64sha256

  role = aws_iam_role.response.arn

  environment {
    variables = {
      BUCKET = aws_s3_bucket.appdata.bucket
    }
  }
}

resource "aws_iam_role" "response" {
  name = "response"

  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "response-basic" {
  role       = aws_iam_role.response.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "response-appdata-write" {
  role       = aws_iam_role.response.name
  policy_arn = aws_iam_policy.appdata-write.arn
}

resource "aws_api_gateway_resource" "response" {
  parent_id   = aws_api_gateway_rest_api.cds.root_resource_id
  path_part   = "hmrc"
  rest_api_id = aws_api_gateway_rest_api.cds.id
}

resource "aws_api_gateway_method" "response" {
  authorization = "NONE"
  http_method   = "POST"
  resource_id   = aws_api_gateway_resource.response.id
  rest_api_id   = aws_api_gateway_rest_api.cds.id
}

resource "aws_api_gateway_integration" "response" {
  http_method = aws_api_gateway_method.response.http_method
  resource_id = aws_api_gateway_resource.response.id
  rest_api_id = aws_api_gateway_rest_api.cds.id
  type        = "AWS_PROXY"
  integration_http_method = "POST"
  uri         = aws_lambda_function.response.invoke_arn
}

resource "aws_lambda_permission" "response" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.response.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.cds.execution_arn}/*/POST/hmrc"
}

resource "aws_cloudwatch_log_group" "response" {
  name              = "/aws/lambda/${aws_lambda_function.response.function_name}"
  retention_in_days = 7
}