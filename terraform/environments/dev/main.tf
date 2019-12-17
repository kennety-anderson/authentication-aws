module "dev" {
  source         = "../../infra/dynamo"
  environment    = "${var.environment}"
  read_capacity  = "${var.read_capacity}"
  write_capacity = "${var.write_capacity}"
}

