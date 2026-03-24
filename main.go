package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Race struct {
	ID   string    `json:"id"`
	Race string    `json:"race"`
	Date time.Time `json:"date"`
}

var races []Race

func loadData() {
	arquivo, err := os.ReadFile("races.json")
	if err != nil {
		log.Fatalf("Error while reading: %v", err)
	}

	err = json.Unmarshal(arquivo, &races)
	if err != nil {
		log.Fatalf("Error while decoding json: %v", err)
	}
}

func getRaces(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(races)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func main() {
	loadData()

	http.HandleFunc("/races", getRaces)

	println("Server running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
