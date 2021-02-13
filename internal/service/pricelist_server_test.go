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

func TestServer(t *testing.T) {
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
