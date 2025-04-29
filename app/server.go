package app

import (
	"log"
	"net/http"
)

// StartServer initializes and starts the HTTP server.
func StartServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})
	http.HandleFunc("/learn", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/learn.html")
	})
	http.HandleFunc("/quiz", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/quiz.html")
	})
	http.HandleFunc("/visualize", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/visualize.html")
	})

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
