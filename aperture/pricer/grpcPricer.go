package pricer

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/motxx/aperture-lnproxy/aperture/pricesrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds all the config values required to initialise the GRPCPricer.
type Config struct {
	// Enabled indicates if the grpcPricer is to be used.
	Enabled bool `long:"enabled" description:"Set to true if a gRPC server is available to query for price data"`

	// GRPCAddress is the address that the pricer gRPC server is serving on.
	GRPCAddress string `long:"grpcaddress" description:"gRPC addr to use for price info for service resources"`

	// Insecure indicates if the connection to the gRPC server should use
	// TLS encryption or not.
	Insecure bool `long:"insecure" description:"Set to true if no TLS encryption is to be used"`

	// TLSCertPath is the path the the tls cert used by the price server.
	TLSCertPath string `long:"tlscertpath" description:"Path to the servers tls cert"`
}

// GRPCPricer uses the pricesrpc PricesClient to query a backend server for
// the price of a service resource given the resource path. It implements the
// Pricer interface.
type GRPCPricer struct {
	rpcConn   *grpc.ClientConn
	rpcClient pricesrpc.PricesClient
}

// NewGRPCPricer initialises a Pricer backed by a gRPC backend server.
func NewGRPCPricer(cfg *Config) (*GRPCPricer, error) {
	var (
		c   GRPCPricer
		err error
		opt grpc.DialOption
	)

	if cfg.Insecure {
		opt = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		tlsCredentials, err := credentials.NewClientTLSFromFile(
			cfg.TLSCertPath, "",
		)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to load TLS cert %s: %v",
				cfg.TLSCertPath, err,
			)
		}
		opt = grpc.WithTransportCredentials(tlsCredentials)
	}

	c.rpcConn, err = grpc.Dial(cfg.GRPCAddress, opt)
	if err != nil {
		return nil, err
	}

	c.rpcClient = pricesrpc.NewPricesClient(c.rpcConn)

	return &c, nil
}

// GetPaymentDetails queries the server for the creator lud16 and price of a resource path
// and returns them. GetPaymentDetails is part of the Pricer interface.
func (c GRPCPricer) GetPaymentDetails(ctx context.Context,
	r *http.Request) (GetPaymentDetailsResponse, error) {

	var b bytes.Buffer
	if err := r.Write(&b); err != nil {
		return GetPaymentDetailsResponse{}, err
	}

	resp, err := c.rpcClient.GetPaymentDetails(ctx, &pricesrpc.GetPaymentDetailsRequest{
		Path:            r.URL.Path,
		HttpRequestText: b.String(),
	})
	if err != nil {
		return GetPaymentDetailsResponse{}, err
	}

	return GetPaymentDetailsResponse{
		RecipientLud16: resp.RecipientLud16,
		Price:          resp.PriceSats,
	}, nil
}

// Close closes the gRPC connection. It is part of the Pricer interface.
func (c GRPCPricer) Close() error {
	return c.rpcConn.Close()
}
