package main

import (
	"PaymentGateway/payment-platform/bank-simulator/functions"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// handleFunc para el pago
	functions.Payment()
	// handleFunc para el reembolso
	functions.Refund()

	fmt.Println("Bank simulator running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
