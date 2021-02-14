package service

import (
	context "context"
	"errors"
	"log"
	"path/filepath"
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

// Fetch csv file with products.
func (s *priceListServer) Fetch(ctx context.Context, u *URL) (response *Response, err error) {
	var filePath string

	switch u.Url {
	case "http://localhost?1":
		filePath = "./mockdata/test1.csv"
	case "http://localhost?2":
		filePath = "./mockdata/test2.csv"
	case "http://localhost?3":
		filePath = "./mockdata/test3.csv"
	default:
		return nil, errors.New("Not found")
	}

	filePath, err = filepath.Abs(filePath)
	if err != nil {
		log.Fatalf("invalid filepath: %v", err)
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

// List products.
func (s *priceListServer) List(ctx context.Context, params *Params) (products *ProductList, err error) {
	productList := &ProductList{}
	mongoPagingParams := mongo.PagingParams{}
	mongoSortingParams := mongo.SortingParams{}

	if params.PagingParams != nil {
		mongoPagingParams.Page = params.PagingParams.Page
		mongoPagingParams.Limit = params.PagingParams.Limit
	}

	if params.SortingParams != nil {
		mongoSortingParams.Name = params.SortingParams.Name
		mongoSortingParams.Price = params.SortingParams.Price
		mongoSortingParams.LastUpdate = params.SortingParams.LastUpdate
		mongoSortingParams.PriceChangedCount = params.SortingParams.PriceChangedCount
	}

	mongoParams := &mongo.Params{
		PagingParams:  mongoPagingParams,
		SortingParams: mongoSortingParams,
	}

	mongoProductList, err := s.mongoClient.GetProductList(mongoParams)
	if err != nil {
		return nil, err
	}

	for _, p := range mongoProductList {
		productList.Products = append(productList.Products, &Product{
			Name:              p.Name,
			Price:             p.Price,
			LastUpdate:        p.LastUpdate,
			PriceChangedCount: p.PriceChangedCount,
		})
	}

	return productList, nil
}
