package store

import (
	"context"
	"ksana-service/internal/model"
)

type Store interface {
	Load(ctx context.Context) (*model.JobStore, error)
	Save(ctx context.Context, jobStore *model.JobStore) error
	List(ctx context.Context) ([]model.Job, error)
	Get(ctx context.Context, id string) (*model.Job, error)
	Put(ctx context.Context, job *model.Job) error
	Delete(ctx context.Context, id string) error
}