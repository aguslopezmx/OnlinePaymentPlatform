package main

import (
	"PaymentGateway/payment-platform/bank-simulator/functions"
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"log"
	"net/http"
)

const BankSimulatorURL = "https://a596-2806-230-4014-c6fc-b4fe-6664-88a1-592e.ngrok-free.app/process-refund"

func HandleRefund(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var request functions.RefundRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Error parsing request body",
		}, nil
	}

	response, err := callBankSimulatorRefund(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error calling bank simulator",
		}, err
	}

	if response.Status == "Failed" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       response.Message,
		}, nil
	}

	if err := updateRefundInDynamo(request); err != nil {
		log.Printf("Error updating refund details: %s", err)
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error marshaling response",
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}

func callBankSimulatorRefund(request functions.RefundRequest) (functions.PaymentResponse, error) {
	client := &http.Client{}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return functions.PaymentResponse{}, err
	}

	resp, err := client.Post(BankSimulatorURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return functions.PaymentResponse{}, err
	}
	defer resp.Body.Close()

	var paymentResponse functions.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return functions.PaymentResponse{}, err
	}

	return paymentResponse, nil
}

func updateRefundInDynamo(request functions.RefundRequest) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)

	update := expression.Set(expression.Name("status"), expression.Value("Refunded"))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("PaymentsTable"),
		Key: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(request.TransactionID),
			},
		},
		ExpressionAttributeNames:  expr.Names(),              // Añade los nombres de los atributos de la expresión
		ExpressionAttributeValues: expr.Values(),             // Añade los valores de los atributos de la expresión
		UpdateExpression:          expr.Update(),             // Añade la expresión de actualización
		ReturnValues:              aws.String("UPDATED_NEW"), // Devuelve los nuevos valores actualizados
	}

	_, err = svc.UpdateItem(input)
	return err
}

func main() {
	lambda.Start(HandleRefund)
}
