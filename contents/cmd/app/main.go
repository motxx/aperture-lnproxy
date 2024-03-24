package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lightningnetwork/lnd/signal"
	content "github.com/motxx/aperture-lnproxy/contents"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interceptor, err := signal.Intercept()
	if err != nil {
		log.Fatal(err)
	}

	s, err := content.NewServer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Stop()

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}

	<-interceptor.ShutdownChannel()
	fmt.Println("Received shutdown signal")

	if err := s.Stop(); err != nil {
		log.Fatal(err)
	}
}
