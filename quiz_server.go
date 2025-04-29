package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Quiz struct {
	Level       int      `json:"level"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// MongoDB collection
	collection := client.Database("gopro").Collection("quizzes")

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))

	// API endpoint to get quizzes by level
	http.HandleFunc("/api/quizzes", func(w http.ResponseWriter, r *http.Request) {
		// Get the level parameter from the URL query
		levelParam := r.URL.Query().Get("level")
		if levelParam == "" {
			http.Error(w, "Missing level parameter", http.StatusBadRequest)
			return
		}

		// Convert level parameter to integer
		level, err := strconv.Atoi(levelParam)
		if err != nil {
			http.Error(w, "Invalid level parameter", http.StatusBadRequest)
			return
		}

		// Query the database for quizzes at the specified level
		filter := bson.M{"level": level}
		cursor, err := collection.Find(context.TODO(), filter)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		// Store the quizzes in a slice
		var quizzes []Quiz
		for cursor.Next(context.TODO()) {
			var quiz Quiz
			if err := cursor.Decode(&quiz); err != nil {
				log.Println("Error decoding quiz:", err)
				continue
			}
			quizzes = append(quizzes, quiz)
		}

		// Check if we encountered any error during iteration
		if err := cursor.Err(); err != nil {
			http.Error(w, "Error iterating cursor", http.StatusInternalServerError)
			return
		}

		// Return quizzes as JSON response
		w.Header().Set("Content-Type", "application/json")
		if len(quizzes) == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "No quizzes found for the specified level",
			})
			return
		}

		// Return the quizzes as a JSON response
		json.NewEncoder(w).Encode(quizzes)
	})

	// Start the server
	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
