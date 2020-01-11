package main

import (
	"bytes"
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
	collection            = "users"
	database              = "slsTest"
	uri                   = aws.String(os.Getenv("MONGO_URI"))
	secretKeyAccessToken  = aws.String(os.Getenv("SECRET_ACCESS_TOKEN"))
	secretKeyRefreshToken = aws.String(os.Getenv("SECRET_REFRESH_TOKEN"))
	tableName             = aws.String(os.Getenv("DYNAMO_TABLE"))
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event Request) (Response, error) {
	var body []byte
	var buf bytes.Buffer

	user := make(map[string]string)
	json.Unmarshal([]byte(event.Body), &user)

	// conexão com o mongodb e busca do usuario
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(*uri))
	mgo := client.Database(database).Collection(collection)

	var result bson.M
	err := mgo.FindOne(ctx, bson.D{{"email", user["email"]}}).Decode(&result)

	if err != nil {
		return Response{StatusCode: 401, Body: "Unauthorized"}, nil
	}

	// verificação de usuario atraves da senha encriptada
	hash := fmt.Sprintf("%v", result["password"])
	hashPassword := []byte(hash)
	password := []byte(user["password"])

	err = bcrypt.CompareHashAndPassword(hashPassword, password)

	if err != nil {
		return Response{StatusCode: 401, Body: "Unauthorized"}, nil
	}

	timeNow := time.Now()

	accessJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"_id":   result["_id"],
		"name":  result["name"],
		"email": result["email"],
		"exp":   timeNow.UTC().Add(3 * time.Minute).Unix(),
		"date":  timeNow,
	})

	refreshJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"_id":  result["_id"],
		"exp":  timeNow.UTC().Add(24 * time.Hour).Unix(),
		"date": timeNow,
	})

	secretAccessToken := []byte(*secretKeyAccessToken)
	secretRefreshToken := []byte(*secretKeyRefreshToken)

	accessToken, _ := accessJwt.SignedString(secretAccessToken)
	refreshToken, _ := refreshJwt.SignedString(secretRefreshToken)

	// PutItems dos tokens de acesso no dynamodb
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	av, _ := dynamodbattribute.MarshalMap(map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(*tableName),
	})

	// verificação se ouve algum erro ao salvar os tokens do dynamo, mesmo se
	// tiver ocorrido um erro permite ao usuario se logar retornando os tokens
	if err != nil {
		fmt.Println("Erro ao salvar tokens no dynamo")
		fmt.Println(err.Error())
	}

	// reposta lambda
	body, err = json.Marshal(map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})

	if err != nil {
		return Response{StatusCode: 500, Body: "Internal Server Error"}, nil
	}

	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
