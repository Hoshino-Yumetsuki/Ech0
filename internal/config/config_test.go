// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package config

import "testing"

func TestGetJWTSecret_PrefersJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "primary")
	t.Setenv("ECH0_JWT_SECRET", "alias")

	if got := string(getJWTSecret()); got != "primary" {
		t.Fatalf("getJWTSecret() = %q, want %q", got, "primary")
	}
}

func TestGetJWTSecret_FallsBackToPrefixedAlias(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("ECH0_JWT_SECRET", "alias")

	if got := string(getJWTSecret()); got != "alias" {
		t.Fatalf("getJWTSecret() = %q, want %q", got, "alias")
	}
}

func TestGetJWTSecret_GeneratesEphemeralSecretWhenUnset(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("ECH0_JWT_SECRET", "")

	first := getJWTSecret()
	second := getJWTSecret()
	if len(first) == 0 {
		t.Fatal("expected a generated secret, got empty")
	}
	if string(first) == string(second) {
		t.Fatal("expected a fresh secret per call when unset")
	}
}
