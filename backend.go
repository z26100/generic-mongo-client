package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type Document interface {
	SetId(id primitive.ObjectID) Document
	GetId() primitive.ObjectID
}

const (
	timeout         = 5 * time.Second
	documentIDField = "_id"
)

type Backend interface{}

type MongoClient struct {
	client        *mongo.Client
	config        *MongoConfig
	databaseLimit []string
}

func NewMongoClient(conf *MongoConfig) (*MongoClient, error) {
	if conf == nil {
		conf = DefaultMongoConfig()
	}
	var cred *options.Credential
	if conf.MongoUser != "" {
		cred = &options.Credential{Username: conf.MongoUser, Password: conf.MongoPassword}
	}
	client, err := getClient(conf.MongoUri, cred)
	if err != nil {
		return nil, err
	}
	err = client.Ping(Ctx(), nil)
	if err != nil {
		return nil, err
	}
	return &MongoClient{
		config:        conf,
		client:        client,
		databaseLimit: conf.databaseLimit,
	}, nil
}

func (b MongoClient) Client() *mongo.Client {
	return b.client
}

func (b MongoClient) Close() error {
	return b.client.Disconnect(Ctx())
}

func getClient(uri string, credentials *options.Credential) (*mongo.Client, error) {
	var client *mongo.Client
	var err error

	if credentials != nil {
		client, err = mongo.NewClient(options.Client().ApplyURI(uri).SetAuth(*credentials))
	} else {
		client, err = mongo.NewClient(options.Client().ApplyURI(uri))
	}
	if err != nil {
		return nil, err
	}
	return client, connect(client)
}

func connect(client *mongo.Client) error {
	return client.Connect(Ctx())
}
func (s MongoClient) ping() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctx.Done()
	err := s.client.Ping(ctx, readpref.Primary())
	if err != nil {
		cancelFunc()
		return err
	}
	cancelFunc = nil
	return nil
}
func _getDatabases(client *mongo.Client) (mongo.ListDatabasesResult, error) {
	result, err := client.ListDatabases(Ctx(), bson.M{})
	return result, err
}

func Ctx() context.Context {
	return context.Background()
}
