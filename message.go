package main

type StatusMessage struct {
	TraceID     string  `json:"trace_id"`
	LogID       string  `json:"log_id"`
	Status      string  `json:"status"`
	AttemptedAt string  `json:"attempted_at"`
	Error       *string `json:"error"`
}
