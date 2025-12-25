package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbClient *mongo.Client

type URL struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	OriginalURL string             `bson:"originalUrl" json:"originalUrl"`
	ShortCode   string             `bson:"shortCode" json:"shortCode"`
}

type POSTURL struct {
	Name        string `bson:"name" json:"name"`
	OriginalURL string `bson:"originalUrl" json:"originalUrl"`
	ShortCode   string `bson:"shortCode" json:"shortCode"`
}

func initDB() {
	if dbClient != nil {
		return
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set MONGODB_URI environment variable")
	}

	clientOptions := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	dbClient = client
	fmt.Println("Connected to MongoDB Atlas!")
}

func getAllUrls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	initDB() // Ensure DB is connected

	collection := dbClient.Database("url-shortner").Collection("lists")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []URL
	if err := cursor.All(ctx, &results); err != nil {
		http.Error(w, "Failed to decode results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

func addShortUrls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user POSTURL

	// Decode JSON body
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	initDB() // Ensure DB is connected

	collection := dbClient.Database("url-shortner").Collection("lists")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.InsertOne(ctx, user)

	if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(cursor)

	fmt.Println("Received user:", user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
	})
}

func init() {
	err := godotenv.Load() // Loads the .env file from the current directory
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("MONGODB_URI:", os.Getenv("MONGODB_URI")) // Output: Tho
}
func main() {

	// Ensure DB connection on startup
	initDB()

	// Optional: Graceful shutdown
	defer func() {
		if dbClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := dbClient.Disconnect(ctx); err != nil {
				log.Println("Error disconnecting from MongoDB:", err)
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "URL Shortener API is running!")
	})

	http.HandleFunc("/urls", getAllUrls)
	http.HandleFunc("/addUrl", addShortUrls)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
