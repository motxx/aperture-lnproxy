package pricer

import (
	"context"
	"net/http"
)

type GetPaymentDetailsResponse struct {
	RecipientLud16 string
	Price          int64
}

// Pricer is an interface used to query price data from a price provider.
type Pricer interface {
	// GetPaymentDetails should return the creator's lud16 and price in satoshis for the given
	// resource path.
	GetPaymentDetails(ctx context.Context, req *http.Request) (GetPaymentDetailsResponse, error)

	// Close should clean up the Pricer implementation if needed.
	Close() error
}
