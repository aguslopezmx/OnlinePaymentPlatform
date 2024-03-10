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
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"log"
	"net/http"
)

var BankSimulatorURL string

func init() {
	// Get the bank simulator URL from parameters store
	var err error
	BankSimulatorURL, err = helpers.GetParameter("/onlinePaymentPlatform/bankRefundURL")
	if err != nil {
		log.Println("Error getting parameter: ", err)
		return
	}
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var request helpers.RefundRequest

	// Validate json schema
	err := helpers.ValidateRefundRequest([]byte(event.Body))
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

	// Validate token (transactionId) if exist on Token table and not exist on Payments table
	err = validateRefundRequest(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       helpers.GetErrorResponseBody("400", err.Error()),
		}, nil
	}

	response, err := callBankSimulatorRefund(request)
	log.Println("Response from bank simulator:", response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       helpers.GetErrorResponseBody("500", "Error calling bank simulator"),
		}, err
	}

	responseBody := helpers.RefundResponse{
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
		if err := updateRefundInDynamo(request); err != nil {
			log.Printf("Error updating refund details: %s", err)
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

func callBankSimulatorRefund(request helpers.RefundRequest) (helpers.PaymentResponse, error) {
	client := &http.Client{}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return helpers.PaymentResponse{}, err
	}

	resp, err := client.Post(BankSimulatorURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return helpers.PaymentResponse{}, err
	}
	defer resp.Body.Close()

	var paymentResponse helpers.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return helpers.PaymentResponse{}, err
	}

	return paymentResponse, nil
}

func updateRefundInDynamo(request helpers.RefundRequest) error {
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
		TableName: aws.String("Payments"),
		Key: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(request.TransactionId),
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

// validateRefundRequest is a function for validation of unique transactionId, merchantId, customerId
func validateRefundRequest(refundRequest helpers.RefundRequest) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)

	// Get from DynamoDB from Token table where transactionId = paymentRequest.transactionId
	resultRefund, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Payments"),
		Key: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(refundRequest.TransactionId),
			},
		},
	})

	if err != nil {
		return err
	}
	if resultRefund.Item == nil {
		return fmt.Errorf("TransactionId not exists. TransactionId : %s", refundRequest.TransactionId)
	}

	if *resultRefund.Item["status"].S == "Refunded" {
		return fmt.Errorf("TransactionId was refunded. TransactionId : %s", refundRequest.TransactionId)
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
