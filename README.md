# enviar

[![Go Reference](https://pkg.go.dev/badge/github.com/LarnTechKe/enviar.svg)](https://pkg.go.dev/github.com/LarnTechKe/enviar)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**enviar** ("to send" in Spanish) is a lightweight Go wrapper around
[github.com/LarnTechKe/work](https://github.com/LarnTechKe/work) that provides
a clean, type-safe interface for enqueuing and processing Redis-backed background
jobs.

## Features

- **Simple enqueuing** — JSON payload marshaling handled automatically.
- **Delayed jobs** — schedule jobs to run after a duration.
- **Retry options** — per-job retry with exponential, linear, or fixed backoff.
- **Unique jobs** — deduplicate identical payloads.
- **Worker pool** — built-in middleware for logging and panic recovery.
- **Cron schedules** — register recurring jobs with cron expressions.
- **Environment-based config** — zero-config defaults with env var overrides.

## Installation

```bash
go get github.com/LarnTechKe/enviar
```

Requires **Go 1.24+** and a running Redis instance.

## Quick Start

### Define a Job

```go
package jobs

import "github.com/LarnTechKe/enviar"

type SendEmailJob struct{}

func (j SendEmailJob) Name() string { return "send_email" }

func (j SendEmailJob) Options() []enviar.JobOption {
    return []enviar.JobOption{
        enviar.WithMaxFails(3),
        enviar.WithMaxConcurrency(5),
    }
}
```

### Implement a Handler

```go
package handlers

import (
    "context"
    "encoding/json"

    "myapp/jobs"
)

type EmailHandler struct{}

func (h *EmailHandler) Job() enviar.Job { return jobs.SendEmailJob{} }

func (h *EmailHandler) PerformJob(ctx context.Context, body string) error {
    var req EmailRequest
    if err := json.Unmarshal([]byte(body), &req); err != nil {
        return err
    }
    // ... send the email ...
    return nil
}
```

### Enqueue a Job

```go
cfg := enviar.LoadEnv()
pool := cfg.NewPool()
defer pool.Close()

enqueuer := enviar.NewEnqueuer(pool, cfg.Namespace)

id, err := enqueuer.EnqueueBody("send_email", map[string]string{
    "to":      "alice@example.com",
    "subject": "Welcome!",
    "body":    "Thanks for signing up.",
})
```

### Enqueue with Delay

```go
id, err := enqueuer.EnqueueBodyIn("generate_report", 30*time.Second, reportRequest)
```

### Enqueue with Retry Options

```go
import "github.com/LarnTechKe/work"

id, err := enqueuer.EnqueueBodyWithRetry("webhook_delivery", payload, work.RetryOptions{
    MaxRetries: 5,
    Strategy:   work.BackoffExponential,
})
```

### Enqueue Unique Job

```go
id, err := enqueuer.EnqueueBodyUnique("process_payment", paymentRequest)
// id == "" when a duplicate is already enqueued
```

### Start Workers

```go
cfg := enviar.LoadEnv()
pool := cfg.NewPool()
defer pool.Close()

wp := enviar.NewWorkerPool(pool, cfg.Namespace, cfg.Concurrency)
wp.AddJobHandlers(
    &handlers.EmailHandler{},
    &handlers.PaymentHandler{},
)

// Optional: recurring jobs
wp.AddRecurringJobs(map[string]string{
    "*/20 * * * *": "generate_report",
})

// Blocks until SIGINT/SIGTERM
wp.Start(context.Background())
```

## Configuration

enviar reads from environment variables with sensible defaults:

| Variable | Default | Description |
|---|---|---|
| `ENVIAR_NAMESPACE` | `enviar` | Redis key namespace |
| `ENVIAR_REDIS_URL` | `localhost:6379` | `host:port` or `redis://:pass@host:port/db` |
| `ENVIAR_REDIS_DB` | `0` | Database number (plain host:port only) |
| `ENVIAR_CONCURRENCY` | `10` | Number of concurrent workers |

## Job Options

| Option | Description |
|---|---|
| `WithMaxConcurrency(n)` | Max concurrent instances of this job |
| `WithMaxFails(n)` | Max failures before dead-lettering |
| `WithPriority(p)` | Scheduling priority (1–10000) |
| `WithHighPriority()` | Shorthand for priority 10 |
| `WithLowPriority()` | Shorthand for priority 1 |
| `WithSkipDead(bool)` | Skip the dead queue on exhaustion |
| `WithBackoff(fn)` | Custom backoff calculator |

## Architecture

```
enviar
├── config.go        — Config, LoadEnv, NewPool
├── enqueuer.go      — Enqueuer (JSON-aware job producer)
├── job.go           — Job, JobHandler interfaces; JobOption helpers
└── worker_pool.go   — WorkerPool (consumer with middleware)
```

enviar wraps the low-level `work.Enqueuer` and `work.WorkerPool` types,
adding JSON payload marshaling, structured logging, panic recovery, and a
functional-options pattern for job configuration.

## License

[MIT](LICENSE) — same as the upstream `work` library.
