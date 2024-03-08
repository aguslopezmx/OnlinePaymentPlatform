package main

import (
	"context"
	"encoding/json"
	"github.com/oklog/ulid/v2"
	"math/rand"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func HandleRequest(ctx context.Context) (events.APIGatewayProxyResponse, error) {

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano())))

	response := TokenResponse{
		Token: id.String(),
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}

	// Devolver la respuesta
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
