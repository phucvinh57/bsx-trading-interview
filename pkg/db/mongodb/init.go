package mongodb

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Order *mongo.Collection
var Raw *mongo.Database

func Init() {
	mongodbUri := os.Getenv("MONGODB_URI")
	if len(mongodbUri) == 0 {
		panic("MONGODB_URI is required")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongodbUri).SetServerAPIOptions(serverAPI)

	var err error
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.M{"ping": 1}).Decode(&result); err != nil {
		panic(err)
	}

	Raw = client.Database("bsx-trading")
	Order = Raw.Collection("orders")
}
