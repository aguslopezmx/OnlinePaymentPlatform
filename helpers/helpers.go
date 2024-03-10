package helpers

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"log"
)

// TokenRequest is a request for a token
type TokenRequest struct {
	MerchantId string `json:"merchantId"`
	CustomerId string `json:"customerId"`
	ApiKey     string `json:"apiKey"`
}

// TokenResponse is a response for a token request
type TokenResponse struct {
	Token string `json:"token"`
}

// PaymentRequest is a request for a payment
type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	CardNumber    string  `json:"cardNumber"`
	ExpiryMonth   int     `json:"expiryMonth"`
	ExpiryYear    int     `json:"expiryYear"`
	MerchantId    string  `json:"merchantId"`
	TransactionId string  `json:"transactionId"`
	CustomerId    string  `json:"customerId"`
}

// PaymentResponse is a response for a payment request
type PaymentResponse struct {
	Status        string `json:"status"`
	TransactionId string `json:"transactionId"`
	Message       string `json:"message"`
}

// RefundRequest is a request for a refund
type RefundRequest struct {
	TransactionId string `json:"transactionId"`
	MerchantId    string `json:"merchantId"`
	CustomerId    string `json:"customerId"`
	ApiKey        string `json:"apiKey"`
}

// RefundResponse is a response for a refund request
type RefundResponse struct {
	Status        string `json:"status"`
	TransactionId string `json:"transactionId"`
	Message       string `json:"message"`
}

// PaymentDetailsRequest is a request for getting payment details
type PaymentDetailsRequest struct {
	TransactionId string `json:"transactionId"`
	MerchantId    string `json:"merchantId"`
	CustomerId    string `json:"customerId"`
	ApiKey        string `json:"apiKey"`
}

// PaymentDetailsResponse is a response for getting payment details
type PaymentDetailsResponse struct {
	TransactionId string  `json:"transactionId"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	CardNumber    string  `json:"cardNumber"`
	Message       string  `json:"message"`
	MerchantId    string  `json:"merchantId"`
	CustomerId    string  `json:"customerId"`
}

// ErrorResponse is a response for an error
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// TokenRequestSchema is a JSON schema for a token request
var TokenRequestSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "customerId": {
      "type": "string"
    },
    "merchantId": {
      "type": "string"
    },
    "apiKey": {
      "type": "string"
    }
  },
  "required": [
    "customerId",
    "merchantId",
    "apiKey"
  ]
}`

// PaymentRequestSchema is a JSON schema for a payment request
// The cardNumber pattern is a regex for a valid credit card number for Visa, Mastercard, American Express, Discover, and Diners Club
var PaymentRequestSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "amount": {
      "type": "number"
    },
    "cardNumber": {
      "type": "string",
	  "pattern": "^(?:4[0-9]{12}(?:[0-9]{3})?|[25][1-7][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})$"
    },
    "expiryMonth": {
      "type": "integer",
	  "minimum": 1,
	  "maximum": 12
    },
    "expiryYear": {
      "type": "integer",
	  "minimum": 2024,
	  "maximum": 2100
    },
    "cvv": {
      "type": "string",
	  "pattern": "^[0-9]{3,4}$"
    },
    "merchantId": {
      "type": "string"
    },
    "customerId": {
      "type": "string"
    },
    "transactionId": {
      "type": "string"
    },
    "apiKey": {
      "type": "string"
    }
  },
  "required": [
    "amount",
    "cardNumber",
    "expiryMonth",
    "expiryYear",
    "cvv",
    "merchantId",
    "customerId",
    "transactionId",
    "apiKey"
  ]
}`

// RefundRequestSchema is a JSON schema for a refund request
var RefundRequestSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
	"transactionId": {
	  "type": "string"
	},
	"merchantId": {
	  "type": "string"
	},
	"customerId": {
	  "type": "string"
	},
	"apiKey": {
	  "type": "string"
	}
  },
  "required": [
	"transactionId",
	"merchantId",
	"customerId",
	"apiKey"
  ]
}`

// GetPaymentDetailsSchema is a JSON schema for a get payment details request
var GetPaymentDetailsSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
	"transactionId": {
	  "type": "string"
	},
	"merchantId": {
	  "type": "string"
	},
	"customerId": {
	  "type": "string"
	},
	"apiKey": {
	  "type": "string"
	}
  },
  "required": [
	"transactionId",
	"merchantId",
	"customerId",
	"apiKey"
  ]
}`

// ValidateTokenRequest is a request to validate a token request
func ValidateTokenRequest(request []byte) error {
	return ValidateJsonSchemaRequest(TokenRequestSchema, request)
}

// ValidatePaymentRequest is a request to validate a payment request
func ValidatePaymentRequest(request []byte) error {
	return ValidateJsonSchemaRequest(PaymentRequestSchema, request)
}

// ValidateRefundRequest is a request to validate a refund request
func ValidateRefundRequest(request []byte) error {
	return ValidateJsonSchemaRequest(RefundRequestSchema, request)
}

// ValidateGetPaymentDetailsRequest is a request to validate a get payment details request
func ValidateGetPaymentDetailsRequest(request []byte) error {
	return ValidateJsonSchemaRequest(GetPaymentDetailsSchema, request)
}

// ValidateJsonSchemaRequest is a request to validate a JSON schema
func ValidateJsonSchemaRequest(jsonSchema string, jsonData []byte) error {
	sch, err := jsonschema.CompileString("schema.json", jsonSchema)
	if err != nil {
		log.Printf("%#v", err)
		return err
	}

	var v interface{}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		log.Println(err)
		return err
	}

	if err = sch.Validate(v); err != nil {
		log.Printf("%#v", err)
		return err
	}
	return nil
}

// GetErrorResponseBody is a request to get an error response body
func GetErrorResponseBody(status string, message string) string {
	errorResponse := ErrorResponse{
		Status:  status,
		Message: message,
	}
	responseBody, err := json.Marshal(errorResponse)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(responseBody)
}

// GetSuccessResponseBody is a request to get a success response body
func GetSuccessResponseBody(successResponse any) string {
	responseBody, err := json.Marshal(successResponse)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(responseBody)
}

// GetParameterByPath is a request to get a parameter by path
func GetParameter(parameterName string) (string, error) {
	// Get the bank simulator URL from parameters store
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Println("Error creating session: ", err)
		return "", err
	}

	parameters := ssm.New(sess)
	parameter, err := parameters.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(parameterName),
	})
	if err != nil {
		log.Println("Error getting parameter: ", err)
		return "", err
	}

	return *parameter.Parameter.Value, nil
}
