package main

import (
	"context"
	"log"

	pb "github.com/skyandrd/grpc-server/internal/service"
	"google.golang.org/grpc"
)

const (
	address = "localhost:80"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to: %v", err)
	}
	defer conn.Close()

	client := pb.NewPriceListClient(conn)
	url := &pb.URL{Url: "http://localhost?2"}
	res, err := client.Fetch(context.Background(), url)
	if err != nil {
		log.Fatalf("service fetch url error: %v", err)
	}

	log.Printf("response from grpc fetch url: %v", res)
}
