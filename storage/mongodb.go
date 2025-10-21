package storage

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PopulateSpec struct {
	Field         string
	RefCollection string
}

type Config struct {
	URI          string
	DatabaseName string
}

var (
	mongoClientInstance *mongo.Client
	databaseName        string
	clientInit          sync.Once
	configError         error
)

func Initialize(cfg Config) error {
	clientInit.Do(func() {
		if cfg.URI == "" {
			configError = fmt.Errorf("MongoDB URI cannot be empty")
			return
		}
		if cfg.DatabaseName == "" {
			configError = fmt.Errorf("database name cannot be empty")
			return
		}

		databaseName = cfg.DatabaseName
		clientOptions := options.Client().ApplyURI(cfg.URI)
		mongoClientInstance, configError = mongo.Connect(context.Background(), clientOptions)
		if configError != nil {
			log.Printf("Failed to initialize MongoDB client: %v", configError)
			return
		}

		if pingErr := mongoClientInstance.Ping(context.Background(), nil); pingErr != nil {
			log.Printf("Failed to ping MongoDB: %v", pingErr)
			configError = pingErr
			return
		}

		log.Println("Connected to DB")
	})
	return configError
}

func getMongoClient() (*mongo.Client, error) {
	if mongoClientInstance == nil {
		return nil, fmt.Errorf("MongoDB client not initialized. Call Initialize() first")
	}
	if configError != nil {
		return nil, configError
	}
	return mongoClientInstance, nil
}

func CheckCollectionExists(ctx context.Context, collectionName string) (string, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return "", fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)

	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return "", fmt.Errorf("failed to list collections: %w", err)
	}

	for _, name := range collections {
		if name == collectionName {
			return "Collection already exists", nil
		}
	}

	if err := db.CreateCollection(ctx, collectionName); err != nil {
		return "", fmt.Errorf("failed to create collection %s: %w", collectionName, err)
	}

	return "Collection " + collectionName + " created successfully", nil
}

func GetCollectionRef(ctx context.Context, collectionName string) *mongo.Collection {
	client, err := getMongoClient()
	if err != nil {
		log.Printf("Failed to get mongo client: %v", err)
		return nil
	}
	db := client.Database(databaseName)
	return db.Collection(collectionName)
}

func AggregateDocuments(ctx context.Context, collectionName string, pipeline mongo.Pipeline) ([]bson.M, error) {
	client, err := getMongoClient()
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

func InsertData(ctx context.Context, collectionName string, data any) (*mongo.InsertOneResult, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to insert data: %w", err)
	}

	return result, nil
}

func FindData(ctx context.Context, collectionName string, filter any, page int, pageSize int) ([]any, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	skip := (page - 1) * pageSize
	limit := int64(pageSize)

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find data: %w", err)
	}
	defer cursor.Close(ctx)

	var results []any
	for cursor.Next(ctx) {
		var result any
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func FindDataNoPagination(ctx context.Context, collectionName string, filter any, sort any) ([]any, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	findOptions := options.Find()
	findOptions.SetSort(sort)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find data: %w", err)
	}
	defer cursor.Close(ctx)

	var results []any
	for cursor.Next(ctx) {
		var result any
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func FindSortedData(ctx context.Context, collectionName string, filter any, page int, pageSize int, sort any) ([]any, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	skip := (page - 1) * pageSize
	limit := int64(pageSize)

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(limit)
	findOptions.SetSort(sort)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find data: %w", err)
	}
	defer cursor.Close(ctx)

	var results []any
	for cursor.Next(ctx) {
		var result any
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func FindById(ctx context.Context, collectionName string, id primitive.ObjectID) (any, error) {
	filter := bson.M{"_id": id}
	results, err := FindOne(ctx, collectionName, filter)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func FindOne(ctx context.Context, collectionName string, filter any) (any, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	var result bson.M
	if err := collection.FindOne(ctx, filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find data: %w", err)
	}
	return result, nil
}

func FindAllData(ctx context.Context, collectionName string, page int, pageSize int) ([]any, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	skip := (page - 1) * pageSize
	limit := int64(pageSize)

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(limit)

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find data: %w", err)
	}
	defer cursor.Close(ctx)

	var results []any
	for cursor.Next(ctx) {
		var result any
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func UpdateOne(ctx context.Context, collectionName string, filter any, update any) (*mongo.UpdateResult, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	updateDoc := bson.M{"$set": update}

	result, err := collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to update data: %w", err)
	}

	return result, nil
}

func DeleteOne(ctx context.Context, collectionName string, filter any) (*mongo.DeleteResult, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete data: %w", err)
	}

	return result, nil
}

func DeleteMany(ctx context.Context, collectionName string, filter any) (*mongo.DeleteResult, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete data: %w", err)
	}

	return result, nil
}

func CountDocuments(ctx context.Context, collectionName string, filter any) (int64, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return 0, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

func DeleteAllData(ctx context.Context, collectionName string) error {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	_, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to delete all data: %w", err)
	}

	return nil
}

func InsertMany(ctx context.Context, collectionName string, data []any) (*mongo.InsertManyResult, error) {
	client, connectionError := getMongoClient()
	if connectionError != nil {
		return nil, fmt.Errorf("error: %w", connectionError)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)
	result, err := collection.InsertMany(ctx, data)

	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return result, nil
}

func FindAndPopulate(ctx context.Context, collectionName string, filter any, populates []PopulateSpec) ([]bson.M, error) {
	client, err := getMongoClient()
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", collectionName, err)
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	// Populate requested fields
	for i, doc := range docs {
		for _, spec := range populates {
			if id, ok := doc[spec.Field].(primitive.ObjectID); ok {
				refColl := db.Collection(spec.RefCollection)
				var refDoc bson.M
				if err := refColl.FindOne(ctx, bson.M{"_id": id}).Decode(&refDoc); err == nil {
					docs[i][spec.Field] = refDoc
				}
			}
		}
	}

	return docs, nil
}

func FindAndPopulateWithPagination(ctx context.Context, collectionName string, filter any, populates []PopulateSpec, page int, pageSize int, sort bson.M) ([]bson.M, error) {
	client, err := getMongoClient()
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	findOptions := options.Find()

	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	skip := int64((page - 1) * pageSize)
	findOptions.SetSkip(skip)
	findOptions.SetSort(sort)
	findOptions.SetLimit(int64(pageSize))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", collectionName, err)
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	// Populate requested fields
	for i, doc := range docs {
		for _, spec := range populates {
			if id, ok := doc[spec.Field].(primitive.ObjectID); ok {
				refColl := db.Collection(spec.RefCollection)
				var refDoc bson.M
				if err := refColl.FindOne(ctx, bson.M{"_id": id}).Decode(&refDoc); err == nil {
					docs[i][spec.Field] = refDoc
				}
			}
		}
	}

	return docs, nil
}

func SumColumn(ctx context.Context, coll *mongo.Collection, field string, match any) (float64, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$" + field}}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	total, ok := results[0]["total"].(float64)
	if !ok {
		return 0, nil
	}

	return total, nil
}
