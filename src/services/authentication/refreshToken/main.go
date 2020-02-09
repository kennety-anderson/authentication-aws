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
	"github.com/dgrijalva/jwt-go"
)

// Request is event output lambda
type Request events.APIGatewayProxyRequest

// Response is event output lambda
type Response events.APIGatewayProxyResponse

var (
	secretKeyAccessToken  = os.Getenv("SECRET_ACCESS_TOKEN")
	secretKeyRefreshToken = os.Getenv("SECRET_REFRESH_TOKEN")
	tableName             = os.Getenv("DYNAMO_TABLE")
)

func verifyToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	res, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return res.Claims.(jwt.MapClaims), nil
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event Request) (Response, error) {
	var buf bytes.Buffer

	secretAccessToken := []byte(secretKeyAccessToken)
	secretRefreshToken := []byte(secretKeyRefreshToken)

	tokenString := event.Headers["Authorization"]

	claims, err := verifyToken(tokenString, secretRefreshToken)

	timeNow := time.Now()

	//gerando um novo accessToken
	accessJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"_id":   claims["_id"],
		"name":  claims["name"],
		"email": claims["email"],
		"exp":   timeNow.UTC().Add(3 * time.Minute).Unix(),
		"date":  timeNow,
	})

	accessToken, _ := accessJwt.SignedString(secretAccessToken)

	if err != nil {
		return Response{StatusCode: 401, Body: "Unauthorized"}, nil
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	email := fmt.Sprintf("%v", claims["email"])

	//Get Item do no dynamodb atraves do email
	data, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	})

	//pegando valor do refreshToken do dynamoDb
	item := *data.Item["refreshToken"].S

	if err != nil || item != tokenString {
		return Response{StatusCode: 401, Body: "Unauthorized"}, nil
	}

	body, err := json.Marshal(map[string]interface{}{
		"accessToken": accessToken,
	})

	if err != nil {
		return Response{StatusCode: 500, Body: "Internal server error!"}, nil
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
