package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"color-blind-simulator-1/app/models"
	"color-blind-simulator-1/app/server"
	"color-blind-simulator-1/app/utils"

	"github.com/disintegration/imaging"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	outputDir = "output"
	port      = ":8080"
)

// ImageProcessingRequest represents a request for image processing
type ImageProcessingRequest struct {
	Operation string  `json:"operation"`
	Angle     float64 `json:"angle,omitempty"`
}

type Quiz struct {
	Level       int      `json:"level"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

// InitializeMongoDB sets up the MongoDB connection
func initializeMongoDB() error {
	if err := models.InitializeMongoDB(); err != nil {
		return fmt.Errorf("failed to initialize MongoDB: %v", err)
	}
	return nil
}

// Add this function to clean up old images
func cleanupOutputDirectory() error {
	// Remove all files in the output directory
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			err := os.Remove(filepath.Join(outputDir, file.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Update the handleUpload function to clean up before processing
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clean up old images before processing new ones
	if err := cleanupOutputDirectory(); err != nil {
		log.Printf("Error cleaning up output directory: %v", err)
		// Continue processing even if cleanup fails
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imgData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading image", http.StatusInternalServerError)
		return
	}

	src, err := imaging.Decode(bytes.NewReader(imgData))
	if err != nil {
		http.Error(w, "Error decoding image", http.StatusInternalServerError)
		return
	}

	// Get all selected operations
	operations := r.URL.Query()["operation"]
	angleStr := r.URL.Query().Get("angle")
	angle := 0.0
	if angleStr != "" {
		angle, _ = strconv.ParseFloat(angleStr, 64)
	}

	// Create output directory
	os.MkdirAll(outputDir, 0755)

	// Process the image with all selected operations
	processedImage := src
	var imageURLs []string

	// Save original image
	originalPath := filepath.Join(outputDir, "original.jpg")
	imaging.Save(processedImage, originalPath)
	imageURLs = append(imageURLs, "/output/original.jpg")

	// Apply each operation and save intermediate results
	for i, operation := range operations {
		switch operation {
		case "flip":
			processedImage = utils.FlipImage(processedImage)
		case "rotate":
			processedImage = utils.RotateImage(processedImage, angle)
		case "rotate_shear":
			processedImage = utils.RotateImageWithShear(processedImage, angle)
		case "grayscale":
			processedImage = utils.ConvertToGrayscale(processedImage)
		case "box_blur":
			processedImage = utils.ApplyBoxBlur(processedImage)
		case "gaussian_blur":
			processedImage = utils.ApplyGaussianBlur(processedImage)
		case "edge_detection":
			processedImage = utils.ApplyEdgeDetection(processedImage)
		case "protanopia":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.ProtanopiaMatrix)
		case "deuteranopia":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.DeuteranopiaMatrix)
		case "tritanopia":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.TritanopiaMatrix)
		case "protanomaly":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.ProtanomalyMatrix)
		case "deuteranomaly":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.DeuteranomalyMatrix)
		case "tritanomaly":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.TritanomalyMatrix)
		case "achromatopsia":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.AchromatopsiaMatrix)
		case "monochromacy":
			processedImage = utils.SimulateColorBlindness(processedImage, utils.MonochromacyMatrix)
		case "daltonize":
			processedImage = utils.Daltonize(processedImage, utils.ProtanopiaMatrix)
		}

		// Save intermediate result
		outputPath := filepath.Join(outputDir, fmt.Sprintf("step_%d_%s.jpg", i+1, operation))
		imaging.Save(processedImage, outputPath)
		imageURLs = append(imageURLs, "/output/step_"+fmt.Sprintf("%d_%s.jpg", i+1, operation))
	}

	// Return all image URLs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"images":     imageURLs,
		"operations": operations,
	})
}

func setupStaticHandlers() {
	fs := http.FileServer(http.Dir(outputDir))
	http.Handle("/output/", http.StripPrefix("/output/", fs))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(fmt.Sprintf("templates/%s.html", tmpl))
	if err != nil {
		log.Printf("Error loading template %s: %v", tmpl, err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmpl, err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

func visualizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handleUpload(w, r)
		return
	}
	renderTemplate(w, "visualize", nil)
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

	// Initialize MongoDB
	if err := initializeMongoDB(); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer models.CloseMongoDB()

	// Insert sample quizzes
	if err := models.InsertSampleQuizzes(); err != nil {
		log.Printf("Warning: Failed to insert sample quizzes: %v", err)
	}

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Start UDP server in a goroutine
	go server.StartUDPServer()

	// Serve static files
	fs := http.FileServer(http.Dir(outputDir))
	http.Handle("/output/", http.StripPrefix("/output/", fs))

	// Serve static assets
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index", nil)
	})
	http.HandleFunc("/learn", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "learn", nil)
	})
	http.HandleFunc("/quiz", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "quiz", nil)
	})
	http.HandleFunc("/visualize", visualizeHandler)

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

	fmt.Printf("🚀 Server started at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
