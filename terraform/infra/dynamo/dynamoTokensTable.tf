resource "aws_dynamodb_table" "customers_access_tokens_table" {
  name     = "${var.environment}-customersAccessTokens"
  hash_key = "id"
  attribute {
    name = "id"
    type = "S"
  }
  write_capacity = "${var.write_capacity}"
  read_capacity  = "${var.read_capacity}"
}

resource "aws_ssm_parameter" "customers_access_tokens_table_name" {
  name  = "/${var.environment}/db/dynamodb/customersAccessTokensTable"
  type  = "String"
  value = "${aws_dynamodb_table.customers_access_tokens_table.name}"
}
