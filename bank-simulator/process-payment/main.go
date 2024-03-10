package main

import (
	"PaymentGateway/helpers"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"time"
)

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var paymentRequest helpers.PaymentRequest
	err := json.Unmarshal([]byte(event.Body), &paymentRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	// Validar el número de la tarjeta
	if len(paymentRequest.CardNumber) != 16 {
		response := helpers.PaymentResponse{
			Status:  "Failed",
			Message: "Invalid card number length",
		}
		resp, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       string(resp),
		}, nil
	}

	// Verificar que la tarjeta no ha expirado
	currentYear, currentMonth := time.Now().Year(), int(time.Now().Month())
	if paymentRequest.ExpiryYear < currentYear || (paymentRequest.ExpiryYear == currentYear && paymentRequest.ExpiryMonth < currentMonth) {
		response := helpers.PaymentResponse{
			Status:  "Failed",
			Message: "Card has expired",
		}
		resp, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       string(resp),
		}, nil
	}
	response := helpers.PaymentResponse{
		Status:        "Success",
		TransactionId: paymentRequest.TransactionId,
		Message:       "Payment processed successfully",
	}

	resp, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(resp),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
