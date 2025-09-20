package model

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

func (j *Job) Validate() error {
	if j.Name == "" {
		return errors.New("job name is required")
	}

	if j.Type != JobTypeHTTP {
		return errors.New("job type must be 'http'")
	}

	if err := j.HTTP.Validate(); err != nil {
		return err
	}

	if err := j.Schedule.Validate(); err != nil {
		return err
	}

	if j.Timeout.ToDuration() <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if j.MaxRetries < 0 {
		return errors.New("max_retries must be non-negative")
	}

	if j.RetryBackoff.ToDuration() < 0 {
		return errors.New("retry_backoff must be non-negative")
	}

	return nil
}

func (h *HTTPConfig) Validate() error {
	if h.Method == "" {
		return errors.New("HTTP method is required")
	}

	method := strings.ToUpper(h.Method)
	if method != HTTPMethodGET && method != HTTPMethodPOST {
		return errors.New("HTTP method must be GET or POST")
	}

	if h.URL == "" {
		return errors.New("HTTP URL is required")
	}

	_, err := url.Parse(h.URL)
	if err != nil {
		return errors.New("invalid HTTP URL")
	}

	return nil
}

func (s *Schedule) Validate() error {
	if s.Kind != ScheduleKindOnce && s.Kind != ScheduleKindEvery {
		return errors.New("schedule kind must be 'once' or 'every'")
	}

	switch s.Kind {
	case ScheduleKindOnce:
		if s.RunAt == nil {
			return errors.New("run_at is required for 'once' schedule")
		}
		if s.RunAt.Before(time.Now().UTC()) {
			return errors.New("run_at must be in the future")
		}
	case ScheduleKindEvery:
		if s.Every.ToDuration() <= 0 {
			return errors.New("every must be greater than 0 for 'every' schedule")
		}
		if s.Jitter.ToDuration() < 0 {
			return errors.New("jitter must be non-negative")
		}
	}

	return nil
}

func (j *Job) SetDefaults() {
	if j.Type == "" {
		j.Type = JobTypeHTTP
	}

	if j.HTTP.Method == "" {
		j.HTTP.Method = HTTPMethodPOST
	}

	if j.HTTP.Headers == nil {
		j.HTTP.Headers = make(map[string]string)
	}

	if j.Timeout.ToDuration() == 0 {
		j.Timeout = DurationFromTimeDuration(10 * time.Second)
	}

	if j.MaxRetries == 0 {
		j.MaxRetries = 3
	}

	if j.RetryBackoff.ToDuration() == 0 {
		j.RetryBackoff = DurationFromTimeDuration(5 * time.Second)
	}

	j.Enabled = true
}