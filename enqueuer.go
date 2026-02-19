package enviar

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LarnTechKe/work"
	"github.com/gomodule/redigo/redis"
)

// BodyKey is the args key used to carry the JSON-encoded payload.
const BodyKey = "body"

// Enqueuer wraps work.Enqueuer with a simpler, JSON-centric interface.
type Enqueuer struct {
	inner *work.Enqueuer
}

// NewEnqueuer creates an Enqueuer backed by the given Redis pool and namespace.
func NewEnqueuer(pool *redis.Pool, namespace string) *Enqueuer {
	return &Enqueuer{inner: work.NewEnqueuer(namespace, pool)}
}

// EnqueueBody marshals payload to JSON and enqueues an immediate job.
// It returns the new job's ID.
func (e *Enqueuer) EnqueueBody(jobName string, payload interface{}) (string, error) {
	args, err := bodyArgs(payload)
	if err != nil {
		return "", fmt.Errorf("enqueue %s: %w", jobName, err)
	}
	job, err := e.inner.Enqueue(jobName, args)
	if err != nil {
		return "", fmt.Errorf("enqueue %s: %w", jobName, err)
	}
	return job.ID, nil
}

// EnqueueBodyIn marshals payload to JSON and enqueues with a delay.
func (e *Enqueuer) EnqueueBodyIn(jobName string, delay time.Duration, payload interface{}) (string, error) {
	args, err := bodyArgs(payload)
	if err != nil {
		return "", fmt.Errorf("enqueue_in %s: %w", jobName, err)
	}
	sj, err := e.inner.EnqueueIn(jobName, int64(delay.Seconds()), args)
	if err != nil {
		return "", fmt.Errorf("enqueue_in %s: %w", jobName, err)
	}
	return sj.Job.ID, nil
}

// EnqueueBodyWithRetry marshals payload and enqueues with retry options.
func (e *Enqueuer) EnqueueBodyWithRetry(jobName string, payload interface{}, opts work.RetryOptions) (string, error) {
	args, err := bodyArgs(payload)
	if err != nil {
		return "", fmt.Errorf("enqueue_with_retry %s: %w", jobName, err)
	}
	job, err := e.inner.EnqueueWithOptions(jobName, args, work.EnqueueOptions{
		Retry: &opts,
	})
	if err != nil {
		return "", fmt.Errorf("enqueue_with_retry %s: %w", jobName, err)
	}
	return job.ID, nil
}

// EnqueueBodyDelayedWithRetry marshals payload and enqueues with both
// a delay and retry options.
func (e *Enqueuer) EnqueueBodyDelayedWithRetry(jobName string, delay time.Duration, payload interface{}, opts work.RetryOptions) (string, error) {
	args, err := bodyArgs(payload)
	if err != nil {
		return "", fmt.Errorf("enqueue_delayed_with_retry %s: %w", jobName, err)
	}
	job, err := e.inner.EnqueueWithOptions(jobName, args, work.EnqueueOptions{
		Delay: delay,
		Retry: &opts,
	})
	if err != nil {
		return "", fmt.Errorf("enqueue_delayed_with_retry %s: %w", jobName, err)
	}
	return job.ID, nil
}

// EnqueueBodyUnique marshals payload and enqueues only if no identical job
// is already queued. Returns ("", nil) when a duplicate is detected.
func (e *Enqueuer) EnqueueBodyUnique(jobName string, payload interface{}) (string, error) {
	args, err := bodyArgs(payload)
	if err != nil {
		return "", fmt.Errorf("enqueue_unique %s: %w", jobName, err)
	}
	job, err := e.inner.EnqueueUnique(jobName, args)
	if err != nil {
		return "", fmt.Errorf("enqueue_unique %s: %w", jobName, err)
	}
	if job == nil {
		return "", nil // already enqueued
	}
	return job.ID, nil
}

// bodyArgs marshals v to JSON and wraps it in work.Q{"body": jsonString}.
func bodyArgs(v interface{}) (work.Q, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	return work.Q{BodyKey: string(data)}, nil
}
