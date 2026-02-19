package enviar

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/LarnTechKe/work"
	"github.com/gomodule/redigo/redis"
)

// WorkerPool manages job handlers, middleware, and periodic schedules.
type WorkerPool interface {
	AddJobHandlers(handlers ...JobHandler)
	AddRecurringJobs(cronTaskMap map[string]string)
	Start(ctx context.Context)
	Stop()
}

type workerPool struct {
	ctx  context.Context
	pool *work.WorkerPool
}

type workerPoolContext struct{}

// NewWorkerPool creates a pool with logging and panic-recovery middleware.
func NewWorkerPool(redisPool *redis.Pool, namespace string, concurrency uint) WorkerPool {
	pool := work.NewWorkerPool(workerPoolContext{}, concurrency, namespace, redisPool)
	pool.Middleware(logMiddleware)
	pool.Middleware(panicRecovery)
	return &workerPool{pool: pool}
}

// AddJobHandlers registers one or more job handlers with the pool.
func (wp *workerPool) AddJobHandlers(handlers ...JobHandler) {
	for _, h := range handlers {
		job := h.Job()

		opts := work.JobOptions{
			Priority: 1,
			MaxFails: 4,
		}
		for _, opt := range job.Options() {
			opt(&opts)
		}

		wrapped := wp.wrapJobHandler(h)
		wp.pool.JobWithOptions(job.Name(), opts, wrapped)
	}
}

// AddRecurringJobs registers cron-scheduled periodic jobs.
func (wp *workerPool) AddRecurringJobs(cronTaskMap map[string]string) {
	for spec, jobName := range cronTaskMap {
		wp.pool.PeriodicallyEnqueue(spec, jobName)
	}
}

// Start begins processing and blocks until SIGINT or SIGTERM is received.
func (wp *workerPool) Start(ctx context.Context) {
	wp.ctx = ctx
	wp.pool.Start()
	log.Printf("[enviar] worker pool started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("[enviar] shutting down...")
	wp.pool.Stop()
	log.Println("[enviar] stopped")
}

// Stop gracefully shuts down the pool.
func (wp *workerPool) Stop() {
	wp.pool.Stop()
}

// wrapJobHandler extracts the body from job args and delegates to PerformJob.
func (wp *workerPool) wrapJobHandler(handler JobHandler) func(job *work.Job) error {
	return func(job *work.Job) error {
		start := time.Now()

		jobCtx := wp.ctx
		if jobCtx == nil {
			jobCtx = context.Background()
		}

		rawBody := job.Args[BodyKey]
		if rawBody == nil {
			log.Printf("[%s] id=%s — no body, skipping", job.Name, job.ID)
			return nil
		}

		body, ok := rawBody.(string)
		if !ok {
			log.Printf("[%s] id=%s — body is not a string, skipping", job.Name, job.ID)
			return nil
		}

		if len(body) == 0 {
			log.Printf("[%s] id=%s — empty body, skipping", job.Name, job.ID)
			return nil
		}

		err := handler.PerformJob(jobCtx, body)
		dur := time.Since(start)

		if err != nil {
			log.Printf("[%s] id=%s FAILED in %s: %v", job.Name, job.ID, dur, err)
		} else {
			log.Printf("[%s] id=%s completed in %s", job.Name, job.ID, dur)
		}

		return err
	}
}

// --- middleware ---

func logMiddleware(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Printf("[middleware] starting job_id=%s job_name=%s", job.ID, job.Name)
	return next()
}

func panicRecovery(job *work.Job, next work.NextMiddlewareFunc) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] job_id=%s job_name=%s: %v\n%s",
				job.ID, job.Name, r, debug.Stack())
			fmt.Fprintf(os.Stderr, "[PANIC] recovered: %v\n", r)
		}
	}()
	return next()
}
