package main

import (
	"context"
	"fmt"
	"log"

	content "github.com/ellemouton/aperture-demo"
	"github.com/lightningnetwork/lnd/signal"
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
