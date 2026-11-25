package main

import (
	"errors"
	"os"
	"strconv"
)

// config is every setting the server reads from the environment.
// It is loaded once at startup and then passed to whatever needs it.
type config struct {
	port          int
	plaidClientID string
	plaidSecret   string
	plaidEnv      string // "sandbox" or "production"
}

// loadConfig reads the environment into a config.
func loadConfig() (config, error) {
	cfg := config{
		port:          8080,
		plaidClientID: os.Getenv("PLAID_CLIENT_ID"),
		plaidSecret:   os.Getenv("PLAID_SECRET"),
		plaidEnv:      os.Getenv("PLAID_ENV"),
	}

	var errs []error

	if p := os.Getenv("PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			errs = append(errs, errors.New("PORT must be an integer from 1 to 65535"))
		} else {
			cfg.port = n
		}
	}

	if cfg.plaidClientID == "" {
		errs = append(errs, errors.New("PLAID_CLIENT_ID is required"))
	}
	if cfg.plaidSecret == "" {
		errs = append(errs, errors.New("PLAID_SECRET is required"))
	}

	switch cfg.plaidEnv {
	case "sandbox", "production":
		// valid
	default:
		errs = append(errs, errors.New("PLAID_ENV must be sandbox or production"))
	}

	if err := errors.Join(errs...); err != nil {
		return config{}, err
	}
	return cfg, nil
}
