package main

import (
	"PaymentGateway/helpers"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var refundRequest helpers.RefundRequest
	err := json.Unmarshal([]byte(event.Body), &refundRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	// Validar el número de la tarjeta
	//  solo verificamos que el TransactionID no esté vacío
	if refundRequest.TransactionId == "" {
		response := helpers.PaymentResponse{
			Status:  "Failed",
			Message: "Invalid transaction ID",
		}
		resp, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       string(resp),
		}, nil
	}

	// reembolso es siempre exitoso
	response := helpers.PaymentResponse{
		Status:        "Success",
		TransactionId: refundRequest.TransactionId,
		Message:       "Refund processed successfully",
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
