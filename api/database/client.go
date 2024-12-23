package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewConn(ctx context.Context) *Connection {
	options := options.Client().
		ApplyURI(ConnectionURI).
		SetConnectTimeout(ConnectionTimeout)

	client, err := mongo.Connect(ctx, options)
	if err != nil {
		log.Fatalf("Could not create mongo client, Error: %v", err.Error())
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Could not ping mongo client, Error: %v", err.Error())
	}
	return &Connection{Client: client, Collections: make(map[string]*mongo.Collection)}
}

func (c *Connection) NewCollection(col string) *mongo.Collection {
	if _, ok := c.Collections[col]; ok {
		return c.Collections[col]
	}

	collection := c.Client.Database(DatabaseName).Collection(col)
	c.Collections[col] = collection
	return collection
}

// Get gets a collection in the database
func (c *Connection) Get(col string) *mongo.Collection {
	return c.NewCollection(col)
}

// Close closes the mongo connection
func (c *Connection) Close(ctx context.Context) {
	if err := c.Client.Disconnect(ctx); err != nil {
		log.Fatalf("Could not close mongo client, Error: %v", err.Error())
	}
}
