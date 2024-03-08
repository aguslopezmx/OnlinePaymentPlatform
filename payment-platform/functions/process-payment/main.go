package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"net/http"
)

type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	CardNumber    string  `json:"cardNumber"`
	ExpiryMonth   int     `json:"expiryMonth"`
	ExpiryYear    int     `json:"expiryYear"`
	MerchantID    string  `json:"merchantId"`
	TransactionID string  `json:"transactionId"` // partition key for DynamoDB
}

type PaymentResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transactionId"`
	Message       string `json:"message"`
}

// URL del simulador de banco con ngrok
const BankSimulatorURL = "https://5df0-2806-230-4014-c6fc-b4fe-6664-88a1-592e.ngrok-free.app/process-payment"

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var request PaymentRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	// Llama al simulador de banco
	response, err := callBankSimulatorPay(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error calling bank simulator",
		}, nil
	}

	//Guarda en DynamoDB
	if err := savePaymentInDynamo(request); err != nil {
		log.Printf("Error saving payment details: %s", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       response.Message,
	}, nil
}

func callBankSimulatorPay(request PaymentRequest) (PaymentResponse, error) {
	client := &http.Client{}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return PaymentResponse{}, err
	}

	resp, err := client.Post(BankSimulatorURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return PaymentResponse{}, err
	}
	defer resp.Body.Close()

	var paymentResponse PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return PaymentResponse{}, err
	}

	return paymentResponse, nil
}

func savePaymentInDynamo(request PaymentRequest) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)
	item, err := dynamodbattribute.MarshalMap(request)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("PaymentsTable"),
		Item:      item,
	}

	_, err = svc.PutItem(input)
	return err
}

func main() {
	lambda.Start(HandleRequest)
}
