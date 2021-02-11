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
}

type factory struct {
	client               *mongo.Client
	cancel               context.CancelFunc
	mongoDb              string
	mongoPriceCollection string
}

// Product is a model.
type Product struct {
	// ObjectID   primitive.ObjectID `json:"-" bson:"_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	LastUpdate int64   `json:"lastUpdate"`
	// ID         primitive.ObjectID `bson:"_id,omitempty"` // only uppercase variables can be exported
	ObjectID primitive.ObjectID `json:"-" bson:"_id"`
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
	var existingProduct Product

	collection := f.client.Database(f.mongoDb).Collection(f.mongoPriceCollection)
	filter := bson.M{
		"name": bson.M{
			"$eq": name,
		},
	}
	result := collection.FindOne(context.TODO(), filter)

	err := result.Decode(&existingProduct)
	if err != nil {
		if err == mongo.ErrNoDocuments { // need to create new document
			p := Product{name, price, time.Now().UnixNano(), primitive.NewObjectID()}
			insertResult, err := collection.InsertOne(context.TODO(), p)
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

	update := bson.M{
		"$set": bson.M{
			"price":      price,
			"lastupdate": time.Now().UnixNano(),
		},
	}

	updatedResult, err := collection.UpdateOne(
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
