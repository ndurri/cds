resource "aws_api_gateway_rest_api" "cds" {
  name = "cds"
  disable_execute_api_endpoint = true
}

resource "aws_api_gateway_deployment" "cds" {
  rest_api_id = aws_api_gateway_rest_api.cds.id

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [aws_api_gateway_method.authorize]
}

resource "aws_api_gateway_stage" "cds" {
  deployment_id = aws_api_gateway_deployment.cds.id
  rest_api_id   = aws_api_gateway_rest_api.cds.id
  stage_name    = "v1"
}
