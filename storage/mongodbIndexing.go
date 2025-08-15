package storage

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Ensure2dsphereIndex(collection *mongo.Collection) {
	homeIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "homeLocation", Value: "2dsphere"},
		},
		Options: options.Index().SetName("homeLocation_2dsphere"),
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), homeIndexModel)
	if err != nil {
		log.Fatalf("Failed to create 2dsphere index: %v", err)
	}
}

func EnsureJob2dsphereIndex(collection *mongo.Collection) {
	jobIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "location", Value: "2dsphere"},
		},
		Options: options.Index().SetName("jobLocation_2dsphere"),
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), jobIndexModel)
	if err != nil {
		log.Fatalf("Failed to create 2dsphere index: %v", err)
	}
}

func EnsureNameIndexes(collection *mongo.Collection) {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "firstName", Value: 1}},
			Options: options.Index().SetName("firstName_index"),
		},
		{
			Keys:    bson.D{{Key: "lastName", Value: 1}},
			Options: options.Index().SetName("lastName_index"),
		},
	}

	_, err := collection.Indexes().CreateMany(context.TODO(), indexes)
	if err != nil {
		log.Fatalf("Failed to create name indexes: %v", err)
	}
}

func CreateTTLIndex(collection *mongo.Collection) {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"createdAt": 1},
		Options: options.Index().SetExpireAfterSeconds(86400),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Failed to create name indexes: %v", err)
	}
}
