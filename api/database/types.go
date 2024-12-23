package database

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ConnectionURI = "mongodb://localhost:27017"
	DatabaseName  = "PaymentProcessor"

	ConnectionTimeout = time.Second * 90
)

type (
	Connection struct {
		Client      *mongo.Client
		Collections map[string]*mongo.Collection
	}

	Production interface {
		
	}
)
