package service

import (
	"context"
	"strings"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/requestlog"
	"github.com/stripe/stripe-go/v82"
	checkout_session "github.com/stripe/stripe-go/v82/checkout/session"
	"go.uber.org/zap"
)

const (
	MinDonationAmountCents int64 = 100
	MaxDonationAmountCents int64 = 50000
)

type DonationService interface {
	CreateCheckoutSession(ctx context.Context, req model.DonationCheckoutRequest) (*model.DonationCheckoutResponse, appErrors.DomainError)
}

type donationService struct {
	secretKey   string
	currency    string
	productName string
}

func NewDonationService(secretKey string, currency string, productName string) DonationService {
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		currency = "eur"
	}

	productName = strings.TrimSpace(productName)
	if productName == "" {
		productName = "TradeLayers donation"
	}

	return &donationService{
		secretKey:   strings.TrimSpace(secretKey),
		currency:    currency,
		productName: productName,
	}
}

func (s *donationService) CreateCheckoutSession(ctx context.Context, req model.DonationCheckoutRequest) (*model.DonationCheckoutResponse, appErrors.DomainError) {
	if s.secretKey == "" {
		return nil, appErrors.ErrDonationNotConfigured
	}

	if req.AmountCents < MinDonationAmountCents || req.AmountCents > MaxDonationAmountCents {
		return nil, appErrors.ErrInvalidDonationAmount
	}

	stripe.Key = s.secretKey

	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		Mode:       stripe.String(stripe.CheckoutSessionModePayment),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(s.currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(s.productName),
					},
					UnitAmount: stripe.Int64(req.AmountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"source": "tradelayers-donation-page",
		},
	}
	params.Context = ctx

	session, err := checkout_session.New(params)
	if err != nil {
		requestlog.FromContext(ctx).Error("failed to create stripe checkout session", zap.Error(err))
		return nil, appErrors.ErrPaymentProviderUnavailable
	}

	if strings.TrimSpace(session.URL) == "" {
		requestlog.FromContext(ctx).Error("stripe checkout session did not include a redirect URL")
		return nil, appErrors.ErrPaymentProviderUnavailable
	}

	return &model.DonationCheckoutResponse{CheckoutURL: session.URL}, appErrors.ErrNone
}
