package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/motxx/aperture-lnproxy/aperture/lsat"
	"github.com/motxx/aperture-lnproxy/aperture/mint"
)

// LsatAuthenticator is an authenticator that uses the L402 protocol to
// authenticate requests.
type LsatAuthenticator struct {
	minter  Minter
	checker InvoiceChecker
}

// A compile time flag to ensure the LsatAuthenticator satisfies the
// Authenticator interface.
var _ Authenticator = (*LsatAuthenticator)(nil)

const L402RightExpiryDuration = time.Hour

// NewLsatAuthenticator creates a new authenticator that authenticates requests
// based on L402 tokens.
func NewLsatAuthenticator(minter Minter,
	checker InvoiceChecker) *LsatAuthenticator {

	return &LsatAuthenticator{
		minter:  minter,
		checker: checker,
	}
}

// Accept returns whether or not the header successfully authenticates the user
// to a given backend service.
//
// NOTE: This is part of the Authenticator interface.
func (l *LsatAuthenticator) Accept(header *http.Header, serviceName string) bool {
	// Try reading the macaroon and preimage from the HTTP header. This can
	// be in different header fields depending on the implementation and/or
	// protocol.
	mac, preimage, err := lsat.FromHeader(header)
	if err != nil {
		log.Debugf("Deny: %v", err)
		return false
	}

	verificationParams := &mint.VerificationParams{
		Macaroon:      mac,
		Preimage:      preimage,
		TargetService: serviceName,
	}
	err = l.minter.VerifyL402(context.Background(), verificationParams)
	if err != nil {
		log.Debugf("Deny: L402 settlement validation failed: %v", err)
		return false
	}

	// Make sure the backend has the invoice recorded as settled.
	err = l.checker.VerifyInvoiceStatus(
		preimage.Hash(), lnrpc.Invoice_SETTLED,
		DefaultInvoiceLookupTimeout,
	)
	if err != nil {
		log.Debugf("Deny: Invoice status mismatch: %v", err)
		return false
	}

	// Make sure the rights are still valid.
	err = l.checker.VerifyRightsWithinExpiry(
		preimage.Hash(), L402RightExpiryDuration,
	)
	if err != nil {
		log.Debugf("Deny: L402 right validation failed: %v", err)
		return false
	}

	return true
}

// FreshChallengeHeader returns a header containing a challenge for the user to
// complete.
//
// NOTE: This is part of the Authenticator interface.
func (l *LsatAuthenticator) FreshChallengeHeader(r *http.Request,
	serviceName string, serviceRecipientLud16 string, servicePrice int64) (http.Header, error) {

	service := lsat.Service{
		Name:           serviceName,
		Tier:           lsat.BaseTier,
		RecipientLud16: serviceRecipientLud16,
		Price:          servicePrice,
	}
	mac, paymentRequest, err := l.minter.MintL402(
		context.Background(), service,
	)
	if err != nil {
		log.Errorf("Error minting L402: %v", err)
		return nil, err
	}
	macBytes, err := mac.MarshalBinary()
	if err != nil {
		log.Errorf("Error serializing L402: %v", err)
	}

	str := fmt.Sprintf("L402 macaroon=\"%s\", invoice=\"%s\"",
		base64.StdEncoding.EncodeToString(macBytes), paymentRequest)
	header := r.Header
	header.Set("WWW-Authenticate", str)

	log.Debugf("Created new challenge header: [%s]", str)
	return header, nil
}
