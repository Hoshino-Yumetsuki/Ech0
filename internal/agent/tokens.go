// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"encoding/json"
	"strings"
	"unicode"
)

// EstimateTokens approximates how many tokens s costs, without a tokenizer.
//
// No single tokenizer is right here — the provider is whatever the operator
// configured — so this aims for the one property budgeting needs: it does not
// undercount. CJK text runs close to a token per character across the BPE
// vocabularies in use; everything else (Latin text, digits, JSON, URLs) runs
// three to four characters per token, and three is the conservative end.
// Counting every rune as a token, as a rune count does, overstates English by
// three- to four-fold and starves every budget built on it.
func EstimateTokens(s string) int {
	if s == "" {
		return 0
	}
	wide, other := 0, 0
	for _, r := range s {
		if isWideRune(r) {
			wide++
		} else {
			other++
		}
	}
	return wide + (other+2)/3
}

func isWideRune(r rune) bool {
	if r < 0x2E80 {
		return false
	}
	return unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) ||
		(r >= 0x3000 && r <= 0x303F) || (r >= 0xFF00 && r <= 0xFFEF)
}

// TruncateTokens cuts s to roughly budget tokens and marks the cut with note.
// It never splits a rune, and returns s unchanged when it already fits.
func TruncateTokens(s string, budget int, note string) string {
	if EstimateTokens(s) <= budget {
		return s
	}
	budget -= EstimateTokens(note)
	if budget <= 0 {
		return note
	}
	cut, used, other := len(s), 0, 0
	for i, r := range s {
		if isWideRune(r) {
			used++
		} else if other++; other%3 == 1 {
			used++
		}
		if used > budget {
			cut = i
			break
		}
	}
	return strings.TrimRightFunc(s[:cut], unicode.IsSpace) + "\n" + note
}

// ToolDefTokens is what the tools' declarations cost, estimated. They are sent
// with every request of a run, whether or not any tool is called.
func ToolDefTokens(tools []Tool) int {
	n := 0
	for _, t := range tools {
		raw, _ := json.Marshal(t.Def)
		n += EstimateTokens(string(raw))
	}
	return n
}
