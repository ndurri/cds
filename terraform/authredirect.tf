data "archive_file" "authredirect" {
  type        = "zip"
  source_file = "../auth-redirect/app.mjs"
  output_path = "../auth-redirect.zip"
}

resource "aws_lambda_function" "authredirect" {
  function_name = "authredirect"

  filename = data.archive_file.authredirect.output_path

  runtime = "nodejs18.x"
  handler = "app.handler"
  memory_size = 256

  source_code_hash = data.archive_file.authredirect.output_base64sha256

  role = aws_iam_role.authredirect.arn

  environment {
    variables = {
      APPDATA_BUCKET = aws_s3_bucket.appdata.bucket
      USERDATA_BUCKET = aws_s3_bucket.userdata.bucket
      TOKEN_URL = "https://test-api.service.hmrc.gov.uk/oauth/token"
      CLIENT_ID = var.CLIENT_ID
      CLIENT_SECRET = var.CLIENT_SECRET
      REDIRECT_URI = var.REDIRECT_URI
    }
  }
}

resource "aws_iam_role" "authredirect" {
  name = "authredirect"

  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "authredirect-basic" {
  role       = aws_iam_role.authredirect.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "authredirect-appdata-read" {
  role       = aws_iam_role.authredirect.name
  policy_arn = aws_iam_policy.appdata-read.arn
}

resource "aws_iam_role_policy_attachment" "authredirect-appdata-write" {
  role       = aws_iam_role.authredirect.name
  policy_arn = aws_iam_policy.appdata-write.arn
}

resource "aws_iam_role_policy_attachment" "authredirect-userdata-write" {
  role       = aws_iam_role.authredirect.name
  policy_arn = aws_iam_policy.userdata-write.arn
}

resource "aws_api_gateway_resource" "authredirect" {
  parent_id   = aws_api_gateway_resource.oauth.id
  path_part   = "redirect"
  rest_api_id = aws_api_gateway_rest_api.cds.id
}

resource "aws_api_gateway_method" "authredirect" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.authredirect.id
  rest_api_id   = aws_api_gateway_rest_api.cds.id
}

resource "aws_api_gateway_integration" "authredirect" {
  http_method = aws_api_gateway_method.authredirect.http_method
  resource_id = aws_api_gateway_resource.authredirect.id
  rest_api_id = aws_api_gateway_rest_api.cds.id
  type        = "AWS_PROXY"
  integration_http_method = "POST"
  uri         = aws_lambda_function.authredirect.invoke_arn
}

resource "aws_lambda_permission" "authredirect" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.authredirect.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.cds.execution_arn}/*/GET/oauth/redirect"
}
