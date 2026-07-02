package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type mockAcknowledger struct {
	ackCount atomic.Int32
}

func (m *mockAcknowledger) Ack(tag uint64, multiple bool) error {
	m.ackCount.Add(1)
	return nil
}

func (m *mockAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error { return nil }
func (m *mockAcknowledger) Reject(tag uint64, requeue bool) error              { return nil }

func makeDelivery(ack *mockAcknowledger, body []byte) amqp.Delivery {
	return amqp.Delivery{Acknowledger: ack, Body: body}
}

func validBody() []byte {
	return []byte(`{"trace_id":"trace-uuid","log_id":"log-uuid","status":"dispatched","attempted_at":"2026-07-02T10:00:00Z","error":null}`)
}

func TestProcess_acksOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ack := &mockAcknowledger{}
	consumer := &Consumer{client: NewAPIClient(server.URL, "token")}

	consumer.process(makeDelivery(ack, validBody()))

	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
}

func TestProcess_acksAndDiscardsOnMalformedJSON(t *testing.T) {
	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ack := &mockAcknowledger{}
	consumer := &Consumer{client: NewAPIClient(server.URL, "token")}

	consumer.process(makeDelivery(ack, []byte("not json")))

	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
	if apiCalled {
		t.Error("API should not be called for a malformed message")
	}
}

func TestProcess_retriesOnFailureThenAcksOnSuccess(t *testing.T) {
	old := baseBackoff
	baseBackoff = 0
	defer func() { baseBackoff = old }()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ack := &mockAcknowledger{}
	consumer := &Consumer{client: NewAPIClient(server.URL, "token")}

	consumer.process(makeDelivery(ack, validBody()))

	if n := requests.Load(); n != 2 {
		t.Errorf("API called %d times, want 2", n)
	}
	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
}

func TestProcess_acksAfterAllRetriesExhausted(t *testing.T) {
	old := baseBackoff
	baseBackoff = 0
	defer func() { baseBackoff = old }()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ack := &mockAcknowledger{}
	consumer := &Consumer{client: NewAPIClient(server.URL, "token")}

	consumer.process(makeDelivery(ack, validBody()))

	if n := requests.Load(); n != maxRetries {
		t.Errorf("API called %d times, want %d", n, maxRetries)
	}
	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
}
