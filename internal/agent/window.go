// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import "strings"

// window is how much one run may send, kept in the provider's own tokens.
//
// Before the first reply the estimator is all there is, and it is blind to
// what a provider adds around the messages — role framing, the tool-use
// preamble, the reasoning a turn replays. Every reply then reports what its
// request really cost, so from the second round on the loop anchors on that
// count and estimates only what changed since: the error is confined to one
// round's growth instead of spread over the whole conversation.
type window struct {
	// limit is RunRequest.MaxContextTokens; zero means the caller set none.
	limit int
	// fixed is what the tool declarations cost, estimated. It rides on every
	// request but is not among the messages trimming measures.
	fixed int
	// offset is how far the last reported count sat above the estimate of the
	// same request; negative when the estimate ran high.
	offset int
	// ceiling tightens limit after a provider refused a request as too long.
	ceiling int
}

// messageBudget is what the messages may take, in estimated tokens, for
// trimContext. Zero means no limit is known.
func (w *window) messageBudget() int {
	limit := w.limit
	if w.ceiling > 0 && (limit <= 0 || w.ceiling < limit) {
		limit = w.ceiling
	}
	if limit <= 0 {
		return 0
	}
	return max(limit-w.fixed-w.offset, 1)
}

// observe anchors on a reply's reported input for the request whose estimate
// was sent. A provider that reports nothing leaves the previous anchor.
func (w *window) observe(sent int, u Usage) {
	if u.InputTokens > 0 {
		w.offset = u.InputTokens - sent
	}
}

// overflowed records that a request estimated at sent was refused as too long.
// Whatever the numbers said, it did not fit, so the next attempt aims a quarter
// below it.
func (w *window) overflowed(sent int) {
	w.ceiling = max((sent+w.offset)*3/4, 1)
}

// contextOverflowMarkers are how providers word a request that is longer than
// the model's window. There is no shared error code — OpenAI and its
// Responses API say context_length_exceeded, Anthropic "prompt is too long",
// and compatible servers (vLLM, DeepSeek and the like) mostly echo OpenAI's
// "maximum context length" sentence — so the message is all there is to go on.
var contextOverflowMarkers = []string{
	"context_length_exceeded",
	"maximum context length",
	"prompt is too long",
	"exceeds the context window",
	"input is too long",
}

// isContextOverflow reports whether err is a provider refusing a request for
// being longer than the model's context window.
func isContextOverflow(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, m := range contextOverflowMarkers {
		if strings.Contains(msg, m) {
			return true
		}
	}
	return false
}
