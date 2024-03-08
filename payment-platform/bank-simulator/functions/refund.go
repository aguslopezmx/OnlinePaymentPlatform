package functions

import (
	"encoding/json"
	"log"
	"net/http"
)

type RefundRequest struct {
	TransactionID string `json:"transactionId"`
}

func Refund() {
	http.HandleFunc("/process-refund", func(w http.ResponseWriter, r *http.Request) {
		var refundRequest RefundRequest
		if err := json.NewDecoder(r.Body).Decode(&refundRequest); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		//  solo verificamos que el TransactionID no esté vacío
		if refundRequest.TransactionID == "" {
			response := PaymentResponse{
				Status:  "Failed",
				Message: "Invalid transaction ID",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// reembolso es siempre exitoso
		response := PaymentResponse{
			Status:        "Success",
			TransactionID: refundRequest.TransactionID,
			Message:       "Refund processed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Fatalf("Error encoding response: %v", err)
		}
	})
}
