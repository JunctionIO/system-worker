# Junction System Worker

Consumes delivery status messages from the `junction.status` queue and calls the Junction API to update `event_log_destinations` records. Single point of API coupling for all status reporting.

## How it fits

Destination workers deliver payloads and publish a status message to `junction.status` for every attempt — success or failure. The system worker picks those messages up and calls `POST /system/status` on the API. This decouples delivery throughput from API availability: destination workers never block waiting for the API.

## Environment variables

| Variable | Description |
|---|---|
| `RABBITMQ_URL` | RabbitMQ connection string (`amqp://user:pass@host:5672/`) |
| `JUNCTION_API_URL` | Base URL of the Junction API (`http://api:8080`) |
| `JUNCTION_WORKER_TOKEN` | Worker JWT — generate with the Junction CLI |

Copy `.env.example` to `.env` and fill in the values before running locally.

## Local development

Requires the Junction API devenv to be running (provides RabbitMQ). Then:

```
cp .env.example .env
# fill in JUNCTION_WORKER_TOKEN
nix develop
go run .
```

Or start with devenv:

```
devenv up
```

## Testing

```
go test ./...
```

## Retry behaviour

Failed API calls are retried up to 3 times with exponential backoff starting at 5 seconds (5s, 10s). After all retries are exhausted the message is acknowledged and discarded — the stuck `event_log_destinations` record (status remains `pending`) serves as the recovery artifact and can be replayed manually once the API is healthy.

Malformed messages that cannot be unmarshalled are acknowledged and discarded immediately without calling the API.
