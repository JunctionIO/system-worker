package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewAPIClient(baseURL, token string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		token:   token,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) PostStatus(msg StatusMessage) error {
	payload := struct {
		LogID       string  `json:"log_id"`
		Status      string  `json:"status"`
		AttemptedAt string  `json:"attempted_at"`
		Error       *string `json:"error"`
	}{
		LogID:       msg.LogID,
		Status:      msg.Status,
		AttemptedAt: msg.AttemptedAt,
		Error:       msg.Error,
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/system/status", bytes.NewReader(body))

	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Junction-Token", c.token)

	resp, err := c.http.Do(req)

	if err != nil {
		return fmt.Errorf("post status: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected response status %d", resp.StatusCode)
	}

	return nil
}
