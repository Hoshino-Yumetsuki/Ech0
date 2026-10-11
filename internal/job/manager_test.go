// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package job_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/job"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
)

const kind jobModel.Kind[struct{}] = "t"

type stubRepo struct {
	mu   sync.Mutex
	rows map[string]jobModel.Job
}

func newStubRepo() *stubRepo { return &stubRepo{rows: map[string]jobModel.Job{}} }

func (r *stubRepo) Upsert(_ context.Context, j *jobModel.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[j.Type] = *j
	return nil
}

func (r *stubRepo) GetByType(_ context.Context, t string) (jobModel.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.rows[t]
	if !ok {
		return jobModel.Job{}, job.ErrNotFound
	}
	return j, nil
}

func (r *stubRepo) SweepRunning(_ context.Context, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, j := range r.rows {
		if !j.Status.IsTerminal() {
			j.Status = jobModel.StatusFailed
			j.Error = reason
			r.rows[k] = j
		}
	}
	return nil
}

func (r *stubRepo) Delete(_ context.Context, t string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rows, t)
	return nil
}

func waitForStatus(t *testing.T, mgr *job.Manager, want jobModel.Status) jobModel.Job {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		jb, err := mgr.Get(context.Background(), kind)
		if err == nil && jb.Status == want {
			return jb
		}
		time.Sleep(5 * time.Millisecond)
	}
	jb, _ := mgr.Get(context.Background(), kind)
	t.Fatalf("job did not reach status %q in time; last=%q", want, jb.Status)
	return jobModel.Job{}
}

func TestSubmit_Success(t *testing.T) {
	mgr := job.NewManager(newStubRepo())
	mgr.Register(kind, func(_ context.Context, _ struct{}, report job.ReportFunc) (any, error) {
		report("indexing", map[string]int{"n": 1})
		return map[string]string{"ok": "yes"}, nil
	})

	jb, err := mgr.Submit(context.Background(), kind, struct{}{})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if jb.Status != jobModel.StatusPending {
		t.Fatalf("expected pending on submit, got %q", jb.Status)
	}

	done := waitForStatus(t, mgr, jobModel.StatusSuccess)
	if done.FinishedAt == nil || done.StartedAt == nil {
		t.Fatalf("expected started/finished timestamps, got %+v", done)
	}
	if done.Payload != `{"ok":"yes"}` {
		t.Fatalf("expected result persisted to payload, got %q", done.Payload)
	}
}

func TestSubmit_Failure(t *testing.T) {
	mgr := job.NewManager(newStubRepo())
	mgr.Register(kind, func(_ context.Context, _ struct{}, _ job.ReportFunc) (any, error) {
		return nil, errors.New("boom")
	})

	if _, err := mgr.Submit(context.Background(), kind, struct{}{}); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	jb := waitForStatus(t, mgr, jobModel.StatusFailed)
	if jb.Error != "boom" {
		t.Fatalf("expected error 'boom' persisted, got %q", jb.Error)
	}
}

func TestSubmit_UnknownType(t *testing.T) {
	mgr := job.NewManager(newStubRepo())
	if _, err := mgr.Submit(context.Background(), jobModel.Kind[struct{}]("nope"), struct{}{}); !errors.Is(err, job.ErrNoRunner) {
		t.Fatalf("expected ErrNoRunner, got %v", err)
	}
}

func TestSubmit_MutexRejectsConcurrent(t *testing.T) {
	mgr := job.NewManager(newStubRepo())
	release := make(chan struct{})
	started := make(chan struct{})
	mgr.Register(kind, func(ctx context.Context, _ struct{}, _ job.ReportFunc) (any, error) {
		close(started)
		<-release
		return nil, nil
	})

	if _, err := mgr.Submit(context.Background(), kind, struct{}{}); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	<-started
	waitForStatus(t, mgr, jobModel.StatusRunning)

	if _, err := mgr.Submit(context.Background(), kind, struct{}{}); !errors.Is(err, job.ErrAlreadyRunning) {
		t.Fatalf("expected ErrAlreadyRunning on concurrent submit, got %v", err)
	}
	close(release)
	waitForStatus(t, mgr, jobModel.StatusSuccess)
}

func TestCancel_RunningJob(t *testing.T) {
	mgr := job.NewManager(newStubRepo())
	started := make(chan struct{})
	mgr.Register(kind, func(ctx context.Context, _ struct{}, _ job.ReportFunc) (any, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	if _, err := mgr.Submit(context.Background(), kind, struct{}{}); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	<-started
	if err := mgr.Cancel(kind); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	waitForStatus(t, mgr, jobModel.StatusCancelled)
}

func TestStart_SweepsOrphans(t *testing.T) {
	repo := newStubRepo()
	_ = repo.Upsert(context.Background(), &jobModel.Job{Type: string(kind), Status: jobModel.StatusRunning})
	mgr := job.NewManager(repo)

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	jb, err := mgr.Get(context.Background(), kind)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if jb.Status != jobModel.StatusFailed {
		t.Fatalf("expected orphan swept to failed, got %q", jb.Status)
	}
}

func TestSubmit_AfterTerminalReplaces(t *testing.T) {
	mgr := job.NewManager(newStubRepo())
	mgr.Register(kind, func(_ context.Context, _ struct{}, _ job.ReportFunc) (any, error) {
		return nil, nil
	})
	if _, err := mgr.Submit(context.Background(), kind, struct{}{}); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	waitForStatus(t, mgr, jobModel.StatusSuccess)
	if _, err := mgr.Submit(context.Background(), kind, struct{}{}); err != nil {
		t.Fatalf("resubmit after terminal failed: %v", err)
	}
	waitForStatus(t, mgr, jobModel.StatusSuccess)
}

func TestSubmit_TypedPayloadRoundTrip(t *testing.T) {
	type payload struct{ Name string }
	const typed jobModel.Kind[payload] = "typed"

	mgr := job.NewManager(newStubRepo())
	got := make(chan payload, 1)
	mgr.Register(typed, func(_ context.Context, p payload, _ job.ReportFunc) (any, error) {
		got <- p
		return nil, nil
	})

	jb, err := mgr.Submit(context.Background(), typed, payload{Name: "x"})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if jb.Payload != `{"Name":"x"}` {
		t.Fatalf("expected encoded payload on pending row, got %q", jb.Payload)
	}
	if p := <-got; p.Name != "x" {
		t.Fatalf("runner received %+v", p)
	}
}

func TestSubmit_EmptyPayloadNotPersisted(t *testing.T) {
	mgr := job.NewManager(newStubRepo())
	mgr.Register(kind, func(_ context.Context, _ struct{}, _ job.ReportFunc) (any, error) { return nil, nil })

	jb, err := mgr.Submit(context.Background(), kind, struct{}{})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if jb.Payload != "" {
		t.Fatalf("expected empty payload for struct{} kind, got %q", jb.Payload)
	}
	waitForStatus(t, mgr, jobModel.StatusSuccess)
}
