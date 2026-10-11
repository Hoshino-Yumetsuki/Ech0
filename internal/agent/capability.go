// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import "strings"

// isOpenAIReasoningModel reports the o-series and GPT-5 family, whose reasoning
// items the Responses provider asks for and replays.
//
// This is the one place the model's name decides anything, and it is safe to
// be incomplete: a model it misses simply runs without reasoning replay, a
// small loss in quality rather than a failed request. Anything whose miss
// would fail a request — sampling parameters, output ceilings — is left to
// the operator's configuration instead of a list kept here.
func isOpenAIReasoningModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	if i := strings.LastIndex(m, "/"); i >= 0 {
		m = m[i+1:]
	}
	for _, p := range []string{"o1", "o3", "o4", "gpt-5"} {
		if strings.HasPrefix(m, p) {
			return true
		}
	}
	return false
}
