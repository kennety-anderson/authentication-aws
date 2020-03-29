package main

import (
	"context"
	"fmt"
	"os"

	body "github.com/kennety-anderson/aws-api-gateway-packages/body"

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

var tableName = os.Getenv("DYNAMO_TABLE")
var secretKeyAccessToken = os.Getenv("SECRET_ACCESS_TOKEN")

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
	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "false",
	}

	secretAccessToken := []byte(secretKeyAccessToken)

	tokenString := event.Headers["Authorization"]

	claims, _ := verifyToken(tokenString, secretAccessToken)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	email := fmt.Sprintf("%v", claims["email"])

	data, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	})

	_, item := data.Item["email"]

	if err != nil || item == false {
		return Response{StatusCode: 401, Body: body.Create(map[string]interface{}{
			"message": "Unauthorized",
		}), Headers: headers}, nil
	}

	_, err = svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	})

	if err != nil {
		return Response{StatusCode: 422, Body: body.Create(map[string]interface{}{
			"message": "Unprocessable Entity",
		}), Headers: headers}, nil

	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Headers:         headers,
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
