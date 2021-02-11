package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FactoryInterface provides an interface.
type FactoryInterface interface {
	InsertProduct(string, float64) (interface{}, error)
	Disconnect()
}

type factory struct {
	client               *mongo.Client
	cancel               context.CancelFunc
	mongoDb              string
	mongoPriceCollection string
}

// Product is a model.
type Product struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// New is factory constructor.
func New(mongoDb string, mongoDbURI string, mongoPriceCollection string, mongoClientConnectionTimeout int) FactoryInterface {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoDbURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mongoClientConnectionTimeout)*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return &factory{client, cancel, mongoDb, mongoPriceCollection}
}

func (f *factory) Disconnect() {
	f.cancel()

	err := f.client.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func (f *factory) InsertProduct(name string, price float64) (interface{}, error) {
	p := Product{name, price}
	collection := f.client.Database(f.mongoDb).Collection(f.mongoPriceCollection)

	insertResult, err := collection.InsertOne(context.TODO(), p)
	if err != nil {
		return nil, err
	}

	return insertResult.InsertedID, nil
}
