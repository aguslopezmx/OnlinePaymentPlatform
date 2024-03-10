package main

import (
	"PaymentGateway/helpers"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oklog/ulid/v2"
	"log"
	"math/rand"
	"time"
)

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var request helpers.TokenRequest

	// Validate json schema
	err := helpers.ValidateTokenRequest([]byte(event.Body))
	log.Println(err)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", err.Error()),
		}, nil
	}

	err = json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", err.Error()),
		}, nil
	}

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano())))
	if err := saveTokenInDynamo(request, id.String()); err != nil {
		log.Printf("Error saving payment details: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", "Error saving token details"),
		}, nil
	}
	response := helpers.TokenResponse{
		Token: id.String(),
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       helpers.GetSuccessResponseBody(response),
	}, nil

}

func saveTokenInDynamo(request helpers.TokenRequest, token string) error {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)

	// Add the request state to the map
	itemRequest, err := dynamodbattribute.MarshalMap(request)
	if err != nil {
		return err
	}

	// Add the transactionId to the request map
	itemRequest["transactionId"] = &dynamodb.AttributeValue{
		S: aws.String(token),
	}

	// Remove apiKey from the request map
	delete(itemRequest, "apiKey")

	// Create the input for the request
	input := &dynamodb.PutItemInput{
		TableName: aws.String("Tokens"),
		Item:      itemRequest,
	}

	// Put the item in the table
	_, err = svc.PutItem(input)
	return err
}

func main() {
	lambda.Start(HandleRequest)
}
