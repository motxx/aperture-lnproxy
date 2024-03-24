package challenger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/lightningnetwork/lnd/channeldb/migration_01_to_11/zpay32"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/motxx/aperture-lnproxy/aperture/lnurl"
)

// LnproxyChallenger is a challenger that uses an lnproxy backend to create new LSAT
// payment challenges.
type LnproxyChallenger struct {
	client        InvoiceClient
	clientCtx     func() context.Context
	genInvoiceReq InvoiceRequestGenerator

	invoiceStates  map[lntypes.Hash]lnrpc.Invoice_InvoiceState
	invoicesMtx    *sync.Mutex
	invoicesCancel func()
	invoicesCond   *sync.Cond

	errChan chan<- error

	quit chan struct{}
	wg   sync.WaitGroup
}

// A compile time flag to ensure the LnproxyChallenger satisfies the Challenger
// interface.
var _ Challenger = (*LnproxyChallenger)(nil)

// NewLnproxyChallenger creates a new challenger that uses the given connection to
// an lnd backend to create payment challenges.
func NewLnproxyChallenger(client InvoiceClient,
	genInvoiceReq InvoiceRequestGenerator,
	ctxFunc func() context.Context,
	errChan chan<- error) (*LnproxyChallenger, error) {

	// Make sure we have a valid context function. This will be called to
	// create a new context for each call to the lnd client.
	if ctxFunc == nil {
		ctxFunc = context.Background
	}

	if genInvoiceReq == nil {
		return nil, fmt.Errorf("genInvoiceReq cannot be nil")
	}

	invoicesMtx := &sync.Mutex{}
	challenger := &LnproxyChallenger{
		client:        client,
		clientCtx:     ctxFunc,
		genInvoiceReq: genInvoiceReq,
		invoiceStates: make(map[lntypes.Hash]lnrpc.Invoice_InvoiceState),
		invoicesMtx:   invoicesMtx,
		invoicesCond:  sync.NewCond(invoicesMtx),
		quit:          make(chan struct{}),
		errChan:       errChan,
	}

	err := challenger.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start challenger: %w", err)
	}

	return challenger, nil
}

// Start starts the challenger's main work which is to keep track of all
// invoices and their states. For that the backing lnd node is queried for all
// invoices on startup and the a subscription to all subsequent invoice updates
// is created.
func (l *LnproxyChallenger) Start() error {
	// These are the default values for the subscription. In case there are
	// no invoices yet, this will instruct lnd to just send us all updates.
	// If there are existing invoices, these indices will be updated to
	// reflect the latest known invoices.
	addIndex := uint64(0)
	settleIndex := uint64(0)

	// Get a list of all existing invoices on startup and add them to our
	// cache. We need to keep track of all invoices, even quite old ones to
	// make sure tokens are valid. But to save space we only keep track of
	// an invoice's state.
	ctx := l.clientCtx()
	invoiceResp, err := l.client.ListInvoices(
		ctx, &lnrpc.ListInvoiceRequest{
			NumMaxInvoices: math.MaxUint64,
		},
	)
	if err != nil {
		return err
	}

	// Advance our indices to the latest known one so we'll only receive
	// updates for new invoices and/or newly settled invoices.
	l.invoicesMtx.Lock()
	for _, invoice := range invoiceResp.Invoices {
		// Some invoices like AMP invoices may not have a payment hash
		// populated.
		if invoice.RHash == nil {
			continue
		}

		if invoice.AddIndex > addIndex {
			addIndex = invoice.AddIndex
		}
		if invoice.SettleIndex > settleIndex {
			settleIndex = invoice.SettleIndex
		}
		hash, err := lntypes.MakeHash(invoice.RHash)
		if err != nil {
			l.invoicesMtx.Unlock()
			return fmt.Errorf("error parsing invoice hash: %v", err)
		}

		// Don't track the state of canceled or expired invoices.
		if invoiceIrrelevant(invoice) {
			continue
		}
		l.invoiceStates[hash] = invoice.State
	}
	l.invoicesMtx.Unlock()

	// We need to be able to cancel any subscription we make.
	ctxc, cancel := context.WithCancel(l.clientCtx())
	l.invoicesCancel = cancel

	subscriptionResp, err := l.client.SubscribeInvoices(
		ctxc, &lnrpc.InvoiceSubscription{
			AddIndex:    addIndex,
			SettleIndex: settleIndex,
		},
	)
	if err != nil {
		cancel()
		return err
	}

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		defer cancel()

		l.readInvoiceStream(subscriptionResp)
	}()

	return nil
}

