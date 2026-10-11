// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const maxStoredChatMessages = 50

// historyTruncateNote marks an old answer cut to fit the history budget.
const historyTruncateNote = "…"

// historyForModel turns the stored conversation into the history a new request
// carries, newest turns first until budgetTokens is spent.
//
// It works in whole turns — a question with the answers that followed it —
// because either half alone misleads: an answer without its question reads as
// the model's own unprompted words, and a request whose history opens on an
// assistant message is malformed for several providers. When even the newest
// turn is over budget it is kept and its answer cut down, since the most recent
// exchange is the one a follow-up ("tell me more about the second one") refers
// to.
func historyForModel(msgs []ChatMessage, locale string, budgetTokens int, loc *time.Location) []agent.Message {
	if len(msgs) == 0 {
		return nil
	}

	lastSourced := -1
	for i, msg := range slices.Backward(msgs) {
		if msg.Role == "assistant" && len(msg.Sources) > 0 {
			lastSourced = i
			break
		}
	}

	contentOf := func(i int) string {
		c := strings.TrimSpace(msgs[i].Content)
		if i != lastSourced {
			return c
		}
		note := fmt.Sprintf(recentSourcesNoteFor(locale), formatSearchResults(msgs[i].Sources, nil, loc))
		if c == "" {
			return note
		}
		return c + "\n\n" + note
	}

	var turns [][]agent.Message
	for i := range msgs {
		content := contentOf(i)
		if content == "" {
			continue
		}
		role := roleFromString(msgs[i].Role)
		if role == agent.RoleUser {
			turns = append(turns, nil)
		} else if len(turns) == 0 {
			continue
		}
		last := len(turns) - 1
		turns[last] = append(turns[last], agent.Message{Role: role, Content: content})
	}

	used := 0
	first := len(turns)
	for first > 0 {
		t := turnTokens(turns[first-1])
		if used+t > budgetTokens {
			if first == len(turns) {
				turns[first-1] = fitTurn(turns[first-1], budgetTokens)
				first--
			}
			break
		}
		used += t
		first--
	}

	var out []agent.Message
	for _, turn := range turns[first:] {
		out = append(out, turn...)
	}
	return out
}

func turnTokens(turn []agent.Message) int {
	n := 0
	for _, m := range turn {
		n += agent.EstimateTokens(m.Content)
	}
	return n
}

// fitTurn cuts a turn's answers so the whole turn fits budget, keeping the
// question intact: it is short, and it is what the answer means.
func fitTurn(turn []agent.Message, budget int) []agent.Message {
	out := slices.Clone(turn)
	left := budget - agent.EstimateTokens(out[0].Content)
	for i := 1; i < len(out); i++ {
		share := max(left/(len(out)-i), 0)
		out[i].Content = agent.TruncateTokens(out[i].Content, share, historyTruncateNote)
		left -= agent.EstimateTokens(out[i].Content)
	}
	return out
}

func roleFromString(r string) agent.Role {
	if r == "assistant" {
		return agent.RoleAssistant
	}
	return agent.RoleUser
}

type ChatMessage struct {
	Role        string                        `json:"role"`
	Content     string                        `json:"content"`
	Sources     []embeddingModel.SearchResult `json:"sources,omitempty"`
	Reasoning   string                        `json:"reasoning,omitempty"`
	ReasoningMs int64                         `json:"reasoning_ms,omitempty"`
	Asks        []AskExchange                 `json:"asks,omitempty"`
}

func chatSessionKey(userID string) string {
	return commonModel.ChatSessionKeyPrefix + userID
}

func (s *CopilotService) loadSession(ctx context.Context, userID string) []ChatMessage {
	if userID == "" {
		return nil
	}
	raw, err := s.durableKV.Get(ctx, chatSessionKey(userID))
	if err != nil {
		return nil
	}
	var msgs []ChatMessage
	if err := json.Unmarshal([]byte(raw), &msgs); err != nil {
		return nil
	}
	return msgs
}

func (s *CopilotService) appendTurn(ctx context.Context, userID string, turn ...ChatMessage) {
	if userID == "" {
		return
	}
	msgs := append(s.loadSession(ctx, userID), turn...)
	if len(msgs) > maxStoredChatMessages {
		msgs = msgs[len(msgs)-maxStoredChatMessages:]
	}
	payload, err := json.Marshal(msgs)
	if err != nil {
		logUtil.GetLogger().Warn("failed to marshal chat session",
			slog.String("module", "copilot"), logUtil.Err(err))
		return
	}
	if err := s.durableKV.Set(ctx, chatSessionKey(userID), string(payload)); err != nil {
		logUtil.GetLogger().Warn("failed to persist chat session",
			slog.String("module", "copilot"), logUtil.Err(err))
	}
}

type assistantTurn struct {
	answer      string
	sources     []embeddingModel.SearchResult
	reasoning   string
	reasoningMs int64
	asks        []AskExchange
}

// persistTurn writes the turn down, or drops it when there is nothing in it.
//
// An answered question counts as something in it. A turn that ended right after
// a confirmation — approved and applied, or declined — may have produced no
// prose and no sources, and dropping it would erase the one exchange the person
// actually took part in.
func (s *CopilotService) persistTurn(ctx context.Context, userID, question string, turn assistantTurn) {
	if strings.TrimSpace(turn.answer) == "" && len(turn.sources) == 0 && len(turn.asks) == 0 {
		return
	}
	s.appendTurn(ctx, userID,
		ChatMessage{Role: "user", Content: question},
		ChatMessage{
			Role:        "assistant",
			Content:     turn.answer,
			Sources:     turn.sources,
			Reasoning:   turn.reasoning,
			ReasoningMs: turn.reasoningMs,
			Asks:        turn.asks,
		},
	)
}

func (s *CopilotService) GetSession(ctx context.Context) ([]ChatMessage, error) {
	user, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	msgs := s.loadSession(ctx, user.ID)
	if msgs == nil {
		return []ChatMessage{}, nil
	}
	return msgs, nil
}

func (s *CopilotService) ClearSession(ctx context.Context) error {
	user, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	return s.durableKV.Delete(ctx, chatSessionKey(user.ID))
}
