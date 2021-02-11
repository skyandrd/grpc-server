package service

import (
	context "context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/skyandrd/grpc-server/internal/csvreader"
	"github.com/skyandrd/grpc-server/internal/mongo"
)

type priceListServer struct {
	UnimplementedPriceListServer
	mongoClient mongo.FactoryInterface
}

// NewPriceListServer return PriceListServer protobuf codegen instance.
func NewPriceListServer(mongoClient mongo.FactoryInterface) PriceListServer {
	s := &priceListServer{UnimplementedPriceListServer{}, mongoClient}

	return s
}

// Fetch ...
func (s *priceListServer) Fetch(ctx context.Context, u *URL) (response *Response, err error) {
	var filePath string

	switch u.Url {
	case "http://localhost?1":
		filePath = "./internal/service/mockdata/test1.csv"
	case "http://localhost?2":
		filePath = "./internal/service/mockdata/test2.csv"
	default:
		return nil, errors.New("Not found")
	}

	records, err := csvreader.Read(filePath)
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

		id, err := s.mongoClient.InsertProduct(productName, price)
		if err != nil {
			log.Fatalf("mongodb insertion error: %v", err)
		}

		log.Printf("inserted new product with id=%v", id)

	}

	return &Response{Status: uint32(200)}, nil
}
