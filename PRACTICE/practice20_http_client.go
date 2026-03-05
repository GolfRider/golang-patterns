package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UserResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var httpClient = &http.Client{Timeout: time.Second * 10}

func PostUser(url string, userData UserRequest) error {
	// 1. Marshal the struct to JSON
	jsonData, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	// 2. Create the request with a context (best practice for cancellation)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 3. Set necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 4. Execute the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// 5. Always close the body to avoid memory leaks
	defer resp.Body.Close()

	// 6. Check status codes (don't just assume 200)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}

func GetUser(url string) (*UserResponse, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	// Use NewDecoder for streaming efficiency
	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}
