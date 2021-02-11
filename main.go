package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/skyandrd/grpc-server/internal/config"
	"github.com/skyandrd/grpc-server/internal/mongo"
	"github.com/skyandrd/grpc-server/internal/service"
	pb "github.com/skyandrd/grpc-server/internal/service"
)

var (
	// Version of the app.
	Version = "0.0.1" // nolint
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

	mongoClient := mongo.New(conf.MongoDb, conf.MongoDbURI, conf.MongoPriceCollection, conf.MongoClientConnectionTimeout)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	priceListSrv := service.NewPriceListServer(mongoClient)
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
