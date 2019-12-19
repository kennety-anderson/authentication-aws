package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Response events.APIGatewayProxyResponse

var (
	dbName = "slsTest"
	uri    = aws.String(os.Getenv("MONGO_URI"))
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	var buf bytes.Buffer

	clientOpts := options.Client().ApplyURI(*uri)
	client, _ := mongo.Connect(ctx, clientOpts)

	db := client.Database(dbName)

	fmt.Println(db.Name()) // output: glottery

	body, err := json.Marshal(map[string]interface{}{
		"message": "Lambda auth in progress",
	})

	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
