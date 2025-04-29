package models

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Quiz represents a quiz question
type Quiz struct {
	ID          string   `bson:"_id,omitempty" json:"id"`
	Level       int      `bson:"level" json:"level"`
	Question    string   `bson:"question" json:"question"`
	Options     []string `bson:"options" json:"options"`
	Answer      string   `bson:"answer" json:"answer"`
	Explanation string   `bson:"explanation" json:"explanation"`
}

var client *mongo.Client
var collection *mongo.Collection

// InitializeMongoDB sets up the MongoDB connection
func InitializeMongoDB() error {
	// MongoDB connection string
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb+srv://user1:Dharun2004@cluster0.aj2kz0d.mongodb.net/colorblind?retryWrites=true&w=majority"
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	// Get collection
	collection = client.Database("colorblind").Collection("quizzes")

	// Create indexes
	_, err = collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "level", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// GetQuizzesByLevel retrieves quizzes for a specific level
func GetQuizzesByLevel(level int) ([]Quiz, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"level": level}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quizzes []Quiz
	if err = cursor.All(ctx, &quizzes); err != nil {
		return nil, err
	}

	return quizzes, nil
}

// InsertQuiz adds a new quiz to the database
func InsertQuiz(quiz Quiz) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, quiz)
	return err
}

// InsertSampleQuizzes adds sample quiz data to the database
func InsertSampleQuizzes() error {
	sampleQuizzes := []Quiz{
		{
			Level:       1,
			Question:    "What is the most common type of color blindness?",
			Options:     []string{"Protanopia", "Deuteranopia", "Tritanopia", "Achromatopsia"},
			Answer:      "Protanopia",
			Explanation: "Protanopia is the most common type of color blindness, affecting about 1% of males.",
		},
		{
			Level:       1,
			Question:    "Which gender is more likely to be color blind?",
			Options:     []string{"Male", "Female", "Both equally", "Neither"},
			Answer:      "Male",
			Explanation: "Color blindness is more common in males because the genes for color vision are located on the X chromosome.",
		},
		{
			Level:       2,
			Question:    "What colors are typically difficult to distinguish for someone with deuteranopia?",
			Options:     []string{"Red and green", "Blue and yellow", "All colors", "None of the above"},
			Answer:      "Red and green",
			Explanation: "Deuteranopia is a type of red-green color blindness where the green cones are missing.",
		},
		{
			Level:       2,
			Question:    "Can color blindness be cured?",
			Options:     []string{"Yes, with surgery", "Yes, with medication", "No, but can be managed", "Yes, with glasses"},
			Answer:      "No, but can be managed",
			Explanation: "Color blindness cannot be cured, but special glasses and apps can help manage the condition.",
		},
		{
			Level:       3,
			Question:    "What is the scientific term for complete color blindness?",
			Options:     []string{"Monochromacy", "Achromatopsia", "Both A and B", "None of the above"},
			Answer:      "Both A and B",
			Explanation: "Both monochromacy and achromatopsia refer to complete color blindness, where a person sees only in shades of gray.",
		},
	}

	for _, quiz := range sampleQuizzes {
		if err := InsertQuiz(quiz); err != nil {
			return err
		}
	}
	return nil
}

// CloseMongoDB closes the MongoDB connection
func CloseMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Disconnect(ctx)
}
