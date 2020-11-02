package mongo

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (b MongoClient) InsertOrReplace(database, collection string, filter bson.M, update interface{}) (bson.M, error) {
	opts := &options.FindOneAndReplaceOptions{
		Upsert: aws.Bool(true),
	}
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	result := col.FindOneAndReplace(Ctx(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var resp bson.M
	err = result.Decode(&resp)
	return resp, err
}

func (b MongoClient) InsertOne(database string, collection string, doc bson.M) (bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	res, err := col.InsertOne(Ctx(), doc)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("result must not be nil")
	}
	id := res.InsertedID
	doc, err = FindOne(col, bson.M{documentIDField: id})
	return doc, err
}

func (b MongoClient) ReplaceOne(database string, collection string, filter bson.M, replacement bson.M, opts ...*options.FindOneAndReplaceOptions) (bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	err = FindOneAndReplace(col, filter, replacement, opts...)
	replacement, err = FindOne(col, filter)
	if err != nil {
		return nil, err
	}
	return replacement, err
}

func (b MongoClient) UpdateOne(database string, collection string, filter bson.M, update bson.M) (bson.M, error) {
	upd := bson.M{"$set": update}
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	err = FindOneAndUpdate(col, filter, upd)
	return update, err
}
