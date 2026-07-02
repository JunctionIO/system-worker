package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestMsg() StatusMessage {
	return StatusMessage{
		TraceID:     "trace-uuid",
		LogID:       "log-uuid",
		Status:      "dispatched",
		AttemptedAt: "2026-07-02T10:00:00Z",
	}
}

func TestPostStatus_sendsCorrectPayload(t *testing.T) {
	var got map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewAPIClient(server.URL, "token").PostStatus(newTestMsg())

	if got["log_id"] != "log-uuid" {
		t.Errorf("log_id = %v, want log-uuid", got["log_id"])
	}
	if got["status"] != "dispatched" {
		t.Errorf("status = %v, want dispatched", got["status"])
	}
	if got["attempted_at"] != "2026-07-02T10:00:00Z" {
		t.Errorf("attempted_at = %v, want 2026-07-02T10:00:00Z", got["attempted_at"])
	}
}

func TestPostStatus_doesNotForwardTraceID(t *testing.T) {
	var got map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewAPIClient(server.URL, "token").PostStatus(newTestMsg())

	if _, ok := got["trace_id"]; ok {
		t.Error("trace_id should not be forwarded to the API")
	}
}

func TestPostStatus_setsContentTypeHeader(t *testing.T) {
	var contentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewAPIClient(server.URL, "token").PostStatus(newTestMsg())

	if contentType != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", contentType)
	}
}

func TestPostStatus_setsAuthHeader(t *testing.T) {
	var token string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = r.Header.Get("X-Junction-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewAPIClient(server.URL, "worker-jwt").PostStatus(newTestMsg())

	if token != "worker-jwt" {
		t.Errorf("X-Junction-Token = %v, want worker-jwt", token)
	}
}

func TestPostStatus_returnsNilOn2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := NewAPIClient(server.URL, "token").PostStatus(newTestMsg()); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestPostStatus_returnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	if err := NewAPIClient(server.URL, "token").PostStatus(newTestMsg()); err == nil {
		t.Error("expected error on 500, got nil")
	}
}

func TestPostStatus_returnsErrorOnNetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	if err := NewAPIClient(server.URL, "token").PostStatus(newTestMsg()); err == nil {
		t.Error("expected error on network failure, got nil")
	}
}
