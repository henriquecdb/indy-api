package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Race struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Date string  `json:"date"`
	Time *string `json:"time,omitempty"`
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				fmt.Printf("warning: could not set %s from .env\n", key)
			}
		}
	}
}

func getRacesFromDB(ctx context.Context, pool *pgxpool.Pool, category string) ([]Race, error) {
	var query string
	var args []interface{}

	if category != "" {
		query = "SELECT id, name, date, time FROM races WHERE LOWER(category) = LOWER($1) ORDER BY date ASC"
		args = append(args, category)
	} else {
		query = "SELECT id, name, date, time FROM races ORDER BY date ASC"
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var races []Race
	for rows.Next() {
		var r Race
		var t *string
		var dt time.Time
		if err := rows.Scan(&r.ID, &r.Name, &dt, &t); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		r.Date = dt.Format("2006-01-02")
		r.Time = t
		races = append(races, r)
	}

	return races, rows.Err()
}

func listRaces(w http.ResponseWriter, r *http.Request, category string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		http.Error(w, "DATABASE_URL not set", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer pool.Close()

	races, err := getRacesFromDB(ctx, pool, category)
	if err != nil {
		log.Printf("error fetching races: %v", err)
		http.Error(w, "Error fetching races", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	if err := json.NewEncoder(w).Encode(races); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

func handleEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/indy", func(w http.ResponseWriter, r *http.Request) {
		listRaces(w, r, "indy")
	})
	mux.HandleFunc("/imsa", func(w http.ResponseWriter, r *http.Request) {
		listRaces(w, r, "imsa")
	})
}

func main() {
	loadDotEnv()
	log.SetOutput(os.Stdout)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Racing API is running."))
	})

	handleEndpoints(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
