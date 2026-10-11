// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package setting

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/kvstore"
)

type Spec[T any] struct {
	Key       string
	Default   func() T
	Normalize func(*T)
	Migrate   func(context.Context, kvstore.Store) (T, bool)
}

// Get reads the stored value, falling back to Pristine when the key is missing
// (no error) or unreadable (with the error).
func (s Spec[T]) Get(ctx context.Context, kv kvstore.Store) (T, error) {
	raw, err := kv.Get(ctx, s.Key)
	if errors.Is(err, kvstore.ErrNotFound) {
		return s.Pristine(), nil
	}
	if err != nil {
		return s.Pristine(), err
	}

	var v T
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return s.Pristine(), err
	}
	s.normalize(&v)
	return v, nil
}

func (s Spec[T]) Set(ctx context.Context, kv kvstore.Store, value T) error {
	s.normalize(&value)
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return kv.Set(ctx, s.Key, string(buf))
}

// Pristine is the normalized default: what Get yields before anything is stored.
func (s Spec[T]) Pristine() T {
	v := s.Default()
	s.normalize(&v)
	return v
}

func (s Spec[T]) normalize(v *T) {
	if s.Normalize != nil {
		s.Normalize(v)
	}
}

type seedable interface {
	seed(ctx context.Context, kv kvstore.Store) error
}

// seed writes the initial value once: migrated from legacy storage when
// possible, otherwise the default. An existing value is left untouched.
func (s Spec[T]) seed(ctx context.Context, kv kvstore.Store) error {
	if _, err := kv.Get(ctx, s.Key); !errors.Is(err, kvstore.ErrNotFound) {
		return err
	}

	var (
		v        T
		migrated bool
	)
	if s.Migrate != nil {
		v, migrated = s.Migrate(ctx, kv)
	}
	if !migrated {
		v = s.Default()
	}
	return s.Set(ctx, kv, v)
}

func Seed(ctx context.Context, kv kvstore.Store) error {
	for _, s := range registry {
		if err := s.seed(ctx, kv); err != nil {
			return err
		}
	}
	return nil
}
