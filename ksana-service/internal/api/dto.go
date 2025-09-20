package api

import (
	"ksana-service/internal/model"
	"time"
)

type CreateJobRequest struct {
	Name         string              `json:"name"`
	Enabled      *bool               `json:"enabled,omitempty"`
	Type         string              `json:"type"`
	HTTP         model.HTTPConfig    `json:"http"`
	Schedule     model.Schedule      `json:"schedule"`
	Timeout      model.Duration      `json:"timeout,omitempty"`
	MaxRetries   *int                `json:"max_retries,omitempty"`
	RetryBackoff model.Duration      `json:"retry_backoff,omitempty"`
}

type UpdateJobRequest struct {
	Name         *string             `json:"name,omitempty"`
	Enabled      *bool               `json:"enabled,omitempty"`
	HTTP         *model.HTTPConfig   `json:"http,omitempty"`
	Schedule     *model.Schedule     `json:"schedule,omitempty"`
	Timeout      *model.Duration     `json:"timeout,omitempty"`
	MaxRetries   *int                `json:"max_retries,omitempty"`
	RetryBackoff *model.Duration     `json:"retry_backoff,omitempty"`
}

type JobResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Enabled       bool                   `json:"enabled"`
	Type          string                 `json:"type"`
	HTTP          model.HTTPConfig       `json:"http"`
	Schedule      model.Schedule         `json:"schedule"`
	Timeout       model.Duration         `json:"timeout"`
	MaxRetries    int                    `json:"max_retries"`
	RetryBackoff  model.Duration         `json:"retry_backoff"`
	LastRunAt     *time.Time             `json:"last_run_at,omitempty"`
	NextRunAt     *time.Time             `json:"next_run_at,omitempty"`
	LastStatus    string                 `json:"last_status"`
	LastError     string                 `json:"last_error"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func (r *CreateJobRequest) ToJob() *model.Job {
	job := &model.Job{
		Name:         r.Name,
		Type:         r.Type,
		HTTP:         r.HTTP,
		Schedule:     r.Schedule,
		Timeout:      r.Timeout,
		MaxRetries:   0,
		RetryBackoff: r.RetryBackoff,
	}

	if r.Enabled != nil {
		job.Enabled = *r.Enabled
	}

	if r.MaxRetries != nil {
		job.MaxRetries = *r.MaxRetries
	}

	job.SetDefaults()
	return job
}

func JobToResponse(job *model.Job) JobResponse {
	return JobResponse{
		ID:            job.ID,
		Name:          job.Name,
		Enabled:       job.Enabled,
		Type:          job.Type,
		HTTP:          job.HTTP,
		Schedule:      job.Schedule,
		Timeout:       job.Timeout,
		MaxRetries:    job.MaxRetries,
		RetryBackoff:  job.RetryBackoff,
		LastRunAt:     job.LastRunAt,
		NextRunAt:     job.NextRunAt,
		LastStatus:    job.LastStatus,
		LastError:     job.LastError,
	}
}