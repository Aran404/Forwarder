package database

import (
	"context"
	"log"
	"reflect"

	"github.com/Aran404/Forwarder/api/types"
	"go.mongodb.org/mongo-driver/bson"
)

// Write writes information to the collection
func (c *Connection) Write(ctx context.Context, name string, w any) error {
	_, err := c.Get(name).InsertOne(ctx, w)
	return err
}

// Update updates the collection based on a query
func (c *Connection) Update(ctx context.Context, name string, query, data any) error {
	update := bson.M{
		"$set": data,
	}
	_, err := c.Get(name).UpdateOne(ctx, query, update)
	return err
}

// Exists checks if a query matches in a collection
func (c *Connection) Exists(ctx context.Context, name string, query any) (bool, error) {
	count, err := c.Get(name).CountDocuments(ctx, query)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CollectionExists checks if a collection exists
func (c *Connection) CollectionExists(ctx context.Context, name string) (bool, error) {
	db := c.Client.Database(DatabaseName)

	collections, err := db.ListCollectionNames(ctx, bson.D{{}})
	if err != nil {
		return false, err
	}

	for _, v := range collections {
		if name == v {
			return true, nil
		}
	}

	return false, nil
}

// Read reads the information in a collection
func (c *Connection) Read(ctx context.Context, name string, v *[]bson.M) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return types.ErrMustBePointer
	}

	cursor, err := c.Get(name).Find(ctx, bson.D{{}})
	if err != nil {
		return err
	}

	return cursor.All(ctx, v)
}

// Convert converts []bson.M into a []*T
func Convert[T Production](c *Connection, col string, v []bson.M) ([]*T, error) {
	var result []*T
	for _, item := range v {
		var temp T
		encoded, err := bson.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := bson.Unmarshal(encoded, &temp); err != nil {
			return nil, err
		}

		result = append(result, &temp)
	}
	return result, nil
}

// Filter looks for a match with the given search query
func (c *Connection) Filter(ctx context.Context, name string, query any, allowCollisions bool) ([]bson.M, error) {
	cursor, err := c.Get(name).Find(ctx, query)
	if err != nil {
		return nil, err
	}

	var matched []bson.M
	if err = cursor.All(ctx, &matched); err != nil {
		return nil, err
	}

	if len(matched) <= 0 {
		return nil, types.ErrNotFound
	}

	if allowCollisions && len(matched) > 1 {
		return nil, types.ErrFilterCollision
	}

	return matched, nil
}

// Delete deletes an item from the collection that matches the query
func (c *Connection) Delete(ctx context.Context, name string, query any) error {
	count, err := c.Get(name).DeleteMany(ctx, query)
	if count.DeletedCount <= 0 {
		return types.ErrNotFound
	}
	return err
}

// Drop drops the collection
func (c *Connection) Drop(ctx context.Context, name string) {
	coll := c.Get(name)
	if err := coll.Drop(ctx); err != nil {
		log.Fatalf("Could not drop collection. Collection Name: %v, Error: %v", name, err)
	}
}

// DropAll drops all the collection
func (c *Connection) DropAll(ctx context.Context) {
	db := c.Client.Database(DatabaseName)

	collections, err := db.ListCollectionNames(ctx, bson.D{{}})
	if err != nil {
		log.Fatalf("Could not get all collections, constants.Error: %v", err)
	}

	for _, name := range collections {
		c.Drop(ctx, name)
	}
}
