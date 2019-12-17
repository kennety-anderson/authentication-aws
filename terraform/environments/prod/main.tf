module "name" {
  source      = "../../infra/dynamo"
  environment = "${var.environment}"
}
