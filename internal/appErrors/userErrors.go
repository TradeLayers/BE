package appErrors

import "github.com/gin-gonic/gin"

type DomainError int

const (
	ErrNone DomainError = iota
	ErrUserNotFound
	ErrEmptyProvidedFields
	ErrInvalidFieldInformation
	ErrInternal
	ErrNoAuthenticationHeader
	ErrJwtExpired
	ErrStockNotFound
	ErrFinnhubUnavailable
	ErrInvalidSymbol
	ErrInvalidQuantity
	ErrInvalidThreshold
	ErrInsufficientBalance
	ErrInsufficientHoldings
	ErrAlreadyWatched
	ErrNotWatched
	ErrHistoricalDataUnavailable
	ErrAlertNotFound
)

type ErrorResponse struct {
	StatusCode   int
	ErrorMessage string
}

var clientErrors = map[DomainError]ErrorResponse{
	ErrUserNotFound:              {404, "User is not found"},
	ErrEmptyProvidedFields:       {400, "Please provide at least one field to update."},
	ErrInvalidFieldInformation:   {400, "Information provided is invalid."},
	ErrInternal:                  {500, "Something went wrong. Please try again later."},
	ErrNoAuthenticationHeader:    {401, "Please sign in to continue."},
	ErrJwtExpired:                {401, "Your session has expired. Please sign in again."},
	ErrStockNotFound:             {404, "Stock not found."},
	ErrFinnhubUnavailable:        {502, "Market data service is temporarily unavailable."},
	ErrInvalidSymbol:             {400, "Invalid stock symbol."},
	ErrInvalidQuantity:           {400, "Quantity must be greater than zero."},
	ErrInvalidThreshold:          {400, "Threshold price must be greater than zero."},
	ErrInsufficientBalance:       {400, "Insufficient balance for this purchase."},
	ErrInsufficientHoldings:      {400, "You do not own enough shares to sell."},
	ErrAlreadyWatched:            {409, "Stock is already in your watchlist."},
	ErrNotWatched:                {404, "Stock is not in your watchlist."},
	ErrHistoricalDataUnavailable: {503, "Historical market data is currently unavailable."},
	ErrAlertNotFound:             {404, "Alert not found."},
}

func ReturnError(c *gin.Context, domainErr DomainError) {
	if domainErr == ErrNone {
		return
	}

	errorResponse, ok := clientErrors[domainErr]
	if !ok {
		errorResponse = clientErrors[ErrInternal]
	}

	c.AbortWithStatusJSON(errorResponse.StatusCode, gin.H{
		"error": errorResponse.ErrorMessage,
	})
}
