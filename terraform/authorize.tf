data "archive_file" "authorize" {
  type        = "zip"
  source_file = "../authorize/app.mjs"
  output_path = "../authorize.zip"
}

resource "aws_lambda_function" "authorize" {
  function_name = "authorize"

  filename = data.archive_file.authorize.output_path

  runtime = "nodejs18.x"
  handler = "app.handler"
  memory_size = 128

  source_code_hash = data.archive_file.authorize.output_base64sha256

  role = aws_iam_role.authorize.arn

  environment {
    variables = {
      APPDATA_BUCKET = aws_s3_bucket.appdata.bucket
      AUTH_URL = "https://test-www.tax.service.gov.uk/oauth/authorize"
      CLIENT_ID = var.CLIENT_ID
      SCOPE = "write:customs-declaration+write:customs-inventory-linking-exports+write:customs-il-imports-movement-validation+write:customs-il-imports-arrival-notifications+write:customs-declarations-information"
      REDIRECT_URI = var.REDIRECT_URI
    }
  }
}

resource "aws_iam_role" "authorize" {
  name = "authorize"

  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "authorize-basic" {
  role       = aws_iam_role.authorize.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "authorize-appdata-write" {
  role       = aws_iam_role.authorize.name
  policy_arn = aws_iam_policy.appdata-write.arn
}

resource "aws_api_gateway_resource" "authorize" {
  parent_id   = aws_api_gateway_resource.oauth.id
  path_part   = "authorize"
  rest_api_id = aws_api_gateway_rest_api.cds.id
}

resource "aws_api_gateway_method" "authorize" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.authorize.id
  rest_api_id   = aws_api_gateway_rest_api.cds.id
}

resource "aws_api_gateway_integration" "authorize" {
  http_method = aws_api_gateway_method.authorize.http_method
  resource_id = aws_api_gateway_resource.authorize.id
  rest_api_id = aws_api_gateway_rest_api.cds.id
  type        = "AWS_PROXY"
  integration_http_method = "POST"
  uri         = aws_lambda_function.authorize.invoke_arn
}

resource "aws_lambda_permission" "authorize" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.authorize.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.cds.execution_arn}/*/GET/oauth/authorize"
}