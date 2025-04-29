package handlers

import (
	"net/http"
	"path/filepath"
)

// VisualizeHandler handles the request to display the uploaded image with filters.
func VisualizeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the uploaded image from the output directory
	filename := filepath.Join("output", "uploaded_image.jpg") // Assuming the image is named 'uploaded_image.jpg'
	http.ServeFile(w, r, filename)
}
