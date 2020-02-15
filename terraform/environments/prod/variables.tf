variable "environment" {
  description = "variavel para definição de ambiente, prod"
  default     = "prod"
}

variable "read_capacity" {
  description = "capacidade de leitura de dados na tabela do dynamo-db"
  default     = 15
}

variable "write_capacity" {
  description = "capacidade de escrita de dados na tabela do dynamo-db"
  default     = 15
}

