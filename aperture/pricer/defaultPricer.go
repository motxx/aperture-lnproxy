package pricer

import (
	"context"
	"net/http"
)

// DefaultPricer provides the same price for any service path. It implements
// the Pricer interface.
type DefaultPricer struct {
	RecipientLud16 string
	Price          int64
}

// NewDefaultPricer initialises a new DefaultPricer provider where each resource
// for the service will have the same price.
func NewDefaultPricer(price int64) *DefaultPricer {
	return &DefaultPricer{Price: price}
}

// GetPaymentDetails returns the creator lud16 and price charged for all resources of a service.
// It is part of the Pricer interface.
func (d *DefaultPricer) GetPaymentDetails(_ context.Context,
	_ *http.Request) (GetPaymentDetailsResponse, error) {

	return GetPaymentDetailsResponse{d.RecipientLud16, d.Price}, nil
}

// Close is part of the Pricer interface. For the DefaultPricer, the method does
// nothing.
func (d *DefaultPricer) Close() error {
	return nil
}
