package model

type DonationCheckoutRequest struct {
	AmountCents int64  `json:"amountCents" binding:"required"`
	SuccessURL  string `json:"successUrl" binding:"required"`
	CancelURL   string `json:"cancelUrl" binding:"required"`
}

type DonationCheckoutResponse struct {
	CheckoutURL string `json:"checkoutUrl"`
}
