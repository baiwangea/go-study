package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User defines the structure for a user document in MongoDB.

type User struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
	Age   int    `bson:"age"`
}

func main() {
	// Use the SetServerAPIOptions() method to set the Stable API version
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetServerAPIOptions(serverAPI)

	// Create a context with a 10-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Check the connection.
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	fmt.Println("Successfully connected to MongoDB!")

	// Get a handle for your collection.
	collection := client.Database("testdb").Collection("users")

	// Clean up the collection before running examples to ensure a fresh start.
	collection.Drop(ctx)

	// Run CRUD examples.
	createExample(ctx, collection)
	readExample(ctx, collection)
	updateExample(ctx, collection)
	deleteExample(ctx, collection)

	// Disconnect from MongoDB.
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
		fmt.Println("\nConnection to MongoDB closed.")
	}()
}

func createExample(ctx context.Context, collection *mongo.Collection) {
	fmt.Println("\n--- Create Example ---")

	// Create a single user document.
	user1 := User{Name: "Alice", Email: "alice@example.com", Age: 28}
	insertOneResult, err := collection.InsertOne(ctx, user1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertOneResult.InsertedID)

	// Create multiple user documents.
	users := []interface{}{ // Must be a slice of interface{}
		User{Name: "Bob", Email: "bob@example.com", Age: 35},
		User{Name: "Charlie", Email: "charlie@example.com", Age: 28},
	}
	insertManyResult, err := collection.InsertMany(ctx, users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)
}

func readExample(ctx context.Context, collection *mongo.Collection) {
	fmt.Println("\n--- Read Example ---")

	// Find a single document.
	var singleResult User
	filter := bson.D{{Key: "name", Value: "Alice"}}
	err := collection.FindOne(ctx, filter).Decode(&singleResult)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found a single document: %+v\n", singleResult)

	// Find multiple documents (e.g., all users with age 28).
	findOptions := options.Find()
	findOptions.SetLimit(2)

	multiFilter := bson.D{{Key: "age", Value: 28}}
	cur, err := collection.Find(ctx, multiFilter, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var results []User
	// Iterate through the cursor.
	for cur.Next(ctx) {
		var elem User
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found multiple documents: %+v\n", results)
}

func updateExample(ctx context.Context, collection *mongo.Collection) {
	fmt.Println("\n--- Update Example ---")

	// Update a single document: Set Bob's age to 36.
	filter := bson.D{{Key: "name", Value: "Bob"}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "age", Value: 36},
		}},
	}

	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
}

func deleteExample(ctx context.Context, collection *mongo.Collection) {
	fmt.Println("\n--- Delete Example ---")

	// Delete a single document.
	filter := bson.D{{Key: "name", Value: "Charlie"}}
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted %v documents.\n", deleteResult.DeletedCount)

	// You can use DeleteMany to delete multiple documents.
}
