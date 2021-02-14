package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FactoryInterface provides an interface.
type FactoryInterface interface {
	InsertProduct(string, float64) (interface{}, error)
	Disconnect()
	GetColection() *mongo.Collection
	GetProductList(*Params) ([]Product, error)
}

type factory struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// Product is a model.
type Product struct {
	// ObjectID   primitive.ObjectID `json:"-" bson:"_id"`
	Name              string             `json:"name"`
	Price             float64            `json:"price"`
	LastUpdate        int64              `json:"lastUpdate"`
	PriceChangedCount int64              `json:"priceChangedCount"`
	ObjectID          primitive.ObjectID `json:"-" bson:"_id"`
}

// Params - parameters to use paging and sorting.
type Params struct {
	PagingParams  PagingParams
	SortingParams SortingParams
}

// PagingParams - parameters to use paging.
type PagingParams struct {
	Page  int64
	Limit int64
}

// SortingParams - parameters to use sorting.
type SortingParams struct {
	Name              string
	Price             float64
	LastUpdate        int64
	PriceChangedCount int64
}

// New is factory constructor.
func New(mongoDb string, mongoDbURI string, mongoPriceCollection string, mongoClientConnectionTimeout int) FactoryInterface {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoDbURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mongoClientConnectionTimeout)*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(mongoDb).Collection(mongoPriceCollection)

	return &factory{client, collection}
}

func (f *factory) Disconnect() {
	err := f.client.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func (f *factory) InsertProduct(name string, price float64) (interface{}, error) {
	var existingProduct Product

	filter := bson.M{
		"name": bson.M{
			"$eq": name,
		},
	}
	result := f.collection.FindOne(context.TODO(), filter)

	err := result.Decode(&existingProduct)
	if err != nil {
		if err == mongo.ErrNoDocuments { // need to create new document
			p := Product{name, price, time.Now().UnixNano(), 0, primitive.NewObjectID()}
			insertResult, err := f.collection.InsertOne(context.TODO(), p)
			if err != nil {
				return nil, err
			}

			return insertResult.InsertedID, nil
		}
		return nil, err
	}

	if existingProduct.Price == price {
		log.Println("exisiting product price doesnt changed")

		return existingProduct.ObjectID, nil
	}

	priceChangedCount := existingProduct.PriceChangedCount + 1
	update := bson.M{
		"$set": bson.M{
			"price":             price,
			"lastupdate":        time.Now().UnixNano(),
			"pricechangedcount": priceChangedCount,
		},
	}

	updatedResult, err := f.collection.UpdateOne(
		context.TODO(),
		filter,
		update,
	)
	if err != nil {
		return nil, err
	}

	// log.Println("UpdateOne() result:", result)
	// log.Println("UpdateOne() result TYPE:", reflect.TypeOf(result))
	log.Println("UpdateOne() result MatchedCount:", updatedResult.MatchedCount)
	log.Println("UpdateOne() result ModifiedCount:", updatedResult.ModifiedCount)
	log.Println("UpdateOne() result UpsertedCount:", updatedResult.UpsertedCount)
	log.Println("UpdateOne() result UpsertedID:", updatedResult.UpsertedID)

	return existingProduct.ObjectID, nil
}

func (f *factory) GetColection() *mongo.Collection {
	return f.collection
}

func (f *factory) GetProductList(params *Params) ([]Product, error) {
	var offset int64
	ctx := context.TODO()
	filter := bson.M{}

	if params.PagingParams.Page > 0 {
		offset = (params.PagingParams.Page - 1) * params.PagingParams.Limit
	}

	findOptions := options.Find() // build a `findOptions`
	findOptions.SetSort(map[string]int64{"_id": 1})
	findOptions.SetSkip(offset) // skip whatever you want, like `offset` clause in mysql
	if params.PagingParams.Limit > 0 {
		findOptions.SetLimit(params.PagingParams.Limit) // like `limit` clause in mysql
	}

	if params.SortingParams.PriceChangedCount > 0 {
		filter = bson.M{
			"pricechangedcount": params.SortingParams.PriceChangedCount,
		}
	}

	if params.SortingParams.Price > 0 {
		findOptions.SetSort(map[string]int64{"price": 1})
	}

	productsCursor, err := f.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	defer productsCursor.Close(ctx)

	var (
		product     Product
		productList []Product
	)

	for productsCursor.Next(ctx) {
		if err = productsCursor.Decode(&product); err != nil {
			return nil, err
		}

		productList = append(productList, product)
	}

	return productList, nil
}
