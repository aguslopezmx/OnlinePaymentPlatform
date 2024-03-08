package main

//
//import (
//	"context"
//	"reflect"
//	"testing"
//	"time"
//)
//
//func TestHandleRequest(t *testing.T) {
//	type Args struct {
//		Ctx     context.Context
//		Request PaymentRequest
//	}
//	testMont := int(time.Now().Month())
//	Test_args := Args{
//		Ctx: context.Background(),
//		Request: PaymentRequest{
//			Amount:        100,
//			CardNumber:    "1234567890123456",
//			ExpiryMonth:   testMont,
//			ExpiryYear:    time.Now().Year(),
//			MerchantID:    "miMerchantId",
//			TransactionID: "miTsId",
//		},
//	}
//	// test case for expired card
//	//Test_args.Request.ExpiryYear = Test_args.Request.ExpiryYear + 1
//	// test case for invalid card number
//	//Test_args.Request.CardNumber = "123456789012345"
//	// test case for invalid payment amount
//	//Test_args.Request.Amount = 0
//	tests := []struct {
//		name    string
//		args    Args
//		want    PaymentResponse
//		wantErr bool
//	}{
//		// Success test case.
//		{name: "Success", args: Test_args, want: PaymentResponse{
//			Status:        "Success",
//			TransactionID: "miTsId",
//			Message:       "Payment processed successfully",
//		}, wantErr: false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := HandleRequest(tt.args.Ctx, tt.args.Request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("HandleRequest() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("HandleRequest() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
