package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"luxmoney/backend/plaidapi"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config:\n%v", err)
	}

	plaidClient := plaidapi.New(cfg.plaidClientID, cfg.plaidSecret, cfg.plaidEnv)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.Handle("POST /api/item/public-token/exchange", handleExchangePublicToken(plaidClient))

	// Sandbox-only shortcut: in production this route doesn't exist at all.
	if cfg.plaidEnv == "sandbox" {
		mux.Handle("POST /api/sandbox/public-token", handleSandboxPublicToken(plaidClient))
	}

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

// handleSandboxPublicToken returns a handler that creates a sandbox public
// token, standing in for the Plaid Link UI while we test with curl.
func handleSandboxPublicToken(plaidClient *plaidapi.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := plaidClient.CreateSandboxPublicToken(r.Context())
		if err != nil {
			log.Printf("handleSandboxPublicToken: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "plaid request failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"public_token": token})
	}
}

// handleExchangePublicToken returns a handler that trades the app's public
// token for an access token. The access token is never sent back.
func handleExchangePublicToken(plaidClient *plaidapi.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			PublicToken string `json:"public_token"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<10) // 1 KB is plenty for a token
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PublicToken == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "public_token is required"})
			return
		}

		item, err := plaidClient.ExchangePublicToken(r.Context(), body.PublicToken)
		if err != nil {
			log.Printf("handleExchangePublicToken: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "plaid request failed"})
			return
		}
		// TODO: store item.AccessToken. Until storage exists, it's dropped here.
		writeJSON(w, http.StatusOK, map[string]string{"item_id": item.ItemID})
	}
}

// writeJSON is our res.status(code).json(data).
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
