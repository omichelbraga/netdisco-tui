package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	BaseURL    string
	APIKey     string
	HTTPTimeout time.Duration
	MaxRetries int
}

func New() (*Config, error) {
	baseURL := os.Getenv("NETDISCO_URL")
	if baseURL == "" {
		baseURL = "https://netdisco.example.com/api/v1"
	}

	apiKey := os.Getenv("NETDISCO_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("NETDISCO_TOKEN env var is required")
	}

	timeout := 30 * time.Second
	if t := os.Getenv("NETDISCO_TIMEOUT"); t != "" {
		if secs, err := strconv.Atoi(t); err == nil {
			timeout = time.Duration(secs) * time.Second
		}
	}

	maxRetries := 3
	if r := os.Getenv("NETDISCO_RETRIES"); r != "" {
		if n, err := strconv.Atoi(r); err == nil {
			maxRetries = n
		}
	}

	return &Config{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		HTTPTimeout: timeout,
		MaxRetries:  maxRetries,
	}, nil
}
