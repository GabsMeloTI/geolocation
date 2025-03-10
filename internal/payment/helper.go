package payment

import (
	"fmt"
	"time"
)

func extractCheckoutSessionData(event map[string]interface{}) CreatePaymentHistRequest {
	dataObject, ok := event["data"].(map[string]interface{})["object"].(map[string]interface{})
	if !ok {
		return CreatePaymentHistRequest{}
	}
	invoice, _ := dataObject["invoice"].(string)

	return CreatePaymentHistRequest{
		UserID:           dataObject["client_reference_id"].(string),
		Email:            dataObject["customer_details"].(map[string]interface{})["email"].(string),
		Name:             dataObject["customer_details"].(map[string]interface{})["name"].(string),
		Value:            float64(dataObject["amount_total"].(float64)) / 100,
		Method:           dataObject["payment_method_types"].([]interface{})[0].(string),
		Automatic:        dataObject["payment_method_options"].(map[string]interface{})["card"].(map[string]interface{})["request_three_d_secure"].(string) == "automatic",
		PaymentDate:      time.Unix(int64(dataObject["created"].(float64)), 0),
		PaymentExpireted: time.Unix(int64(dataObject["expires_at"].(float64)), 0),
		PaymentStatus:    dataObject["payment_status"].(string),
		Currency:         dataObject["currency"].(string),
		Invoice:          invoice,
		Customer:         dataObject["customer"].(string),
	}
}

func extractInvoiceData(event map[string]interface{}) CreatePaymentHistRequest {
	dataObject, ok := event["data"].(map[string]interface{})["object"].(map[string]interface{})
	if !ok {
		return CreatePaymentHistRequest{}
	}

	var invoice string
	if lines, exists := dataObject["lines"].(map[string]interface{})["data"].([]interface{}); exists && len(lines) > 0 {
		if firstItem, ok := lines[0].(map[string]interface{}); ok {
			invoice, _ = firstItem["invoice"].(string)
		}
	}

	var interval string
	if lines, exists := dataObject["lines"].(map[string]interface{})["data"].([]interface{}); exists && len(lines) > 0 {
		if firstItem, ok := lines[0].(map[string]interface{}); ok {
			if plan, exists := firstItem["plan"].(map[string]interface{}); exists {
				interval, _ = plan["interval"].(string)
			}
		}
	}
	fmt.Println(interval)

	return CreatePaymentHistRequest{
		Email:    dataObject["customer_email"].(string),
		Name:     dataObject["customer_name"].(string),
		Currency: dataObject["currency"].(string),
		Invoice:  invoice,
		Customer: dataObject["customer"].(string),
		Interval: interval,
	}
}

func mergePayments(a, b CreatePaymentHistRequest) CreatePaymentHistRequest {
	if a.Email == "" {
		a.Email = b.Email
	}
	if a.Name == "" {
		a.Name = b.Name
	}
	if a.Value == 0 {
		a.Value = b.Value
	}
	if a.Method == "" {
		a.Method = b.Method
	}
	if !a.Automatic {
		a.Automatic = b.Automatic
	}
	if a.PaymentDate.IsZero() {
		a.PaymentDate = b.PaymentDate
	}
	if a.PaymentExpireted.IsZero() {
		a.PaymentExpireted = b.PaymentExpireted
	}
	if a.PaymentStatus == "" {
		a.PaymentStatus = b.PaymentStatus
	}
	if a.Currency == "" {
		a.Currency = b.Currency
	}
	if a.Interval == "" {
		a.Interval = b.Interval
	}
	return a
}
