package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

//go:embed indy.json imsa.json
var embeddedFiles embed.FS

type Race struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Date string  `json:"date"`
	Time *string `json:"time,omitempty"`
}

func loadData(filename string) []Race {
	data, err := embeddedFiles.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading %s: %v", filename, err)
	}

	var result []Race
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatalf("error decoding %s: %v", filename, err)
	}

	return result
}

func listRaces(w http.ResponseWriter, r *http.Request, category string) {
	races := loadData(category)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(races); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

func handleEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/indy", func(w http.ResponseWriter, r *http.Request) {
		listRaces(w, r, "indy.json")
	})
	mux.HandleFunc("/imsa", func(w http.ResponseWriter, r *http.Request) {
		listRaces(w, r, "imsa.json")
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Indy API is running"))
	})

	handleEndpoints(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
