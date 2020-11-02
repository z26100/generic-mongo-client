package mongo

import (
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoConfig struct {
	MongoUser         string
	MongoPassword     string
	MongoUri          string
	Timeout           time.Duration
	databaseLimit     []string
	databaseOptions   *options.DatabaseOptions
	collectionOptions *options.CollectionOptions
}

func DefaultMongoConfig() *MongoConfig {
	return &MongoConfig{
		databaseOptions:   nil,
		databaseLimit:     []string{},
		collectionOptions: nil,
	}
}
