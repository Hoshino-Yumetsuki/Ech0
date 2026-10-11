// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
)

func src(content string) embeddingModel.SearchResult {
	return embeddingModel.SearchResult{Content: content, EchoCreated: 0}
}

func TestHistoryForModel_Empty(t *testing.T) {
	if got := historyForModel(nil, "zh-CN", maxHistoryTokens, time.UTC); len(got) != 0 {
		t.Fatalf("expected empty history, got %d messages", len(got))
	}
}

func TestHistoryForModel_DropsOldSourcesFoldsRecent(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: "q1"},
		{Role: "assistant", Content: "a1", Sources: []embeddingModel.SearchResult{src("OLD_ECHO")}},
		{Role: "user", Content: "q2"},
		{Role: "assistant", Content: "a2", Sources: []embeddingModel.SearchResult{src("RECENT_ECHO")}},
	}

	got := historyForModel(msgs, "zh-CN", maxHistoryTokens, time.UTC)
	if len(got) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(got))
	}

	wantRoles := []agent.Role{agent.RoleUser, agent.RoleAssistant, agent.RoleUser, agent.RoleAssistant}
	for i, r := range wantRoles {
		if got[i].Role != r {
			t.Fatalf("msg %d: want role %q, got %q", i, r, got[i].Role)
		}
	}

	a1 := got[1].Content
	if a1 != "a1" {
		t.Fatalf("old assistant content should stay plain, got %q", a1)
	}
	for _, m := range got {
		if strings.Contains(m.Content, "OLD_ECHO") {
			t.Fatalf("old sources should be dropped, but found in %q", m.Content)
		}
	}

	a2 := got[3].Content
	if !strings.Contains(a2, "a2") || !strings.Contains(a2, "RECENT_ECHO") {
		t.Fatalf("recent assistant should fold in its sources, got %q", a2)
	}
}

func TestHistoryForModel_SkipsEmpty(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: "q1"},
		{Role: "assistant", Content: ""},
		{Role: "user", Content: "q2"},
		{Role: "assistant", Content: "a2"},
	}

	got := historyForModel(msgs, "en-US", maxHistoryTokens, time.UTC)
	if len(got) != 3 {
		t.Fatalf("expected 3 messages (empty skipped), got %d", len(got))
	}
	for _, m := range got {
		if strings.TrimSpace(m.Content) == "" {
			t.Fatalf("empty message should have been skipped")
		}
	}
}

func TestHistoryForModel_TokenBudgetKeepsRecentInOrder(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: strings.Repeat("a", 10)},
		{Role: "assistant", Content: strings.Repeat("b", 10)},
		{Role: "user", Content: strings.Repeat("c", 10)},
		{Role: "assistant", Content: strings.Repeat("d", 10)},
	}

	// Each 10-letter message is ~4 tokens, so one turn is ~8: room for one.
	got := historyForModel(msgs, "zh-CN", 10, time.UTC)
	if len(got) != 2 {
		t.Fatalf("expected 2 messages within budget, got %d", len(got))
	}
	if got[0].Content != strings.Repeat("c", 10) || got[1].Content != strings.Repeat("d", 10) {
		t.Fatalf("expected most-recent two in time order, got [%q, %q]", got[0].Content, got[1].Content)
	}
}

func TestHistoryForModel_TinyBudgetKeepsAtLeastOne(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: "q1"},
		{Role: "assistant", Content: strings.Repeat("z", 100)},
	}

	got := historyForModel(msgs, "zh-CN", 10, time.UTC)
	if len(got) != 2 {
		t.Fatalf("expected the newest turn kept whole-shaped under a tiny budget, got %d messages", len(got))
	}
	if got[0].Role != agent.RoleUser || got[0].Content != "q1" {
		t.Fatalf("the question must survive; an answer alone reads as unprompted, got %+v", got[0])
	}
	if got[1].Role != agent.RoleAssistant || len(got[1].Content) >= 100 || !strings.HasSuffix(got[1].Content, historyTruncateNote) {
		t.Fatalf("the answer should be cut to fit and marked, got %q", got[1].Content)
	}
}

func TestHistoryForModel_NeverOpensOnAssistant(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "assistant", Content: "orphan answer"},
		{Role: "user", Content: strings.Repeat("长", 50)},
		{Role: "assistant", Content: strings.Repeat("答", 50)},
		{Role: "user", Content: "q2"},
		{Role: "assistant", Content: "a2"},
	}
	for _, budget := range []int{5, 60, 150, 10_000} {
		got := historyForModel(msgs, "zh-CN", budget, time.UTC)
		if len(got) == 0 || got[0].Role != agent.RoleUser {
			t.Fatalf("budget %d: history must open on a question, got %+v", budget, got)
		}
	}
}
