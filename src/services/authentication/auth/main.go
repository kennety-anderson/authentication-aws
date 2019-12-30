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
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

var (
	collection            = "users"
	database              = "slsTest"
	uri                   = aws.String(os.Getenv("MONGO_URI"))
	secretKeyAccessToken  = aws.String(os.Getenv("SECRET_ACCESS_TOKEN"))
	secretKeyRefreshToken = aws.String(os.Getenv("SECRET_REFRESH_TOKEN"))
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event Request) (Response, error) {
	var body []byte
	var buf bytes.Buffer

	user := make(map[string]string)
	json.Unmarshal([]byte(event.Body), &user)

	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(*uri))
	db := client.Database(database).Collection(collection)

	var result bson.M
	err := db.FindOne(ctx, bson.D{{"email", user["email"]}}).Decode(&result)

	if err != nil {
		return Response{StatusCode: 401, Body: "Unauthorized"}, nil
	}

	hash := fmt.Sprintf("%v", result["password"])
	hashPassword := []byte(hash)
	password := []byte(user["password"])

	err = bcrypt.CompareHashAndPassword(hashPassword, password)

	if err != nil {
		return Response{StatusCode: 401, Body: "Unauthorized"}, nil
	}

	claims := &jwt.MapClaims{
		"_id":   result["_id"],
		"name":  result["name"],
		"email": result["email"],
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretAccessToken := []byte(*secretKeyAccessToken)
	secretRefreshToken := []byte(*secretKeyRefreshToken)

	accessToken, _ := token.SignedString(secretAccessToken)
	refreshToken, _ := token.SignedString(secretRefreshToken)

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
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
