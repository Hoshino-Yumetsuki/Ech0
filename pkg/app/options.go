// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package app

import (
	"context"
	"os"
	"syscall"
	"time"
)

type Hook func(context.Context) error

type Option func(*options)

type options struct {
	ctx        context.Context
	components []Component

	beforeStart []Hook
	afterStart  []Hook
	beforeStop  []Hook
	afterStop   []Hook

	sigs        []os.Signal
	stopTimeout time.Duration
}

func defaultOptions() options {
	return options{
		ctx:         context.Background(),
		sigs:        []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		stopTimeout: 5 * time.Second,
	}
}

func Components(components ...Component) Option {
	return func(o *options) {
		o.components = append(o.components, components...)
	}
}

func BeforeStart(h Hook) Option {
	return func(o *options) {
		if h != nil {
			o.beforeStart = append(o.beforeStart, h)
		}
	}
}

func AfterStart(h Hook) Option {
	return func(o *options) {
		if h != nil {
			o.afterStart = append(o.afterStart, h)
		}
	}
}

func BeforeStop(h Hook) Option {
	return func(o *options) {
		if h != nil {
			o.beforeStop = append(o.beforeStop, h)
		}
	}
}

func AfterStop(h Hook) Option {
	return func(o *options) {
		if h != nil {
			o.afterStop = append(o.afterStop, h)
		}
	}
}

func Signals(sigs ...os.Signal) Option {
	return func(o *options) {
		if len(sigs) == 0 {
			return
		}
		o.sigs = append([]os.Signal(nil), sigs...)
	}
}

func StopTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout < 0 {
			return
		}
		o.stopTimeout = timeout
	}
}

func Context(ctx context.Context) Option {
	return func(o *options) {
		if ctx != nil {
			o.ctx = ctx
		}
	}
}
