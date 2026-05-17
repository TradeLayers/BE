package handler

import (
	"net/http"
	"net/url"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DonationHandler struct {
	service        service.DonationService
	allowedOrigins []string
}

func NewDonationHandler(s service.DonationService, allowedOrigins []string) *DonationHandler {
	return &DonationHandler{service: s, allowedOrigins: allowedOrigins}
}

func (h *DonationHandler) CreateCheckoutSession(c *gin.Context) {
	log := requestLogger(c)

	var req model.DonationCheckoutRequest = model.DonationCheckoutRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("failed to bind donation checkout request", zap.Error(err))
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	if !h.isAllowedRedirectURL(req.SuccessURL) || !h.isAllowedRedirectURL(req.CancelURL) {
		log.Warn("blocked donation checkout redirect URL", zap.String("successUrl", req.SuccessURL), zap.String("cancelUrl", req.CancelURL))
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	response, domainErr := h.service.CreateCheckoutSession(c.Request.Context(), req)
	if domainErr != appErrors.ErrNone {
		appErrors.ReturnError(c, domainErr)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *DonationHandler) isAllowedRedirectURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	origin := parsedURL.Scheme + "://" + parsedURL.Host
	for _, allowedOrigin := range h.allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}
