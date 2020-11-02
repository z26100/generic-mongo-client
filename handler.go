package mongo

import (
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func GetRoutes(mongoClient *MongoClient) []Route {
	routes := []Route{
		{Path: "/{database:[a-z]+}/{collection:[a-z]+}/{document:[a-z,0-9,-]+}", HandlerFc: getDocument(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z]+}/{collection:[a-z]+}/{document:[a-z,0-9,-]+}", HandlerFc: putDocument(mongoClient), Methods: "POST,PUT"},
		{Path: "/{database:[a-z]+}/{collection:[a-z]+}/{document:[a-z,0-9,-]+}", HandlerFc: patchDocument(mongoClient), Methods: "PATCH"},
		{Path: "/{database}/{collection}/{document:[a-z,0-9,-]+}", HandlerFc: deleteDocument(mongoClient), Methods: "DELETE"},
		{Path: "/{database:[a-z]+}/{collection:[a-z]+}", HandlerFc: putDocument(mongoClient), Methods: "POST,PUT"},
		{Path: "/{database:[a-z]+}/{collection:[a-z]+}", HandlerFc: getDocuments(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z]+}", HandlerFc: getCollections(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z]+}/{collection:[a-z]+}", HandlerFc: deleteCollection(mongoClient), Methods: "DELETE"},
		{Path: "/{database:[a-z]+}", HandlerFc: deleteDatabase(mongoClient), Methods: "DELETE"},
		{Path: "/", HandlerFc: getDatabases(mongoClient), Methods: "GET"},
	}
	return routes
}
func getCollections(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		v := r.URL.Query()["nameOnly"]
		database := vars["database"]
		nameOnly := false
		if v != nil && len(v) > 0 {
			nameOnly, _ = strconv.ParseBool(v[0])
		}
		if check(func() bool { return database == "" }, w) {
			return
		}
		data, err := mongoClient.GetCollections(database, nameOnly)
		if data == nil {
			w.Write([]byte("[]"))
			return
		}
		if checkError(err, w) {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func getDocuments(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		if check(func() bool { return collection == "" || database == "" }, w) {
			return
		}
		data, err := mongoClient.FindAll(database, collection)
		if checkError(err, w) {
			return
		}
		if data == nil {
			w.Write([]byte("[]"))
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, false, false)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func deleteDatabase(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]

		if check(func() bool { return database == "" }, w) {
			return
		}
		err := mongoClient.DropDatabase(database)
		if checkError(err, w) {
			return
		}
	}
}

func deleteCollection(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		if check(func() bool { return collection == "" || database == "" }, w) {
			return
		}
		err := mongoClient.DropCollection(database, collection)
		if checkError(err, w) {
			return
		}
	}
}

func getDatabases(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()["nameOnly"]
		nameOnly := false
		if v != nil && len(v) > 0 {
			nameOnly, _ = strconv.ParseBool(v[0])
		}
		data, err := mongoClient.GetDatabases(&options.DatabaseOptions{}, nameOnly)
		if checkError(err, w) {
			return
		}
		if data == nil {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func getDocument(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		document := vars["document"]
		if check(func() bool { return collection == "" || database == "" || document == "" }, w) {
			return
		}

		filter := bson.M{}
		data := make([]bson.M, 0)
		var err error
		switch document {
		case "search":
			for k, v := range r.URL.Query() {
				if strings.HasPrefix(v[0], "_d") {
					numValue, err := strconv.ParseInt(strings.TrimPrefix(v[0], "_d"), 10, 0)
					if err != nil {
						break
					}
					filter[k] = numValue
				} else {
					filter[k] = v[0]
				}
			}
		default:
			filter = bson.M{"_id": document}
		}
		data, err = mongoClient.FindMany(database, collection, filter)
		if checkError(err, w) {
			return
		}
		if data == nil {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, false, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func putDocument(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		document := vars["document"]
		if check(func() bool { return collection == "" || database == "" }, w) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if checkError(err, w) {
			return
		}
		doc := bson.M{}
		err = bson.UnmarshalExtJSON(body, true, &doc)
		if checkError(err, w) {
			return
		}
		var data bson.M
		if document != "" {
			filter := bson.M{"_id": document}
			opts := &options.FindOneAndReplaceOptions{
				Upsert: proto.Bool(true),
			}
			data, err = mongoClient.ReplaceOne(database, collection, filter, doc, opts)
		} else {
			doc["_id"] = uuid.New().String()
			data, err = mongoClient.InsertOne(database, collection, doc)
		}
		if checkError(err, w) {
			return
		}
		if data == nil {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func patchDocument(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		document := vars["document"]
		if check(func() bool { return collection == "" || database == "" || document == "" }, w) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if checkError(err, w) {
			return
		}
		doc := bson.M{}
		err = bson.UnmarshalExtJSON(body, true, &doc)
		if checkError(err, w) {
			return
		}
		id, err := primitive.ObjectIDFromHex(document)
		if checkError(err, w) {
			return
		}
		filter := bson.M{"_id": id}
		data, err := mongoClient.UpdateOne(database, collection, filter, doc)
		if checkError(err, w) {
			return
		}
		if data == nil {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func postDocument(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		if check(func() bool { return collection == "" || database == "" }, w) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if checkError(err, w) {
			return
		}
		doc := bson.M{}
		err = bson.UnmarshalExtJSON(body, true, &doc)
		if checkError(err, w) {
			return
		}

		data, err := mongoClient.InsertOne(database, collection, doc)
		if checkError(err, w) {
			return
		}
		if data == nil {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func deleteDocument(mongoClient *MongoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		id := vars["document"]
		if check(func() bool { return collection == "" || database == "" || id == "" }, w) {
			return
		}
		filter := bson.M{"_id": id}
		err := mongoClient.DeleteOne(database, collection, filter)
		if checkError(err, w) {
			return
		}
	}
}

func check(condition func() bool, w http.ResponseWriter) bool {
	if condition() {
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return true
	}
	return false
}

func checkError(err error, w http.ResponseWriter) bool {
	return check(func() bool {
		if err != nil {
			log.Println(err)
		}
		return err != nil
	}, w)
}

func checkDataAndError(data interface{}, err error, w http.ResponseWriter) bool {
	return check(func() bool {
		if err != nil {
			log.Println(err)
		}
		return err != nil || data == nil
	}, w)
}
