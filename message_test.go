package main

import (
	"encoding/json"
	"testing"
)

func TestStatusMessage_unmarshalFields(t *testing.T) {
	raw := `{"trace_id":"trace-uuid","log_id":"log-uuid","status":"dispatched","attempted_at":"2026-07-02T10:00:00Z","error":null}`

	var msg StatusMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if msg.TraceID != "trace-uuid" {
		t.Errorf("TraceID = %v, want trace-uuid", msg.TraceID)
	}
	if msg.LogID != "log-uuid" {
		t.Errorf("LogID = %v, want log-uuid", msg.LogID)
	}
	if msg.Status != "dispatched" {
		t.Errorf("Status = %v, want dispatched", msg.Status)
	}
	if msg.AttemptedAt != "2026-07-02T10:00:00Z" {
		t.Errorf("AttemptedAt = %v, want 2026-07-02T10:00:00Z", msg.AttemptedAt)
	}
	if msg.Error != nil {
		t.Errorf("Error = %v, want nil", msg.Error)
	}
}

func TestStatusMessage_errorField(t *testing.T) {
	raw := `{"trace_id":"","log_id":"","status":"errored","attempted_at":"","error":"Connection refused"}`

	var msg StatusMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if msg.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if *msg.Error != "Connection refused" {
		t.Errorf("*Error = %v, want Connection refused", *msg.Error)
	}
}
