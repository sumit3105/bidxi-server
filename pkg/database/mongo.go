package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoClient holds the MongoDB client and database reference.
type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// global singleton
var DB *MongoClient

// Connect establishes a connection to MongoDB and verifies it with a ping.
func Connect(uri, dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(30 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	DB = &MongoClient{
		Client:   client,
		Database: client.Database(dbName),
	}

	log.Printf("✅ Connected to MongoDB database: %s", dbName)
	return nil
}

// Disconnect gracefully closes the MongoDB connection.
func Disconnect() {
	if DB != nil && DB.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := DB.Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("🔌 Disconnected from MongoDB")
		}
	}
}

// Collection returns a handle to a named collection in the database.
func Collection(name string) *mongo.Collection {
	return DB.Database.Collection(name)
}
