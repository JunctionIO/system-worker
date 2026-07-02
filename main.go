package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitmqURL := mustEnv("RABBITMQ_URL")
	apiURL := mustEnv("JUNCTION_API_URL")
	workerToken := mustEnv("JUNCTION_WORKER_TOKEN")

	conn, err := amqp.Dial(rabbitmqURL)

	if err != nil {
		slog.Error("failed to connect to rabbitmq", "error", err)
		os.Exit(1)
	}

	defer conn.Close()

	client := NewAPIClient(apiURL, workerToken)
	consumer := NewConsumer(conn, client)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errc := make(chan error, 1)

	go func() {
		errc <- consumer.Run()
	}()

	select {
	case <-quit:
		slog.Info("shutting down")
		conn.Close()

		if err := <-errc; err != nil {
			slog.Error("consumer error on shutdown", "error", err)
		}
	case err := <-errc:
		if err != nil {
			slog.Error("consumer exited unexpectedly", "error", err)
			os.Exit(1)
		}
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)

	if v == "" {
		slog.Error("missing required environment variable", "key", key)
		os.Exit(1)
	}

	return v
}
