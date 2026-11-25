package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config:\n%v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on :%d (plaid: %s)", cfg.port, cfg.plaidEnv)
	log.Fatal(srv.ListenAndServe())
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// writeJSON is our res.status(code).json(data).
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
