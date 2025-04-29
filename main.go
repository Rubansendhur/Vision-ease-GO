package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"color-blind-simulator-1/app/server"
	"color-blind-simulator-1/app/utils"

	"github.com/disintegration/imaging"
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

	fmt.Printf("🚀 Server started at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
