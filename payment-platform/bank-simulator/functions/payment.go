package functions

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// PaymentRequest es una solicitud de pago
type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	CardNumber    string  `json:"cardNumber"`
	ExpiryMonth   int     `json:"expiryMonth"`
	ExpiryYear    int     `json:"expiryYear"`
	MerchantID    string  `json:"merchantId"`
	TransactionID string  `json:"transactionId"`
}

// PaymentResponse representa la respuesta de la simulación bancaria
type PaymentResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transactionId"`
	Message       string `json:"message"`
}

func Payment() {
	http.HandleFunc("/process-payment", func(w http.ResponseWriter, r *http.Request) {
		var paymentRequest PaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		// Validar el número de la tarjeta
		if len(paymentRequest.CardNumber) != 16 {
			response := PaymentResponse{
				Status:  "Failed",
				Message: "Invalid card number length",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Verificar que la tarjeta no ha expirado
		currentYear, currentMonth := time.Now().Year(), int(time.Now().Month())
		if paymentRequest.ExpiryYear < currentYear || (paymentRequest.ExpiryYear == currentYear && paymentRequest.ExpiryMonth < currentMonth) {
			response := PaymentResponse{
				Status:  "Failed",
				Message: "Card has expired",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		response := PaymentResponse{
			Status:        "Success",
			TransactionID: paymentRequest.TransactionID,
			Message:       "Payment processed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Fatalf("Error encoding response: %v", err)
		}
	})
}
