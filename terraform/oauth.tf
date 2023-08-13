resource "aws_api_gateway_resource" "oauth" {
  parent_id   = aws_api_gateway_rest_api.cds.root_resource_id
  path_part   = "oauth"
  rest_api_id = aws_api_gateway_rest_api.cds.id
}