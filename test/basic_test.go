package test

import (
	mongo "github.com/z26100/generic-mongo-client"
	"log"
	"testing"
)

func getClient() (*mongo.MongoClient, error) {
	conf := mongo.DefaultMongoConfig()
	conf.MongoUri = "mongodb://localhost:27017"
	conf.MongoUser = "mongoadmin"
	conf.MongoPassword = "secret"
	return mongo.NewMongoClient(conf)
}
func TestPing(t *testing.T) {
	client, err := getClient()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(client)
}
