package main

import (
	"context"

	body "github.com/kennety-anderson/aws-api-gateway-packages/body"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Request is event input lambda
type Request events.APIGatewayProxyRequest

// Response is event output lambda
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event Request) (Response, error) {
	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "false",
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body: body.Create(map[string]interface{}{
			"message": "pong",
		}),
		Headers: headers,
	}

	return resp, nil

}

func main() {
	lambda.Start(Handler)
}
