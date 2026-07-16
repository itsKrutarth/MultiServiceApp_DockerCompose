package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.Background()

type Note struct {
	ID   string `json:"id,omitempty" bson:"_id,omitempty"`
	Text string `json:"text" bson:"text"`
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cursor.Close(ctx)

	notes := []Note{}
	for cursor.Next(ctx) {
		var n Note
		cursor.Decode(&n)
		notes = append(notes, n)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(notes)
}

func addNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var n Note
	json.NewDecoder(r.Body).Decode(&n)

	_, err := collection.InsertOne(ctx, bson.M{"text": n.Text})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
}

func handleNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	if r.Method == "GET" {
		getNotes(w, r)
	} else if r.Method == "POST" {
		addNote(w, r)
	}
}

func main() {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
	)

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("mongo connect failed:", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		log.Fatal("mongo ping failed:", err)
	}

	collection = client.Database(os.Getenv("DB_NAME")).Collection("notes")

	http.HandleFunc("/notes", handleNotes)
	log.Println("backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
