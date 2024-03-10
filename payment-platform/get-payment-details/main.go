package main

import (
	"PaymentGateway/helpers"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"net/http"
)

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var request helpers.PaymentDetailsRequest

	// Validate json schema
	err := helpers.ValidateGetPaymentDetailsRequest([]byte(event.Body))
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
			StatusCode: http.StatusBadRequest,
			Body:       helpers.GetErrorResponseBody("400", "Error parsing request body"),
		}, nil
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       helpers.GetErrorResponseBody("500", err.Error()),
		}, nil
	}

	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Payments"),
		Key: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(request.TransactionId),
			},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", err.Error()),
		}, nil
	}
	if result.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", fmt.Errorf("payment details not found for TransactionId: %s", request.TransactionId).Error()),
		}, nil
	}

	var paymentDetails helpers.PaymentDetailsResponse
	if err := dynamodbattribute.UnmarshalMap(result.Item, &paymentDetails); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", fmt.Errorf("failed to unmarshal response: %s", err).Error()),
		}, nil
	}

	responseBody, err := json.Marshal(paymentDetails)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", fmt.Errorf("failed to marshal response: %s", err).Error()),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
