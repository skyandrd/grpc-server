package service_test

import (
	"context"
	"log"
	"testing"

	"github.com/skyandrd/grpc-server/internal/config"
	"github.com/skyandrd/grpc-server/internal/mongo"
	"github.com/skyandrd/grpc-server/internal/service"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestServerFetch(t *testing.T) {
	conf, err := config.GetConfig()
	if err != nil {
		t.Fatalf("cannot read service config - %v", err)
	}

	mongoClient := mongo.New(conf.MongoDb, conf.MongoDbURI, conf.MongoPriceCollection, conf.MongoClientConnectionTimeout)

	collection := mongoClient.GetColection()

	ctx := context.Background()

	_, err = collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		t.Fatalf("cannot clear collection - %v", err)
	}

	priceListSrv := service.NewPriceListServer(mongoClient)

	url := &service.URL{Url: "http://localhost?1"}
	resp, err := priceListSrv.Fetch(ctx, url)

	require.NoError(t, err)

	require.Equal(t, uint32(200), resp.Status)

	productsCursor, err := collection.Find(ctx, bson.M{})
	require.NoError(t, err)
	defer productsCursor.Close(ctx)

	var product mongo.Product

	for productsCursor.Next(ctx) {
		if err = productsCursor.Decode(&product); err != nil {
			log.Fatal(err)
		}

		if product.Name == "product1" {
			require.Equal(t, product.Price, 123.12)
			require.Equal(t, product.PriceChangedCount, int64(0))
		} else if product.Name == "product2" {
			require.Equal(t, product.Price, 456.45)
			require.Equal(t, product.PriceChangedCount, int64(0))
		}
	}

	url = &service.URL{Url: "http://localhost?2"}
	resp, err = priceListSrv.Fetch(ctx, url)

	require.NoError(t, err)

	require.Equal(t, uint32(200), resp.Status)

	productsCursor, err = collection.Find(ctx, bson.M{})
	require.NoError(t, err)
	defer productsCursor.Close(ctx)

	for productsCursor.Next(ctx) {
		if err = productsCursor.Decode(&product); err != nil {
			log.Fatal(err)
		}

		if product.Name == "product1" {
			require.Equal(t, product.Price, 999.99)
			require.Equal(t, product.PriceChangedCount, int64(1))
		} else if product.Name == "product3" {
			require.Equal(t, product.Price, 333.33)
			require.Equal(t, product.PriceChangedCount, int64(0))
		}
	}
}

func TestServerList(t *testing.T) {
	conf, err := config.GetConfig()
	if err != nil {
		t.Fatalf("cannot read service config - %v", err)
	}

	mongoClient := mongo.New(conf.MongoDb, conf.MongoDbURI, conf.MongoPriceCollection, conf.MongoClientConnectionTimeout)
	collection := mongoClient.GetColection()

	ctx := context.Background()
	params := &service.Params{}
	priceListSrv := service.NewPriceListServer(mongoClient)

	url := &service.URL{Url: "http://localhost?3"}
	_, err = priceListSrv.Fetch(ctx, url)
	require.NoError(t, err)

	resp, err := priceListSrv.List(ctx, params)
	require.NoError(t, err)

	productsCursor, err := collection.Find(ctx, bson.M{})
	require.NoError(t, err)
	defer productsCursor.Close(ctx)

	var (
		p mongo.Product
	)

	productList := &service.ProductList{}

	for productsCursor.Next(ctx) {
		if err = productsCursor.Decode(&p); err != nil {
			log.Fatal(err)
		}
		productList.Products = append(productList.Products, &service.Product{
			Name:              p.Name,
			Price:             p.Price,
			LastUpdate:        p.LastUpdate,
			PriceChangedCount: p.PriceChangedCount,
		})
	}

	require.Equal(t, productList, resp)
	require.Equal(t, 13, len(resp.Products))

	params.PagingParams = &service.PagingParams{
		Page:  1,
		Limit: 5,
	}
	resp, err = priceListSrv.List(ctx, params)
	require.NoError(t, err)
	require.Equal(t, 5, len(resp.Products))

	params.PagingParams.Page = 2
	resp, err = priceListSrv.List(ctx, params)
	require.NoError(t, err)
	require.Equal(t, 5, len(resp.Products))

	params.PagingParams.Page = 3
	resp, err = priceListSrv.List(ctx, params)
	require.NoError(t, err)
	require.Equal(t, 3, len(resp.Products))

	params.PagingParams.Page = 1
	params.SortingParams = &service.SortingParams{
		PriceChangedCount: 1,
	}
	resp, err = priceListSrv.List(ctx, params)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Products))

	params.PagingParams.Limit = 10
	params.SortingParams.PriceChangedCount = 0
	params.SortingParams.Price = 1
	resp, err = priceListSrv.List(ctx, params)
	require.NoError(t, err)
	require.Equal(t, 10, len(resp.Products))

	require.Equal(t, 333.33, resp.Products[0].Price)
	require.Equal(t, "product9", resp.Products[0].Name)

	require.Equal(t, 456.45, resp.Products[6].Price)
	require.Equal(t, "product2", resp.Products[6].Name)

	require.Equal(t, 999.99, resp.Products[9].Price)
	require.Equal(t, "product10", resp.Products[9].Name)
}
