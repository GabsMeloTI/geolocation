package payment

import (
	"context"
	"fmt"
	db "geolocation/db/sqlc"
	"geolocation/infra/token"
	"log"
	"strconv"
	"sync"
)

var paymentCache = sync.Map{}

type InterfaceService interface {
	ProcessStripeEvent(ctx context.Context, eventType string, event map[string]interface{}) (PaymentHistResponse, error)
	GetPaymentHistService(ctx context.Context, id int64) ([]PaymentHistResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
	maker            token.Maker
}

func NewPaymentService(InterfaceService InterfaceRepository, maker token.Maker) *Service {
	return &Service{InterfaceService, maker}
}

func (p *Service) ProcessStripeEvent(ctx context.Context, eventType string, event map[string]interface{}) (PaymentHistResponse, error) {
	log.Println("ProcessStripeEvent")
	var payment CreatePaymentHistRequest
	var userID int64

	switch eventType {
	case "checkout.session.completed":
		log.Println("corpo:", payment)
		log.Println("payment.UserID:", payment.UserID)
		payment = extractCheckoutSessionData(event)

		decryptedUserID, err := p.maker.VerifyTokenUserID(payment.UserID)
		fmt.Println("decryptedUserID:", decryptedUserID)
		if err != nil {
			return PaymentHistResponse{}, err
		}
		userID = decryptedUserID.UserID

		fmt.Println(decryptedUserID)
		fmt.Println(decryptedUserID.UserID)
	case "invoice.payment_succeeded":
		payment = extractInvoiceData(event)
	default:
		return PaymentHistResponse{}, nil
	}

	cacheKey := payment.Invoice + ":" + payment.Customer
	existingPayment, exists := paymentCache.Load(cacheKey)
	if exists {
		existing := existingPayment.(CreatePaymentHistRequest)
		mergedPayment := mergePayments(existing, payment)

		data := db.CreatePaymentHistParams{
			UserID:           userID,
			Email:            mergedPayment.Email,
			Name:             mergedPayment.Name,
			Value:            mergedPayment.Value,
			Method:           mergedPayment.Method,
			Automatic:        mergedPayment.Automatic,
			PaymentDate:      mergedPayment.PaymentDate,
			PaymentExpireted: mergedPayment.PaymentExpireted,
			PaymentStatus:    mergedPayment.PaymentStatus,
			Currency:         mergedPayment.Currency,
			Invoice:          mergedPayment.Invoice,
			Customer:         mergedPayment.Customer,
			Interval:         mergedPayment.Interval,
		}
		result, err := p.InterfaceService.CreatePaymentHist(ctx, data)
		if err != nil {
			return PaymentHistResponse{}, err
		}

		paymentCache.Delete(cacheKey)

		response := PaymentHistResponse{}
		response.ParseFromPaymentHistObject(result)
		return response, nil
	}

	payment.UserID = strconv.FormatInt(userID, 10)
	paymentCache.Store(cacheKey, payment)

	return PaymentHistResponse{}, nil
}

func (p *Service) GetPaymentHistService(ctx context.Context, id int64) ([]PaymentHistResponse, error) {
	result, err := p.InterfaceService.GetPaymentHist(ctx, id)
	if err != nil {
		return []PaymentHistResponse{}, err
	}

	var getAllPaymentHist []PaymentHistResponse
	for _, paymentHist := range result {
		getPaymentHistResponse := PaymentHistResponse{}
		getPaymentHistResponse.ParseFromPaymentHistObject(paymentHist)
		getAllPaymentHist = append(getAllPaymentHist, getPaymentHistResponse)
	}

	return getAllPaymentHist, nil
}
