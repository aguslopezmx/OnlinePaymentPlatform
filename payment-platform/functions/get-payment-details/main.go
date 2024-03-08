package main

import (
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
)

type PaymentDetailsResponse struct {
	TransactionID string  `json:"transactionId"`
	Amount        float64 `json:"amount"`
	CardNumber    string  `json:"cardNumber"`
	Status        string  `json:"status"`
	MerchantID    string  `json:"merchantId"`
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	transactionID := event.PathParameters["transactionId"]
	log.Printf("TransactionID: %s", transactionID)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to create session: %s", err)
	}

	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("PaymentsTable"),
		Key: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(transactionID),
			},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to get item: %s", err)
	}
	if result.Item == nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("payment details not found for TransactionID: %s", transactionID)
	}

	var paymentDetails PaymentDetailsResponse
	if err := dynamodbattribute.UnmarshalMap(result.Item, &paymentDetails); err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	responseBody, err := json.Marshal(paymentDetails)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to marshal response: %s", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
