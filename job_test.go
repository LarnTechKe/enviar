package enviar

import (
	"testing"

	"github.com/LarnTechKe/work"
)

func TestWithMaxConcurrency(t *testing.T) {
	var opts work.JobOptions
	WithMaxConcurrency(5)(&opts)
	if opts.MaxConcurrency != 5 {
		t.Errorf("MaxConcurrency = %d, want 5", opts.MaxConcurrency)
	}
}

func TestWithMaxFails(t *testing.T) {
	var opts work.JobOptions
	WithMaxFails(3)(&opts)
	if opts.MaxFails != 3 {
		t.Errorf("MaxFails = %d, want 3", opts.MaxFails)
	}
}

func TestWithPriority(t *testing.T) {
	var opts work.JobOptions
	WithPriority(7)(&opts)
	if opts.Priority != 7 {
		t.Errorf("Priority = %d, want 7", opts.Priority)
	}
}

func TestWithHighPriority(t *testing.T) {
	var opts work.JobOptions
	WithHighPriority()(&opts)
	if opts.Priority != 10 {
		t.Errorf("Priority = %d, want 10", opts.Priority)
	}
}

func TestWithLowPriority(t *testing.T) {
	var opts work.JobOptions
	WithLowPriority()(&opts)
	if opts.Priority != 1 {
		t.Errorf("Priority = %d, want 1", opts.Priority)
	}
}

func TestWithSkipDead(t *testing.T) {
	var opts work.JobOptions
	WithSkipDead(true)(&opts)
	if !opts.SkipDead {
		t.Error("SkipDead should be true")
	}

	WithSkipDead(false)(&opts)
	if opts.SkipDead {
		t.Error("SkipDead should be false")
	}
}

func TestWithBackoff(t *testing.T) {
	fn := func(j *work.Job) int64 { return 42 }
	var opts work.JobOptions
	WithBackoff(fn)(&opts)
	if opts.Backoff == nil {
		t.Fatal("Backoff should not be nil")
	}
	job := &work.Job{}
	if got := opts.Backoff(job); got != 42 {
		t.Errorf("Backoff(job) = %d, want 42", got)
	}
}

func TestJobOptions_Composable(t *testing.T) {
	options := []JobOption{
		WithMaxConcurrency(3),
		WithMaxFails(5),
		WithHighPriority(),
		WithSkipDead(true),
	}

	var opts work.JobOptions
	for _, o := range options {
		o(&opts)
	}

	if opts.MaxConcurrency != 3 {
		t.Errorf("MaxConcurrency = %d, want 3", opts.MaxConcurrency)
	}
	if opts.MaxFails != 5 {
		t.Errorf("MaxFails = %d, want 5", opts.MaxFails)
	}
	if opts.Priority != 10 {
		t.Errorf("Priority = %d, want 10", opts.Priority)
	}
	if !opts.SkipDead {
		t.Error("SkipDead should be true")
	}
}
