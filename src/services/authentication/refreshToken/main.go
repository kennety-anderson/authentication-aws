package main

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Request is event output lambda
type Request events.APIGatewayProxyRequest

// Response is event output lambda
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, events Request) (Response, error) {
	var body []byte
	var buf bytes.Buffer

	body, err := json.Marshal(map[string]interface{}{
		"message": "lambda refreshToken in progress",
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
