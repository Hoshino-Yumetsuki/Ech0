// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import migratorModel "github.com/lin-snow/ech0/internal/model/migrator"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSuccess   Status = "success"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

func (s Status) IsTerminal() bool {
	return s == StatusSuccess || s == StatusFailed || s == StatusCancelled
}

// Kind names a job type and binds it to its payload type P at compile time,
// so a runner and every Submit for the same kind must agree on P.
type Kind[P any] string

const (
	Reindex   Kind[struct{}]                       = "reindex"
	Migration Kind[migratorModel.MigrationPayload] = "migration"
	Export    Kind[migratorModel.ExportPayload]    = "export"
)

type Job struct {
	Type       string `gorm:"primaryKey;size:64"      json:"type"`
	Status     Status `gorm:"type:varchar(32);index"  json:"status"`
	Phase      string `gorm:"type:varchar(64)"        json:"phase"`
	Error      string `gorm:"type:text"               json:"error"`
	Payload    string `gorm:"type:text"               json:"payload"`
	StartedAt  *int64 `                               json:"started_at"`
	FinishedAt *int64 `                               json:"finished_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime"          json:"updated_at"`
}

func (Job) TableName() string {
	return "jobs"
}
