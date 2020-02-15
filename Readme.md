# Authentication service

## Deploy
O deploy do serviço é dividido em duas partes, aplicação e infraestrutura.

### Deploy service
Para subir seu serviço basta rodar este comando na raiz do projeto.
>make deploy

### Deploy infra
Para subir sua infraestrutura e recursos basta rodar este comando na raiz do projeto.
Tendo um para ambiente `dev` e um para `prod`.

> make tf-apply-dev

Ou 

> make tf-apply-prod

Para subir direto para um abiente de produção.


## Sobre

Este é um serviço de authenticação auth 2.0 para lambdas aws com golang e serverless framework,
utilizando terraform para gerenciar recusos na aws.

Este serviço esta dividido em 5 lambdas

### Auth 
Esta lambda realiza a autenticação de usuario, ela recebe um `email` e`password` do usuario para autenticação, no qual ela busca por `email` em um banco de dados mongodb (NoSql), e faz a validação do `password` encriptado no banco.
Apos o processo de validação do usuario a mesma salva os campos `email` (hash-key) e `refreshToken` no aws dynamodb.
Retornando um `accessToken` e `refreshToken` ao usuario como resposta.

> /auth - POST

Header 

```
  Content-Type: application/json
```

Body response 200 - ok

```json
{
 "accessToken": "value-accessToken",
 "refreshToken": "value-refreshToken"
}
```
### Authorizer 
A lambda Authorizer funciona como validador de permissões liberando acesso a lambda subsequente, retornando um `Unauthorized` caso não seja enviado um `accessToken` ou ele não seja valido ou um policy de permissão, liberando o acesso a lambda.

### Ping
Esta lambda serve justamente para validar o funcionamento da lambda `Authorizer` pois ele se encontra antes dela.

> /ping - GET

Header

```
  Content-Type: application/json
  Authorization: [valid-accessToken]
```

Body response 200 - ok

```json
 {
   "message": "pong"
 }
 ```

 ### Refresh-Token 
 A lambda refreshToken serve para a geração de um novo `accessToken`, devido ao baixo tempo de vida de cada accessToken, recebendo como header `Authorization` o `refreshToken` para realizar a validação do mesmo buscando direto no dynamodb.

 > /refreshtoken - POST

 Header 

 ```
  Content-Type: application/json
  Authorization: [valid-refreshToken]
```
Body response 200 - ok

```json
{
 "accessToken": "value-accessToken"
}
```
### Logout 
Realiza o logout do usuario, esta lambda possui um `Àuhorizer` na frente, sendo assim necessario o envio de um `accessToken` valido.
Apos a validação no Authorizer ele busca e deleta o `accessToken` do usuario no dynamodb, impossibilitando assim o usuario realizar uma chamada a lambda `refreshToken` para conseguir um novo `accessToken`, tendo que o usuario se logar novamente na lambda `Auth`.

> /logout - POST

Header

```
Content-Type: application/json
Authorization: [valid-accessToken]
```
Response 200 - ok

### Obs
A lambda logout não possui nenhum body de retorno somente um statusCode 200
