package enviar

import (
	"context"
	"errors"
	"testing"

	"github.com/LarnTechKe/work"
)

// --- test helpers ---

type testJob struct {
	name string
	opts []JobOption
}

func (j testJob) Name() string       { return j.name }
func (j testJob) Options() []JobOption { return j.opts }

type testHandler struct {
	job     Job
	called  bool
	lastCtx context.Context
	lastBody string
	err     error
}

func (h *testHandler) Job() Job { return h.job }

func (h *testHandler) PerformJob(ctx context.Context, body string) error {
	h.called = true
	h.lastCtx = ctx
	h.lastBody = body
	return h.err
}

// --- tests ---

func TestWrapJobHandler_CallsPerformJob(t *testing.T) {
	handler := &testHandler{
		job: testJob{name: "test_job", opts: nil},
	}

	wp := &workerPool{}
	wrapped := wp.wrapJobHandler(handler)

	job := &work.Job{
		Name: "test_job",
		ID:   "abc-123",
		Args: map[string]interface{}{
			BodyKey: `{"email":"test@example.com"}`,
		},
	}

	err := wrapped(job)
	if err != nil {
		t.Fatalf("wrapped() returned error: %v", err)
	}
	if !handler.called {
		t.Fatal("handler.PerformJob was not called")
	}
	if handler.lastBody != `{"email":"test@example.com"}` {
		t.Errorf("body = %q, want JSON string", handler.lastBody)
	}
}

func TestWrapJobHandler_NilBody(t *testing.T) {
	handler := &testHandler{
		job: testJob{name: "test_job"},
	}

	wp := &workerPool{}
	wrapped := wp.wrapJobHandler(handler)

	job := &work.Job{
		Name: "test_job",
		ID:   "abc-456",
		Args: map[string]interface{}{},
	}

	err := wrapped(job)
	if err != nil {
		t.Fatalf("wrapped() returned error: %v", err)
	}
	if handler.called {
		t.Error("handler should NOT have been called for nil body")
	}
}

func TestWrapJobHandler_EmptyBody(t *testing.T) {
	handler := &testHandler{
		job: testJob{name: "test_job"},
	}

	wp := &workerPool{}
	wrapped := wp.wrapJobHandler(handler)

	job := &work.Job{
		Name: "test_job",
		ID:   "abc-789",
		Args: map[string]interface{}{
			BodyKey: "",
		},
	}

	err := wrapped(job)
	if err != nil {
		t.Fatalf("wrapped() returned error: %v", err)
	}
	if handler.called {
		t.Error("handler should NOT have been called for empty body")
	}
}

func TestWrapJobHandler_NonStringBody(t *testing.T) {
	handler := &testHandler{
		job: testJob{name: "test_job"},
	}

	wp := &workerPool{}
	wrapped := wp.wrapJobHandler(handler)

	job := &work.Job{
		Name: "test_job",
		ID:   "abc-000",
		Args: map[string]interface{}{
			BodyKey: 12345, // not a string
		},
	}

	err := wrapped(job)
	if err != nil {
		t.Fatalf("wrapped() returned error: %v", err)
	}
	if handler.called {
		t.Error("handler should NOT have been called for non-string body")
	}
}

func TestWrapJobHandler_ReturnsHandlerError(t *testing.T) {
	expectedErr := errors.New("processing failed")
	handler := &testHandler{
		job: testJob{name: "fail_job"},
		err: expectedErr,
	}

	wp := &workerPool{}
	wrapped := wp.wrapJobHandler(handler)

	job := &work.Job{
		Name: "fail_job",
		ID:   "fail-123",
		Args: map[string]interface{}{
			BodyKey: `{"data":"value"}`,
		},
	}

	err := wrapped(job)
	if !errors.Is(err, expectedErr) {
		t.Errorf("wrapped() = %v, want %v", err, expectedErr)
	}
}

func TestWrapJobHandler_UsesPoolContext(t *testing.T) {
	handler := &testHandler{
		job: testJob{name: "ctx_job"},
	}

	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	wp := &workerPool{ctx: ctx}
	wrapped := wp.wrapJobHandler(handler)

	job := &work.Job{
		Name: "ctx_job",
		ID:   "ctx-123",
		Args: map[string]interface{}{
			BodyKey: `{"msg":"hello"}`,
		},
	}

	if err := wrapped(job); err != nil {
		t.Fatalf("wrapped() returned error: %v", err)
	}

	if handler.lastCtx != ctx {
		t.Error("handler did not receive the pool's context")
	}
	val, ok := handler.lastCtx.Value("test_key").(string)
	if !ok || val != "test_value" {
		t.Errorf("context value = %q, want %q", val, "test_value")
	}
}

func TestWrapJobHandler_NilContext_FallsBackToBackground(t *testing.T) {
	handler := &testHandler{
		job: testJob{name: "bg_job"},
	}

	wp := &workerPool{} // ctx is nil
	wrapped := wp.wrapJobHandler(handler)

	job := &work.Job{
		Name: "bg_job",
		ID:   "bg-123",
		Args: map[string]interface{}{
			BodyKey: `{"data":"ok"}`,
		},
	}

	if err := wrapped(job); err != nil {
		t.Fatalf("wrapped() returned error: %v", err)
	}

	if handler.lastCtx == nil {
		t.Fatal("handler received nil context")
	}
}

func TestTestJob_ImplementsInterfaces(t *testing.T) {
	// Compile-time check that testJob satisfies Job.
	var _ Job = testJob{}

	// Compile-time check that testHandler satisfies JobHandler.
	var _ JobHandler = &testHandler{}
}

func TestJobHandler_OptionsApplied(t *testing.T) {
	job := testJob{
		name: "priority_job",
		opts: []JobOption{
			WithMaxFails(2),
			WithMaxConcurrency(3),
			WithHighPriority(),
		},
	}

	opts := work.JobOptions{
		Priority: 1,
		MaxFails: 4,
	}
	for _, o := range job.Options() {
		o(&opts)
	}

	if opts.MaxFails != 2 {
		t.Errorf("MaxFails = %d, want 2", opts.MaxFails)
	}
	if opts.MaxConcurrency != 3 {
		t.Errorf("MaxConcurrency = %d, want 3", opts.MaxConcurrency)
	}
	if opts.Priority != 10 {
		t.Errorf("Priority = %d, want 10", opts.Priority)
	}
}
