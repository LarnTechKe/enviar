package enviar

import (
	"context"

	"github.com/LarnTechKe/work"
)

// Job defines a background job type's name and worker-side options.
type Job interface {
	Name() string
	Options() []JobOption
}

// JobHandler processes jobs of a specific type.
// PerformJob receives the JSON-encoded body string from the job args.
type JobHandler interface {
	Job() Job
	PerformJob(ctx context.Context, body string) error
}

// JobOption is a functional option that configures a work.JobOptions value.
type JobOption func(*work.JobOptions)

// WithMaxConcurrency limits how many instances of this job may run at once.
func WithMaxConcurrency(n uint) JobOption {
	return func(o *work.JobOptions) { o.MaxConcurrency = n }
}

// WithMaxFails sets the maximum number of failures before the job is sent
// to the dead queue.
func WithMaxFails(n uint) JobOption {
	return func(o *work.JobOptions) { o.MaxFails = n }
}

// WithPriority sets the job's scheduling priority.
func WithPriority(p uint) JobOption {
	return func(o *work.JobOptions) { o.Priority = p }
}

// WithHighPriority is a convenience alias for WithPriority(10).
func WithHighPriority() JobOption { return WithPriority(10) }

// WithLowPriority is a convenience alias for WithPriority(1).
func WithLowPriority() JobOption { return WithPriority(1) }

// WithSkipDead controls whether failed jobs bypass the dead queue.
func WithSkipDead(skip bool) JobOption {
	return func(o *work.JobOptions) { o.SkipDead = skip }
}

// WithBackoff sets a custom backoff function for retries.
func WithBackoff(fn func(*work.Job) int64) JobOption {
	return func(o *work.JobOptions) { o.Backoff = fn }
}
