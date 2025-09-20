package model

import (
	"time"
)

type Job struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Enabled       bool       `json:"enabled"`
	Type          string     `json:"type"`
	HTTP          HTTPConfig `json:"http"`
	Schedule      Schedule   `json:"schedule"`
	Timeout       Duration   `json:"timeout"`
	MaxRetries    int        `json:"max_retries"`
	RetryBackoff  Duration   `json:"retry_backoff"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	NextRunAt     *time.Time `json:"next_run_at,omitempty"`
	LastStatus    string     `json:"last_status"`
	LastError     string     `json:"last_error"`
}

type HTTPConfig struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Schedule struct {
	Kind    string     `json:"kind"`
	RunAt   *time.Time `json:"run_at,omitempty"`
	Every   Duration   `json:"every,omitempty"`
	StartAt *time.Time `json:"start_at,omitempty"`
	Jitter  Duration   `json:"jitter,omitempty"`
}

type JobStore struct {
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	Jobs      []Job     `json:"jobs"`
}

const (
	JobStatusSuccess = "success"
	JobStatusFailed  = "failed"
	JobStatusTimeout = "timeout"
	JobStatusSkipped = "skipped"
	JobStatusPaused  = "paused"
	JobStatusMissed  = "missed"
)

const (
	ScheduleKindOnce  = "once"
	ScheduleKindEvery = "every"
)

const (
	JobTypeHTTP = "http"
)

const (
	HTTPMethodGET  = "GET"
	HTTPMethodPOST = "POST"
)