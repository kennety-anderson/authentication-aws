variable "environment" {
  description = "variavel para definição de ambiente, dev"
  default     = "dev"
}

variable "read_capacity" {
  description = "capacidade de leitura de dados na tabela do dynamo-db"
  default     = 2
}

variable "write_capacity" {
  description = "capacidade de escrita de dados na tabela do dynamo-db"
  default     = 2
}

