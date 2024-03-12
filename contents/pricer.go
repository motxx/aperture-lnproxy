package content

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lightninglabs/aperture/pricesrpc"
)

var _ pricesrpc.PricesServer = (*Server)(nil)

func (s *Server) GetPrice(ctx context.Context,
	req *pricesrpc.GetPriceRequest) (*pricesrpc.GetPriceResponse, error) {

	if !strings.Contains(req.Path, "content") {
		return nil, fmt.Errorf("no prices " +
			"for given path")
	}

	id := strings.TrimLeft(req.Path, "/content/")
	log.Printf("Received request for quote with id: %s", id)

	c, err := s.DB.GetContent(id)
	if err != nil {
		return nil, err
	}

	return &pricesrpc.GetPriceResponse{
		PriceSats: c.Price,
	}, nil
}
