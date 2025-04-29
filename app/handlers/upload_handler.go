package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// UploadHandler handles the image upload request.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Println("Error retrieving the file:", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll("output", os.ModePerm); err != nil {
		log.Println("Error creating output directory:", err)
		http.Error(w, "Error creating output directory", http.StatusInternalServerError)
		return
	}

	// Save the uploaded file
	filename := filepath.Join("output", handler.Filename)
	dst, err := os.Create(filename)
	if err != nil {
		log.Println("Error creating the file:", err)
		http.Error(w, "Error creating the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		log.Println("Error saving the file:", err)
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	log.Println("File uploaded successfully:", handler.Filename)
	http.Redirect(w, r, "/visualize", http.StatusSeeOther)
}
