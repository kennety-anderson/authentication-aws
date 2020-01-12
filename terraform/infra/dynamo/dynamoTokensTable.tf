resource "aws_dynamodb_table" "customers_refresh_tokens_table" {
  name     = "${var.environment}-customersRefreshTokens"
  hash_key = "refreshToken"
  attribute {
    name = "refreshToken"
    type = "S"
  }
  write_capacity = "${var.write_capacity}"
  read_capacity  = "${var.read_capacity}"
}

resource "aws_ssm_parameter" "customers_refresh_tokens_table_name" {
  name  = "/${var.environment}/db/dynamodb/customersRefreshTokensTable"
  type  = "String"
  value = "${aws_dynamodb_table.customers_refresh_tokens_table.name}"
}
