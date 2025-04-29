package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Quiz struct {
	Level       int      `json:"level"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

// QuizHandler handles quiz-related requests
func QuizHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
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
}

func handleAddQuiz(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var quiz Quiz
	if err := json.NewDecoder(r.Body).Decode(&quiz); err != nil {
		http.Error(w, "Invalid quiz data", http.StatusBadRequest)
		return
	}

	// Validate quiz data
	if quiz.Question == "" || len(quiz.Options) == 0 || quiz.Answer == "" {
		http.Error(w, "Missing required quiz fields", http.StatusBadRequest)
		return
	}

	// Insert quiz into database
	_, err := collection.InsertOne(context.TODO(), quiz)
	if err != nil {
		http.Error(w, "Error adding quiz", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Quiz added successfully",
	})
}
