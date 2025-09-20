package api

import (
	"encoding/json"
	"fmt"
	"ksana-service/internal/model"
	"ksana-service/internal/store"
	"log/slog"
	"net/http"
	"strings"
)

type SchedulerService interface {
	AddJob(job *model.Job) error
	UpdateJob(job *model.Job) error
	RemoveJob(jobID string)
	RunNow(jobID string) error
}

type JobHandler struct {
	store     store.Store
	scheduler SchedulerService
	logger    *slog.Logger
}

func NewJobHandler(store store.Store, scheduler SchedulerService, logger *slog.Logger) *JobHandler {
	return &JobHandler{
		store:     store,
		scheduler: scheduler,
		logger:    logger,
	}
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	job := req.ToJob()
	if err := job.Validate(); err != nil {
		h.writeError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	if err := h.store.Put(r.Context(), job); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to save job", err.Error())
		return
	}

	if err := h.scheduler.AddJob(job); err != nil {
		h.logger.Error("Failed to add job to scheduler", "job_id", job.ID, "error", err)
	}

	h.writeJSON(w, http.StatusCreated, JobToResponse(job))
}

func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.store.List(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list jobs", err.Error())
		return
	}

	var responses []JobResponse
	for _, job := range jobs {
		responses = append(responses, JobToResponse(&job))
	}

	h.writeJSON(w, http.StatusOK, responses)
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	jobID := h.extractJobID(r)
	if jobID == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID", "")
		return
	}

	job, err := h.store.Get(r.Context(), jobID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, JobToResponse(job))
}

func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	jobID := h.extractJobID(r)
	if jobID == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID", "")
		return
	}

	job, err := h.store.Get(r.Context(), jobID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	var req UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	h.applyJobUpdates(job, &req)

	if err := job.Validate(); err != nil {
		h.writeError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	if err := h.store.Put(r.Context(), job); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update job", err.Error())
		return
	}

	if err := h.scheduler.UpdateJob(job); err != nil {
		h.logger.Error("Failed to update job in scheduler", "job_id", job.ID, "error", err)
	}

	h.writeJSON(w, http.StatusOK, JobToResponse(job))
}

func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	jobID := h.extractJobID(r)
	if jobID == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID", "")
		return
	}

	if err := h.store.Delete(r.Context(), jobID); err != nil {
		h.writeError(w, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	h.scheduler.RemoveJob(jobID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *JobHandler) RunNow(w http.ResponseWriter, r *http.Request) {
	jobID := h.extractJobID(r)
	if jobID == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID", "")
		return
	}

	if err := h.scheduler.RunNow(jobID); err != nil {
		h.writeError(w, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job triggered successfully"})
}

func (h *JobHandler) PauseJob(w http.ResponseWriter, r *http.Request) {
	h.setJobEnabled(w, r, false)
}

func (h *JobHandler) ResumeJob(w http.ResponseWriter, r *http.Request) {
	h.setJobEnabled(w, r, true)
}

func (h *JobHandler) setJobEnabled(w http.ResponseWriter, r *http.Request, enabled bool) {
	jobID := h.extractJobID(r)
	if jobID == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID", "")
		return
	}

	job, err := h.store.Get(r.Context(), jobID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	job.Enabled = enabled

	if err := h.store.Put(r.Context(), job); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update job", err.Error())
		return
	}

	if err := h.scheduler.UpdateJob(job); err != nil {
		h.logger.Error("Failed to update job in scheduler", "job_id", job.ID, "error", err)
	}

	action := "paused"
	if enabled {
		action = "resumed"
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("Job %s successfully", action)})
}

func (h *JobHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *JobHandler) applyJobUpdates(job *model.Job, req *UpdateJobRequest) {
	if req.Name != nil {
		job.Name = *req.Name
	}
	if req.Enabled != nil {
		job.Enabled = *req.Enabled
	}
	if req.HTTP != nil {
		job.HTTP = *req.HTTP
	}
	if req.Schedule != nil {
		job.Schedule = *req.Schedule
		job.NextRunAt = nil
	}
	if req.Timeout != nil {
		job.Timeout = *req.Timeout
	}
	if req.MaxRetries != nil {
		job.MaxRetries = *req.MaxRetries
	}
	if req.RetryBackoff != nil {
		job.RetryBackoff = *req.RetryBackoff
	}
}

func (h *JobHandler) extractJobID(r *http.Request) string {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "jobs" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func (h *JobHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *JobHandler) writeError(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   error,
		Message: message,
	})
}