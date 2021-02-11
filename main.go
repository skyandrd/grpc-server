package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"google.golang.org/grpc"

	"github.com/skyandrd/grpc-server/internal/config"
	"github.com/skyandrd/grpc-server/internal/csvreader"
	"github.com/skyandrd/grpc-server/internal/mongo"
	pb "github.com/skyandrd/grpc-server/internal/service"
)

var (
	// Version of the app.
	Version     = "0.0.1" // nolint
	mongoClient mongo.FactoryInterface
)

func main() {
	conf, err := config.GetConfig()
	if err != nil {
		fmt.Printf("cannot read service config - %v", err)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	mongoClient = mongo.New(conf.MongoDb, conf.MongoDbURI, conf.MongoPriceCollection, conf.MongoClientConnectionTimeout)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	priceListSrv := newServer()
	pb.RegisterPriceListServer(grpcServer, priceListSrv)
	log.Printf("grpc srver is listening on port:%d\n", conf.Port)
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, []os.Signal{syscall.SIGINT, syscall.SIGTERM}...)

	// Block until we receive our signal.
	<-c

	mongoClient.Disconnect()
	log.Printf("shutting down")
	os.Exit(0)
}

func newServer() pb.PriceListServer {
	s := &priceListServer{pb.UnimplementedPriceListServer{}}

	return s
}

type priceListServer struct {
	pb.UnimplementedPriceListServer
}

//Fetch ...
func (s *priceListServer) Fetch(ctx context.Context, u *pb.URL) (response *pb.Response, err error) {
	if u.Url == "http://localhost" {
		records, err := csvreader.Read("./internal/service/mockdata/test.csv")
		if err != nil {
			log.Fatalf("csv file reading error: %v", err)
		}

		log.Println(records)

		for _, r := range records {
			log.Printf("record=%v", r)

			productName := r[0]
			price, err := strconv.ParseFloat(strings.TrimSpace(r[1]), 64)
			if err != nil {
				log.Printf("can't convert string %v to float64", r[1])

				continue
			}

			id, err := mongoClient.InsertProduct(productName, price)
			if err != nil {
				log.Fatalf("mongodb insertion error: %v", err)
			}

			log.Printf("inserted new product with id=%v", id)

		}
		return &pb.Response{Status: uint32(200)}, nil
	}

	return nil, errors.New("Not found")
}
