package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"ksana-service/internal/model"
	"ksana-service/internal/store"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type Executor interface {
	Execute(ctx context.Context, job *model.Job) error
}

type Scheduler struct {
	store    store.Store
	executor Executor
	clock    Clock
	jobHeap  *JobHeap
	jobIndex map[string]*JobItem
	timer    *time.Timer
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	logger   *slog.Logger
}

func NewScheduler(store store.Store, executor Executor, clock Clock, logger *slog.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		store:    store,
		executor: executor,
		clock:    clock,
		jobHeap:  NewJobHeap(),
		jobIndex: make(map[string]*JobItem),
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
	}
}

func (s *Scheduler) Start() error {
	jobStore, err := s.store.Load(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now()
	for i := range jobStore.Jobs {
		job := &jobStore.Jobs[i]
		if !job.Enabled {
			continue
		}

		if err := s.calculateNextRun(job, now); err != nil {
			s.logger.Error("Failed to calculate next run for job", "job_id", job.ID, "error", err)
			continue
		}

		if job.NextRunAt != nil {
			s.addJobToHeap(job, *job.NextRunAt)
		}
	}

	s.wg.Add(1)
	go s.schedulerLoop()

	return nil
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *Scheduler) AddJob(job *model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !job.Enabled {
		return nil
	}

	now := s.clock.Now()
	if err := s.calculateNextRun(job, now); err != nil {
		return err
	}

	if job.NextRunAt != nil {
		s.addJobToHeap(job, *job.NextRunAt)
		s.resetTimer()
	}

	return nil
}

func (s *Scheduler) UpdateJob(job *model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, exists := s.jobIndex[job.ID]; exists {
		s.jobHeap.remove(item)
		delete(s.jobIndex, job.ID)
	}

	if !job.Enabled {
		s.resetTimer()
		return nil
	}

	now := s.clock.Now()
	if err := s.calculateNextRun(job, now); err != nil {
		return err
	}

	if job.NextRunAt != nil {
		s.addJobToHeap(job, *job.NextRunAt)
		s.resetTimer()
	}

	return nil
}

func (s *Scheduler) RemoveJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, exists := s.jobIndex[jobID]; exists {
		s.jobHeap.remove(item)
		delete(s.jobIndex, jobID)
		s.resetTimer()
	}
}

func (s *Scheduler) RunNow(jobID string) error {
	job, err := s.store.Get(s.ctx, jobID)
	if err != nil {
		return err
	}

	go func() {
		if err := s.executor.Execute(s.ctx, job); err != nil {
			s.logger.Error("Failed to execute job", "job_id", job.ID, "error", err)
		}
	}()

	return nil
}

func (s *Scheduler) schedulerLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processReadyJobs()
		case <-s.getTimerChannel():
			s.processReadyJobs()
		}
	}
}

func (s *Scheduler) processReadyJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now()
	var readyJobs []*model.Job

	for s.jobHeap.Len() > 0 {
		item := (*s.jobHeap)[0]
		if item.RunTime.After(now) {
			break
		}

		item = heap.Pop(s.jobHeap).(*JobItem)
		delete(s.jobIndex, item.Job.ID)
		readyJobs = append(readyJobs, item.Job)
	}

	s.resetTimer()

	for _, job := range readyJobs {
		s.executeJob(job, now)
	}
}

func (s *Scheduler) executeJob(job *model.Job, scheduledTime time.Time) {
	go func() {
		if err := s.executor.Execute(s.ctx, job); err != nil {
			s.logger.Error("Failed to execute job", "job_id", job.ID, "error", err)
		}

		if job.Schedule.Kind == model.ScheduleKindEvery {
			s.scheduleNextRun(job, scheduledTime)
		}
	}()
}

func (s *Scheduler) scheduleNextRun(job *model.Job, lastScheduled time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now()
	nextRun := lastScheduled.Add(job.Schedule.Every.ToDuration())

	for nextRun.Before(now) || nextRun.Equal(now) {
		nextRun = nextRun.Add(job.Schedule.Every.ToDuration())
	}

	if job.Schedule.Jitter.ToDuration() > 0 {
		jitter := time.Duration(rand.Int63n(int64(job.Schedule.Jitter.ToDuration())))
		nextRun = nextRun.Add(jitter)
	}

	job.NextRunAt = &nextRun
	s.addJobToHeap(job, nextRun)
	s.resetTimer()

	if err := s.store.Put(s.ctx, job); err != nil {
		s.logger.Error("Failed to update job next run time", "job_id", job.ID, "error", err)
	}
}

func (s *Scheduler) calculateNextRun(job *model.Job, now time.Time) error {
	switch job.Schedule.Kind {
	case model.ScheduleKindOnce:
		if job.Schedule.RunAt == nil {
			return fmt.Errorf("run_at is required for once schedule")
		}
		if job.Schedule.RunAt.Before(now) {
			job.LastStatus = model.JobStatusMissed
			job.NextRunAt = nil
			return nil
		}
		job.NextRunAt = job.Schedule.RunAt

	case model.ScheduleKindEvery:
		if job.NextRunAt == nil {
			startAt := now
			if job.Schedule.StartAt != nil {
				startAt = *job.Schedule.StartAt
			}

			nextRun := startAt
			for nextRun.Before(now) || nextRun.Equal(now) {
				nextRun = nextRun.Add(job.Schedule.Every.ToDuration())
			}

			if job.Schedule.Jitter.ToDuration() > 0 {
				jitter := time.Duration(rand.Int63n(int64(job.Schedule.Jitter.ToDuration())))
				nextRun = nextRun.Add(jitter)
			}

			job.NextRunAt = &nextRun
		}
	}

	return nil
}

func (s *Scheduler) addJobToHeap(job *model.Job, runTime time.Time) {
	item := &JobItem{
		Job:     job,
		RunTime: runTime,
	}
	heap.Push(s.jobHeap, item)
	s.jobIndex[job.ID] = item
}

func (s *Scheduler) resetTimer() {
	if s.timer != nil {
		s.timer.Stop()
	}

	if s.jobHeap.Len() == 0 {
		return
	}

	next := (*s.jobHeap)[0].RunTime
	now := s.clock.Now()
	delay := next.Sub(now)

	if delay <= 0 {
		delay = time.Millisecond
	}

	s.timer = time.NewTimer(delay)
}

func (s *Scheduler) getTimerChannel() <-chan time.Time {
	if s.timer == nil {
		return nil
	}
	return s.timer.C
}