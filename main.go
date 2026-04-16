package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Race struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Date string  `json:"date"`
	Time *string `json:"time,omitempty"`
}

var races []Race

func loadData() {
	data, err := os.ReadFile("races.json")
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

	http.HandleFunc("/races", listRaces)

	log.Println("Server running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
