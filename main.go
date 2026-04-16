package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

//go:embed races.json
var embeddedFiles embed.FS

type Race struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Date string  `json:"date"`
	Time *string `json:"time,omitempty"`
}

var races []Race

func loadData() {
	data, err := embeddedFiles.ReadFile("races.json")
	if err != nil {
		log.Fatalf("error reading races.json: %v", err)
	}

	if err := json.Unmarshal(data, &races); err != nil {
		log.Fatalf("error decoding races.json: %v", err)
	}
}

func listRaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(races); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

func main() {
	loadData()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Indy API is running"))
	})
	mux.HandleFunc("/races", listRaces)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
