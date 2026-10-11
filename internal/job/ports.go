// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package job

import (
	"context"
	"errors"

	jobModel "github.com/lin-snow/ech0/internal/model/job"
)

var ErrNotFound = errors.New("job not found")

type ReportFunc func(phase string, snapshot any)

// RunFunc executes one job of a kind whose payload type is P.
type RunFunc[P any] func(ctx context.Context, p P, report ReportFunc) (result any, err error)

type JobRepository interface {
	Upsert(ctx context.Context, j *jobModel.Job) error
	GetByType(ctx context.Context, jobType string) (jobModel.Job, error)
	SweepRunning(ctx context.Context, reason string) error
	Delete(ctx context.Context, jobType string) error
}

type Progress struct {
	Phase    string
	Snapshot any
}
