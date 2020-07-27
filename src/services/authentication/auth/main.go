package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/dgrijalva/jwt-go"
	body "github.com/kennety-anderson/aws-api-gateway-packages/body"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Request is event input lambda
type Request events.APIGatewayProxyRequest

// Response is event output lambda
type Response events.APIGatewayProxyResponse

var (
	collection            = "customers"
	database              = "slsTest"
	uri                   = os.Getenv("MONGO_URI")
	secretKeyAccessToken  = os.Getenv("SECRET_ACCESS_TOKEN")
	secretKeyRefreshToken = os.Getenv("SECRET_REFRESH_TOKEN")
	tableName             = os.Getenv("DYNAMO_TABLE")
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event Request) (Response, error) {
	// var buf bytes.Buffer

	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "false",
	}

	user := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	json.Unmarshal([]byte(event.Body), &user)

	// conexão com o mongodb e busca do usuario
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	mgo := client.Database(database).Collection(collection)

	result := struct {
		ID       string `json:"_id" bson:"_id"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}{}

	err := mgo.FindOne(ctx, bson.D{{"email", user.Email}}).Decode(&result)

	if err != nil {
		return Response{StatusCode: 401, Body: body.Create(map[string]interface{}{
			"message": "Unauthorized",
		}), Headers: headers}, nil
	}

	// verificação de usuario atraves da senha encriptada
	hashPassword := []byte(result.Password)
	password := []byte(user.Password)

	err = bcrypt.CompareHashAndPassword(hashPassword, password)

	if err != nil {
		return Response{StatusCode: 401, Body: body.Create(map[string]interface{}{
			"message": "Unauthorized",
		}), Headers: headers}, nil
	}

	timeNow := time.Now()

	accessJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"_id":   result.ID,
		"name":  result.Name,
		"email": user.Email,
		"exp":   timeNow.UTC().Add(3 * time.Minute).Unix(),
		"date":  timeNow,
	})

	refreshJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"_id":   result.ID,
		"name":  result.Name,
		"email": user.Email,
		"exp":   timeNow.UTC().Add(24 * time.Hour).Unix(),
		"date":  timeNow,
	})

	secretAccessToken := []byte(secretKeyAccessToken)
	secretRefreshToken := []byte(secretKeyRefreshToken)

	accessToken, _ := accessJwt.SignedString(secretAccessToken)
	refreshToken, _ := refreshJwt.SignedString(secretRefreshToken)

	// config da sesseion
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//criação de uma nova session de conexão
	svc := dynamodb.New(sess)

	// cria um map de itens para serem adicionados a tabela
	av, _ := dynamodbattribute.MarshalMap(map[string]interface{}{
		"email":        user.Email,
		"refreshToken": refreshToken,
		"ttl":          timeNow.UTC().Add(24 * time.Hour).Unix(),
	})

	// PutItems do refreshToken no dynamodb
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	})

	// verificação se ouve algum erro ao salvar os tokens do dynamo, mesmo se
	// tiver ocorrido um erro permite ao usuario se logar retornando os tokens
	if err != nil {
		fmt.Println(err.Error())
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body: body.Create(map[string]interface{}{
			"_id":          result.ID,
			"name":         result.Name,
			"email":        user.Email,
			"accessToken":  accessToken,
			"refreshtoken": refreshToken,
		}),
		Headers: headers,
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
