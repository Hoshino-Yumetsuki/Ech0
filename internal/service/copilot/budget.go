// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

// contextPlan divides one model's context window between the parts of a chat
// request, all in estimated tokens (agent.EstimateTokens).
//
// The parts are sized together because they share one window: a summary sized
// against the whole window leaves no room for the prompt and history it is sent
// alongside, and the loop's trimming then drops the summary itself — the model
// writes a year-end review from nothing. Each share here is carved out of what
// the others leave.
type contextPlan struct {
	// Input is the ceiling for the whole request, handed to the run loop.
	Input int
	// History is what earlier turns of the conversation may take.
	History int
	// Material is what one summarize_echos result may take.
	Material int
}

const (
	// defaultContextWindow is assumed when the operator has not set one. It is
	// deliberately modest: overestimating the window fails the request with
	// context_length_exceeded, while underestimating only means summaries go
	// through one more map-reduce pass.
	defaultContextWindow = 64_000
	minContextWindow     = 4_000

	// The model's own output is budgeted out of the window first.
	minOutputReserve = 1_024
	maxOutputReserve = 16_000

	// estimateMargin absorbs the estimator's error against real tokenizers.
	estimateMargin = 0.9

	minHistoryTokens = 300
	maxHistoryTokens = 16_000

	// minAggregateBudget keeps map-reduce chunks large enough to be worth a call.
	minAggregateBudget = 1_000
)

func planContext(window, fixedTokens int) contextPlan {
	if window <= 0 {
		window = defaultContextWindow
	}
	window = max(window, minContextWindow)

	reserve := min(max(window/4, minOutputReserve), maxOutputReserve)
	input := int(float64(window-reserve) * estimateMargin)
	free := max(input-fixedTokens, 0)

	history := min(max(free/4, minHistoryTokens), maxHistoryTokens, free/2)
	material := max((free-history)*6/10, minAggregateBudget)

	return contextPlan{Input: input, History: history, Material: material}
}