// readInvoiceStream reads the invoice update messages sent on the stream until
// the stream is aborted or the challenger is shutting down.
func (l *LnproxyChallenger) readInvoiceStream(
	stream lnrpc.Lightning_SubscribeInvoicesClient) {

	for {
		// In case we receive the shutdown signal right after receiving
		// an update, we can exit early.
		select {
		case <-l.quit:
			return
		default:
		}

		// Wait for an update to arrive. This will block until either a
		// message receives, an error occurs or the underlying context
		// is canceled (which will also result in an error).
		invoice, err := stream.Recv()
		switch {

		case err == io.EOF:
			// The connection is shutting down, we can't continue
			// to function properly. Signal the error to the main
			// goroutine to force a shutdown/restart.
			select {
			case l.errChan <- err:
			case <-l.quit:
			default:
			}

			return

		case err != nil && strings.Contains(
			err.Error(), context.Canceled.Error(),
		):

			// The context has been canceled, we are shutting down.
			// So no need to forward the error to the main
			// goroutine.
			return

		case err != nil:
			log.Errorf("Received error from invoice subscription: "+
				"%v", err)

			// The connection is faulty, we can't continue to
			// function properly. Signal the error to the main
			// goroutine to force a shutdown/restart.
			select {
			case l.errChan <- err:
			case <-l.quit:
			default:
			}

			return

		default:
		}

		// Some invoices like AMP invoices may not have a payment hash
		// populated.
		if invoice.RHash == nil {
			continue
		}

		hash, err := lntypes.MakeHash(invoice.RHash)
		if err != nil {
			log.Errorf("Error parsing invoice hash: %v", err)
			return
		}

		l.invoicesMtx.Lock()
		if invoiceIrrelevant(invoice) {
			// Don't keep the state of canceled or expired invoices.
			delete(l.invoiceStates, hash)
		} else {
			l.invoiceStates[hash] = invoice.State
		}

		// Before releasing the lock, notify our conditions that listen
		// for updates on the invoice state.
		l.invoicesCond.Broadcast()
		l.invoicesMtx.Unlock()
	}
}

// Stop shuts down the challenger.
func (l *LnproxyChallenger) Stop() {
	l.invoicesCancel()
	close(l.quit)
	l.wg.Wait()
}

type ApertureConfig struct {
	LnproxyUrl string `env:"LNPROXY_URL"`
}

type ProxyParameters struct {
	Invoice         string  `json:"invoice"`
	RoutingMsat     *uint64 `json:"routing_msat,string"`
	Description     *string `json:"description"`
	DescriptionHash *string `json:"description_hash"`
}

func getRoutingMsat(amount_sats int64) *uint64 {
	routingMsat := uint64(amount_sats * 3 / 100)
	if routingMsat < 10_000 {
		minFeeSat := uint64(10_000)
		return &minFeeSat
	}
	return &routingMsat
}

type LnproxySpecErrorResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type LnproxySpecSuccessResponse struct {
	WrappedInvoice string `json:"proxy_invoice"`
}

// NewChallenge creates a new LSAT payment challenge, returning a payment
// request (invoice) and the corresponding payment hash.
// The price is given in satoshis.
//
// NOTE: This is part of the mint.Challenger interface.
func (l *LnproxyChallenger) NewChallenge(recipientLud16 string, price int64) (string, lntypes.Hash,
	error) {

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	creatorInvoice, err := getCreatorInvoice(recipientLud16, price)
	if err != nil {
		return "", lntypes.ZeroHash, fmt.Errorf("error getting creator invoice: %v", err)
	}

	routingMsat := getRoutingMsat(price)
	log.Infof("Price: %d, RoutingMsat: %d", price, *routingMsat)

	wrappedInvoice, err := requestWrappedInvoice(ProxyParameters{
		Invoice:     creatorInvoice,
		RoutingMsat: routingMsat,
	})
	if err != nil {
		return "", lntypes.ZeroHash, fmt.Errorf("error requesting wrapped invoice: %v", err)
	}

	paymentHash, err := extractPaymentHash(wrappedInvoice)
	if err != nil {
		return "", lntypes.ZeroHash, fmt.Errorf("error extracting payment hash: %v", err)
	}
	log.Info("Payment hash: ", paymentHash)

	return wrappedInvoice, paymentHash, nil
}

