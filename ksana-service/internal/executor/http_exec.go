package executor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"ksana-service/internal/model"
	"ksana-service/internal/store"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HTTPExecutor struct {
	client     *http.Client
	store      store.Store
	workerPool chan struct{}
	wg         sync.WaitGroup
	logger     *slog.Logger
}

func NewHTTPExecutor(workers int, timeout time.Duration, store store.Store, logger *slog.Logger) *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		store:      store,
		workerPool: make(chan struct{}, workers),
		logger:     logger,
	}
}

func (e *HTTPExecutor) Execute(ctx context.Context, job *model.Job) error {
	e.workerPool <- struct{}{}
	defer func() { <-e.workerPool }()

	e.wg.Add(1)
	defer e.wg.Done()

	runID := e.generateRunID()
	startTime := time.Now().UTC()

	e.logger.Info("Starting job execution",
		"job_id", job.ID,
		"job_name", job.Name,
		"run_id", runID)

	var lastErr error
	for attempt := 0; attempt <= job.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := job.RetryBackoff.ToDuration()
			e.logger.Info("Retrying job execution",
				"job_id", job.ID,
				"attempt", attempt,
				"backoff", backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		execCtx, cancel := context.WithTimeout(ctx, job.Timeout.ToDuration())

		err := e.executeHTTPRequest(execCtx, job, runID, startTime)
		cancel()

		if err == nil {
			e.updateJobStatus(job, model.JobStatusSuccess, "", startTime)
			e.logger.Info("Job executed successfully",
				"job_id", job.ID,
				"run_id", runID,
				"latency_ms", time.Since(startTime).Milliseconds())
			return nil
		}

		lastErr = err

		if !e.isRetryableError(err) {
			break
		}
	}

	status := model.JobStatusFailed
	if e.isTimeoutError(lastErr) {
		status = model.JobStatusTimeout
	}

	e.updateJobStatus(job, status, lastErr.Error(), startTime)
	e.logger.Error("Job execution failed",
		"job_id", job.ID,
		"run_id", runID,
		"error", lastErr,
		"latency_ms", time.Since(startTime).Milliseconds())

	return lastErr
}

func (e *HTTPExecutor) executeHTTPRequest(ctx context.Context, job *model.Job, runID string, triggeredAt time.Time) error {
	req, err := http.NewRequestWithContext(ctx, job.HTTP.Method, job.HTTP.URL, strings.NewReader(job.HTTP.Body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range job.HTTP.Headers {
		req.Header.Set(key, value)
	}

	req.Header.Set("X-Ksana-Run-Id", runID)

	if job.HTTP.Method == "POST" && req.Header.Get("Content-Type") == "" {
		if job.HTTP.Body != "" && (strings.HasPrefix(strings.TrimSpace(job.HTTP.Body), "{") || strings.HasPrefix(strings.TrimSpace(job.HTTP.Body), "[")) {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "text/plain")
		}
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
}

func (e *HTTPExecutor) updateJobStatus(job *model.Job, status, errorMsg string, runTime time.Time) {
	job.LastRunAt = &runTime
	job.LastStatus = status
	job.LastError = errorMsg

	if err := e.store.Put(context.Background(), job); err != nil {
		e.logger.Error("Failed to update job status", "job_id", job.ID, "error", err)
	}
}

func (e *HTTPExecutor) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
		return true
	}

	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") {
		return true
	}

	if strings.Contains(errStr, "status 408") || strings.Contains(errStr, "status 429") || strings.Contains(errStr, "status 5") {
		return true
	}

	return false
}

func (e *HTTPExecutor) isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded")
}

func (e *HTTPExecutor) generateRunID() string {
	bytes := make([]byte, 16)
	io.ReadFull(rand.Reader, bytes)
	return hex.EncodeToString(bytes)
}

func (e *HTTPExecutor) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}