package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"ksana-service/internal/model"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JSONStore struct {
	dataDir string
	mu      sync.RWMutex
	data    *model.JobStore
	jobMap  map[string]*model.Job
}

func NewJSONStore(dataDir string) *JSONStore {
	return &JSONStore{
		dataDir: dataDir,
		data:    nil,
		jobMap:  make(map[string]*model.Job),
	}
}

func (s *JSONStore) Load(ctx context.Context) (*model.JobStore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := filepath.Join(s.dataDir, "jobs.json")

	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		s.data = &model.JobStore{
			Version:   1,
			UpdatedAt: time.Now().UTC(),
			Jobs:      []model.Job{},
		}
		s.jobMap = make(map[string]*model.Job)
		return s.data, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read jobs file: %w", err)
	}

	var jobStore model.JobStore
	if err := json.Unmarshal(data, &jobStore); err != nil {
		backupPath := filepath.Join(s.dataDir, fmt.Sprintf("jobs.bad.%d.json", time.Now().Unix()))
		os.Rename(filePath, backupPath)

		s.data = &model.JobStore{
			Version:   1,
			UpdatedAt: time.Now().UTC(),
			Jobs:      []model.Job{},
		}
		s.jobMap = make(map[string]*model.Job)
		return s.data, nil
	}

	s.data = &jobStore
	s.rebuildJobMap()
	return s.data, nil
}

func (s *JSONStore) Save(ctx context.Context, jobStore *model.JobStore) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobStore.UpdatedAt = time.Now().UTC()
	s.data = jobStore
	s.rebuildJobMap()

	return s.atomicWrite(jobStore)
}

func (s *JSONStore) List(ctx context.Context) ([]model.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil {
		return nil, fmt.Errorf("store not loaded")
	}

	jobs := make([]model.Job, len(s.data.Jobs))
	copy(jobs, s.data.Jobs)
	return jobs, nil
}

func (s *JSONStore) Get(ctx context.Context, id string) (*model.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobMap[id]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}

	jobCopy := *job
	return &jobCopy, nil
}

func (s *JSONStore) Put(ctx context.Context, job *model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		return fmt.Errorf("store not loaded")
	}

	if job.ID == "" {
		job.ID = s.generateID()
	}

	_, exists := s.jobMap[job.ID]
	if exists {
		for i, j := range s.data.Jobs {
			if j.ID == job.ID {
				s.data.Jobs[i] = *job
				break
			}
		}
	} else {
		s.data.Jobs = append(s.data.Jobs, *job)
	}

	s.rebuildJobMap()
	return s.atomicWrite(s.data)
}

func (s *JSONStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		return fmt.Errorf("store not loaded")
	}

	_, exists := s.jobMap[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	var newJobs []model.Job
	for _, job := range s.data.Jobs {
		if job.ID != id {
			newJobs = append(newJobs, job)
		}
	}

	s.data.Jobs = newJobs
	s.rebuildJobMap()
	return s.atomicWrite(s.data)
}

func (s *JSONStore) rebuildJobMap() {
	s.jobMap = make(map[string]*model.Job)
	for i := range s.data.Jobs {
		s.jobMap[s.data.Jobs[i].ID] = &s.data.Jobs[i]
	}
}

func (s *JSONStore) atomicWrite(jobStore *model.JobStore) error {
	data, err := json.MarshalIndent(jobStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal job store: %w", err)
	}

	filePath := filepath.Join(s.dataDir, "jobs.json")
	tmpPath := filePath + ".tmp"

	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = tmpFile.Write(data)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	tmpFile.Close()

	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (s *JSONStore) generateID() string {
	bytes := make([]byte, 16)
	io.ReadFull(rand.Reader, bytes)
	return hex.EncodeToString(bytes)
}