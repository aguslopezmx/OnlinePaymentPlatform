package main

import (
	"PaymentGateway/helpers"
	"bytes"
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

// BankSimulatorURL is the URL of the bank simulator
var BankSimulatorURL string

func init() {
	// Get the bank simulator URL from parameters store
	var err error
	BankSimulatorURL, err = helpers.GetParameter("/onlinePaymentPlatform/bankPaymentURL")
	if err != nil {
		log.Println("Error getting parameter: ", err)
		return
	}

}
func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var request helpers.PaymentRequest

	// Validate json schema
	err := helpers.ValidatePaymentRequest([]byte(event.Body))
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

	// Validate token (transactionId) if exist on Token table and not exist on Payments table
	err = validatePaymentRequest(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", err.Error()),
		}, nil
	}

	// Call the bank simulator
	response, err := callBankSimulatorPay(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       helpers.GetErrorResponseBody("500", "Error calling bank simulator"),
		}, nil
	}

	responseBody := helpers.PaymentResponse{
		Status:        "200",
		TransactionId: request.TransactionId,
		Message:       response.Message,
	}

	switch response.Status {
	case "Failed":
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", response.Message),
		}, nil
	case "Success":
		if err := savePaymentInDynamo(request, response.Status); err != nil {
			log.Printf("Error saving payment details: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       helpers.GetErrorResponseBody("400", err.Error()),
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       helpers.GetSuccessResponseBody(responseBody),
		}, nil
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       helpers.GetErrorResponseBody("500", "Internal error - Bank Simulator unhandled error"),
		}, nil
	}
}

// callBankSimulatorPay calls the bank simulator
func callBankSimulatorPay(request helpers.PaymentRequest) (helpers.PaymentResponse, error) {
	client := &http.Client{}
	requestBody, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return helpers.PaymentResponse{}, err

	}

	resp, err := client.Post(BankSimulatorURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err)
		return helpers.PaymentResponse{}, err

	}
	defer resp.Body.Close()

	var paymentResponse helpers.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		fmt.Println(err)
		return helpers.PaymentResponse{}, err
	}

	return paymentResponse, nil
}

func savePaymentInDynamo(request helpers.PaymentRequest, responseStatus string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)

	// Create a map to hold the request state
	itemRequest, err := dynamodbattribute.MarshalMap(request)
	if err != nil {
		return err
	}

	// Add the status to the response map
	itemResponse := map[string]*dynamodb.AttributeValue{
		"status": {
			S: aws.String(responseStatus),
		},
	}

	// Combine the request and response maps
	for key, value := range itemResponse {
		itemRequest[key] = value
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Payments"),
		Item:      itemRequest,
	}

	_, err = svc.PutItem(input)
	return err
}

// validatePaymentRequest is a function for validation of unique transactionId, merchantId, customerId
func validatePaymentRequest(paymentRequest helpers.PaymentRequest) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)

	// Get from DynamoDB from Token table where transactionId = paymentRequest.transactionId
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Tokens"),
		Key: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(paymentRequest.TransactionId),
			},
		},
	})

	if err != nil {
		return err
	}
	if result.Item == nil {
		return fmt.Errorf("TransactionId not found for token, please create a new token. TransactionId : %s", paymentRequest.TransactionId)
	}

	// Get from DynamoDB from Token table where transactionId = paymentRequest.transactionId
	resultPayments, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Payments"),
		Key: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(paymentRequest.TransactionId),
			},
		},
	})

	if err != nil {
		return err
	}
	if resultPayments.Item != nil {
		return fmt.Errorf("TransactionId exists, please create a new token. TransactionId : %s", paymentRequest.TransactionId)
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