func getCreatorInvoice(lud16 string, price int64) (string, error) {
	lu, err := lnurl.NewLnurl(lud16)
	if err != nil {
		return "", fmt.Errorf("error creating lnurl: %v", err)
	}

	invoice, err := lu.GetInvoice(price)
	if err != nil {
		return "", fmt.Errorf("error getting creator invoice: %v", err)
	}
	return invoice, nil
}

func requestWrappedInvoice(p ProxyParameters) (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec parameter: %v", p)
	}

	var conf ApertureConfig
	if err := env.Parse(&conf); err != nil {
		panic(err)
	}

	u, err := url.Parse(conf.LnproxyUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse lnproxy url: %v", err)
	}
	u.Path = path.Join(u.Path, "spec")
	res, err := http.Post(u.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var rawJSON json.RawMessage
	err = json.NewDecoder(res.Body).Decode(&rawJSON)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	if bytes.Contains(rawJSON, []byte("ERROR")) {
		var errResp LnproxySpecErrorResponse
		err := json.Unmarshal(rawJSON, &errResp)
		if err != nil {
			log.Info("TODO: Only Alby response is supported. If you think the response is valid, support the response here.")
			return "", fmt.Errorf("error decoding error response: %v", err)
		}
		return "", fmt.Errorf("error response: %s", errResp.Reason)
	}

	var resp LnproxySpecSuccessResponse
	err = json.Unmarshal([]byte(rawJSON), &resp)
	if err != nil {
		return "", fmt.Errorf("error decoding success response: %v", err)
	}
	return resp.WrappedInvoice, nil
}

func extractPaymentHash(invoice string) (lntypes.Hash, error) {
	decodedInvoice, err := zpay32.Decode(invoice, &chaincfg.MainNetParams)
	if err != nil {
		return lntypes.ZeroHash, fmt.Errorf("error decoding invoice: %v", err)
	}
	paymentHash := decodedInvoice.PaymentHash
	return *paymentHash, nil
}

// VerifyInvoiceStatus checks that an invoice identified by a payment
// hash has the desired status. To make sure we don't fail while the
// invoice update is still on its way, we try several times until either
// the desired status is set or the given timeout is reached.
//
// NOTE: This is part of the auth.InvoiceChecker interface.
func (l *LnproxyChallenger) VerifyInvoiceStatus(hash lntypes.Hash,
	state lnrpc.Invoice_InvoiceState, timeout time.Duration) error {

	// Prevent the challenger to be shut down while we're still waiting for
	// status updates.
	l.wg.Add(1)
	defer l.wg.Done()

	var (
		condWg         sync.WaitGroup
		doneChan       = make(chan struct{})
		timeoutReached bool
		hasInvoice     bool
		invoiceState   lnrpc.Invoice_InvoiceState
	)

	// First of all, spawn a goroutine that will signal us on timeout.
	// Otherwise if a client subscribes to an update on an invoice that
	// never arrives, and there is no other activity, it would block
	// forever in the condition.
	condWg.Add(1)
	go func() {
		defer condWg.Done()

		select {
		case <-doneChan:
		case <-time.After(timeout):
		case <-l.quit:
		}

		l.invoicesCond.L.Lock()
		timeoutReached = true
		l.invoicesCond.Broadcast()
		l.invoicesCond.L.Unlock()
	}()

	// Now create the main goroutine that blocks until an update is received
	// on the condition.
	condWg.Add(1)
	go func() {
		defer condWg.Done()
		l.invoicesCond.L.Lock()

		// Block here until our condition is met or the allowed time is
		// up. The Wait() will return whenever a signal is broadcast.
		invoiceState, hasInvoice = l.invoiceStates[hash]
		for !(hasInvoice && invoiceState == state) && !timeoutReached {
			l.invoicesCond.Wait()

			// The Wait() above has re-acquired the lock so we can
			// safely access the states map.
			invoiceState, hasInvoice = l.invoiceStates[hash]
		}

		// We're now done.
		l.invoicesCond.L.Unlock()
		close(doneChan)
	}()

	// Wait until we're either done or timed out.
	condWg.Wait()

	// Interpret the result so we can return a more descriptive error than
	// just "failed".
	switch {
	case !hasInvoice:
		return fmt.Errorf("no active or settled invoice found for "+
			"hash=%v", hash)

	case invoiceState != state:
		return fmt.Errorf("invoice status not correct before timeout, "+
			"hash=%v, status=%v", hash, invoiceState)

	default:
		return nil
	}
}
