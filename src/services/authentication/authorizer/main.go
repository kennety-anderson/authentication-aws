package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/dgrijalva/jwt-go"
)

// Request is event input lambda
type Request events.APIGatewayCustomAuthorizerRequest

// Response is event output lambda
type Response events.APIGatewayCustomAuthorizerResponse

var secretKeyAccessToken = aws.String(os.Getenv("SECRET_ACCESS_TOKEN"))

func generatePolicy(principalID, effect, resource string) Response {
	authResponse := Response{PrincipalID: principalID}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}
	// authResponse.Context = map[string]interface{}{
	// 	"stringKey":  "stringval",
	// 	"numberKey":  123,
	// 	"booleanKey": true,
	// }
	return authResponse
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event Request) (Response, error) {

	tokenString := event.AuthorizationToken
	secretAccessToken := []byte(*secretKeyAccessToken)

	if tokenString == "" {
		return Response{}, errors.New("Unauthorized")
	}

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return secretAccessToken, nil
	})

	if err != nil {
		return Response{}, errors.New("Unauthorized")
	}

	return generatePolicy("customer", "Allow", event.MethodArn), nil
}

func main() {
	lambda.Start(Handler)
}
