package auth_test

import (
	"context"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/motxx/aperture-lnproxy/aperture/auth"
	"github.com/motxx/aperture-lnproxy/aperture/lsat"
	"github.com/motxx/aperture-lnproxy/aperture/mint"
	"gopkg.in/macaroon.v2"
)

type mockMint struct {
}

var _ auth.Minter = (*mockMint)(nil)

func (m *mockMint) MintLSAT(_ context.Context,
	services ...lsat.Service) (*macaroon.Macaroon, string, error) {

	return nil, "", nil
}

func (m *mockMint) VerifyLSAT(_ context.Context, p *mint.VerificationParams) error {
	return nil
}

type mockChecker struct {
	err error
}

var _ auth.InvoiceChecker = (*mockChecker)(nil)

func (m *mockChecker) VerifyInvoiceStatus(lntypes.Hash,
	lnrpc.Invoice_InvoiceState, time.Duration) error {

	return m.err
}
