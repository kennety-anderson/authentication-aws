resource "aws_dynamodb_table" "access_tokens" {
  name     = "${var.environment}_customers_access_tokens"
  hash_key = "id"
  attribute {
    name = "id"
    type = "S"
  }
  write_capacity = "${var.write_capacity}"
  read_capacity  = "${var.read_capacity}"
}
