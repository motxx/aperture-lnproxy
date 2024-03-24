package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/motxx/aperture-lnproxy/contents/contentrpc"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	app := cli.NewApp()
	app.Name = "appcli"
	app.Usage = "Control plane for your content delivery app"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rpcserver",
			Value: "localhost:8080",
			Usage: "content app daemon address host:port",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "addcontent",
			Usage:  "add a content",
			Action: addContent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "id",
				},
				cli.StringFlag{
					Name: "title",
				},
				cli.StringFlag{
					Name: "author",
				},
				cli.StringFlag{
					Name: "filepath",
				},
				cli.StringFlag{
					Name: "recipient_lud16",
				},
				cli.Int64Flag{
					Name: "price",
				},
			},
		},
		{
			Name:   "updatecontent",
			Usage:  "update the content",
			Action: updateContent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "id",
				},
				cli.StringFlag{
					Name: "title",
				},
				cli.StringFlag{
					Name: "author",
				},
				cli.StringFlag{
					Name: "filepath",
				},
				cli.StringFlag{
					Name: "recipient_lud16",
				},
				cli.Int64Flag{
					Name: "price",
				},
			},
		},
		{
			Name:   "removecontent",
			Usage:  "remove the content",
			Action: removeContent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "id",
				},
			},
		},
		{
			Name:   "getcontent",
			Usage:  "get the content",
			Action: getContent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "id",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getClient(ctx *cli.Context) (contentrpc.ContentServiceClient, func(), error) {
	rpcServer := ctx.GlobalString("rpcserver")

	conn, err := grpc.Dial(rpcServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to RPC server: %v", err)
	}

	cleanup := func() { _ = conn.Close() }

	sessionsClient := contentrpc.NewContentServiceClient(conn)
	return sessionsClient, cleanup, nil
}

func addContent(ctx *cli.Context) error {
	client, cleanup, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")
	if id == "" {
		return fmt.Errorf("must set an id for the content")
	}

	title := ctx.String("title")
	if title == "" {
		return fmt.Errorf("must set a title for the content")
	}

	author := ctx.String("author")
	if author == "" {
		return fmt.Errorf("must set a author for the content")
	}

	filepath := ctx.String("filepath")
	if filepath == "" {
		return fmt.Errorf("must set filepath for the content")
	}

	recipientLud16 := ctx.String("recipient_lud16")
	if recipientLud16 == "" {
		return fmt.Errorf("must set recipient_lud16 for the content")
	}

	price := ctx.Int64("price")
	if price < 0 {
		return fmt.Errorf("cant have a negative price")
	}

	resp, err := client.AddContent(context.Background(),
		&contentrpc.AddContentRequest{
			Content: &contentrpc.Content{
				Id:             id,
				Title:          title,
				Author:         author,
				Filepath:       filepath,
				RecipientLud16: recipientLud16,
				Price:          price,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("Success to add content. Content id: %s\n", resp.Id)
	return nil
}

func updateContent(ctx *cli.Context) error {
	client, cleanup, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")
	if id == "" {
		return fmt.Errorf("must set an id for the content")
	}

	proto, err := getContentImpl(ctx, id)
	if err != nil {
		return fmt.Errorf("no content with id %s", id)
	}

	var moreThanOneSet bool

	title := ctx.String("title")
	if title != "" {
		proto.Title = title
		moreThanOneSet = true
	}

	author := ctx.String("author")
	if author != "" {
		proto.Author = author
		moreThanOneSet = true
	}

	filepath := ctx.String("filepath")
	if filepath != "" {
		proto.Filepath = filepath
		moreThanOneSet = true
	}

	recipientLud16 := ctx.String("recipient_lud16")
	if recipientLud16 != "" {
		proto.RecipientLud16 = recipientLud16
		moreThanOneSet = true
	}

	price := ctx.Int64("price")
	if price > 0 {
		proto.Price = price
		moreThanOneSet = true
	}

	if !moreThanOneSet {
		return fmt.Errorf("must set at least one field to update")
	}

	resp, err := client.UpdateContent(context.Background(),
		&contentrpc.UpdateContentRequest{
			Content: proto,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("Success to update content. Content id: %s\n", resp.Id)
	return nil
}

func removeContent(ctx *cli.Context) error {
	client, cleanup, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")
	if id == "" {
		return fmt.Errorf("must set an id for the content")
	}

	resp, err := client.RemoveContent(context.Background(),
		&contentrpc.RemoveContentRequest{
			Id: id,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("Success to remove content. Content id: %s\n", resp.Id)
	return nil
}

func getContent(ctx *cli.Context) error {
	id := ctx.String("id")
	if id == "" {
		return fmt.Errorf("must set an id for the content")
	}

	proto, err := getContentImpl(ctx, id)
	if err != nil {
		return err
	}

	fmt.Printf("Content: %+v\n", proto)
	return nil
}

func getContentImpl(ctx *cli.Context, id string) (*contentrpc.Content, error) {
	client, cleanup, err := getClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp, err := client.GetContent(context.Background(),
		&contentrpc.GetContentRequest{
			Id: id,
		},
	)
	if err != nil {
		return nil, err
	}

	return resp.Content, nil
}
