package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Request is event output lambda
type Request events.APIGatewayProxyRequest

// Response is event output lambda
type Response events.APIGatewayProxyResponse

var tableName = aws.String(os.Getenv("DYNAMO_TABLE"))

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event Request) (Response, error) {

	tokenString := event.Headers["Authorization"]

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	data, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(*tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"refreshToken": {
				S: aws.String(tokenString),
			},
		},
	})

	_, item := data.Item["refreshToken"]

	if err != nil || item == false {
		return Response{StatusCode: 401, Body: "Unauthorized"}, nil
	}

	_, err = svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(*tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"refreshToken": {
				S: aws.String(tokenString),
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return Response{StatusCode: 422, Body: "Unprocessable Entity"}, nil
	}

	if err != nil {
		return Response{StatusCode: 500, Body: "Internal server error"}, nil
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
