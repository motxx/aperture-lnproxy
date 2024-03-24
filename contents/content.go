package content

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	pricespb "github.com/motxx/aperture-lnproxy/aperture/pricesrpc"
	pb "github.com/motxx/aperture-lnproxy/contents/contentrpc"
	"github.com/motxx/aperture-lnproxy/contents/db"
	"google.golang.org/grpc"
)

type Server struct {
	DB *db.DB

	*pb.UnimplementedContentServiceServer
	*pricespb.UnimplementedPricesServer

	contentServiceServer *grpc.Server

	// TODO: refactor
	cfg    aws.Config
	signer *v4.Signer
	cred   aws.Credentials
	ctx    context.Context
	conf   ServerConfig

	pricesServer *grpc.Server
}

type ServerConfig struct {
	Region   string `env:"AWS_REGION" envDefault:"ap-northeast-1"`
	S3Bucket string `env:"AWS_S3_BUCKET"`
}

func NewServer(ctx context.Context) (*Server, error) {
	db, err := db.NewDB()
	if err != nil {
		return nil, err
	}

	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	var conf ServerConfig
	if err := env.Parse(&conf); err != nil {
		panic(err)
	}

	s := &Server{
		DB:   db,
		conf: conf,
	}

	err = s.SetupAWSConfig(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) SetupAWSConfig(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-northeast-1"),
	)
	if err != nil {
		return err
	}

	signer := v4.NewSigner(func(signer *v4.SignerOptions) {
		signer.DisableURIPathEscaping = true
	})
	cred, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return err
	}

	s.signer = signer
	s.cred = cred
	s.cfg = cfg
	s.ctx = ctx

	return nil
}

func (s *Server) Start() error {
	// Start the Content gRPC server.
	s.contentServiceServer = grpc.NewServer()
	pb.RegisterContentServiceServer(s.contentServiceServer, s)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	log.Printf("Content Server serving at %s", ":8080")
	go func() {
		if err := s.contentServiceServer.Serve(lis); err != nil {
			fmt.Printf("error starting content service server: %v\n", err)
		}
	}()

	// Start the Content gRPC server.
	s.pricesServer = grpc.NewServer()
	pricespb.RegisterPricesServer(s.pricesServer, s)

	lis2, err := net.Listen("tcp", ":8083")
	if err != nil {
		return err
	}

	log.Printf("Prices Server serving at %s", ":8083")
	go func() {
		if err := s.pricesServer.Serve(lis2); err != nil {
			fmt.Printf("error starting content server: %v\n", err)
		}
	}()

	// Start the http server that listens for content requests.
	r := mux.NewRouter()
	r.HandleFunc("/test", freebeeHandler).Methods("GET")
	r.HandleFunc("/content/{id}", s.contentHandler).Methods("GET")

	log.Printf("Serving HTTP server on port %s", ":9000")
	go func() {
		if err := http.ListenAndServe(":9000", r); err != nil {
			fmt.Printf("error starting http server: %v\n", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.contentServiceServer.Stop()

	return s.DB.Close()
}

var _ pb.ContentServiceServer = (*Server)(nil)

func (s *Server) AddContent(ctx context.Context, req *pb.AddContentRequest) (*pb.AddContentResponse, error) {
	content := req.GetContent()
	id, err := s.DB.AddContent(&db.Content{
		Id:             content.Id,
		Title:          content.Title,
		Author:         content.Author,
		Filepath:       content.Filepath,
		RecipientLud16: content.RecipientLud16,
		Price:          content.Price,
	})
	if err != nil {
		return nil, err
	}

	return &pb.AddContentResponse{
		Id: id,
	}, nil
}

func (s *Server) UpdateContent(ctx context.Context, req *pb.UpdateContentRequest) (*pb.UpdateContentResponse, error) {
	content := req.GetContent()
	id, err := s.DB.UpdateContent(&db.Content{
		Id:             content.Id,
		Title:          content.Title,
		Author:         content.Author,
		Filepath:       content.Filepath,
		RecipientLud16: content.RecipientLud16,
		Price:          content.Price,
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateContentResponse{
		Id: id,
	}, nil
}

func (s *Server) RemoveContent(ctx context.Context, req *pb.RemoveContentRequest) (*pb.RemoveContentResponse, error) {
	id, err := s.DB.RemoveContent(req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.RemoveContentResponse{
		Id: id,
	}, nil
}

func (s *Server) GetContent(ctx context.Context, req *pb.GetContentRequest) (*pb.GetContentResponse, error) {
	content, err := s.DB.GetContent(req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.GetContentResponse{
		Content: &pb.Content{
			Id:             content.Id,
			Title:          content.Title,
			Author:         content.Author,
			Filepath:       content.Filepath,
			RecipientLud16: content.RecipientLud16,
			Price:          content.Price,
		},
	}, nil
}

func freebeeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Freebee endpoint test")
}

func (s *Server) contentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	content, err := s.DB.GetContent(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	originalURL, err := url.Parse(
		fmt.Sprintf(
			"https://%s.s3-%s.amazonaws.com/%s",
			s.conf.S3Bucket,
			s.cfg.Region,
			content.Filepath,
		),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, originalURL.String(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := ioutil.ReadAll(resp.Body)
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/plain")

	if _, err := w.Write([]byte(sEnc)); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
