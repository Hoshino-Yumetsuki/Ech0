// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func doneWith(ev Event) Event {
	ev.Kind = EventDone
	return ev
}

// A round cut off by the output limit must not run what it asked for: the
// arguments may stop mid-string, and a write built from them is a write nobody
// asked for.
func TestRunLoop_TruncatedRoundDropsCallsAndMarksAnswer(t *testing.T) {
	tool, execs := countingTool("create_echo", ToolOutput{Content: "ok"}, nil)
	fp := &fakeProvider{scripts: [][]Event{{
		textEvent("写到一半"),
		toolCallEvent("c1", "create_echo", `{"content":"未写完`),
		doneWith(Event{Truncated: true}),
	}}}

	evs := runLoopSync(context.Background(), fp, RunRequest{
		Setting: enabledSetting(),
		Tools:   []Tool{tool},
	})

	if *execs != 0 {
		t.Fatalf("a truncated round's tool call ran %d times", *execs)
	}
	if fp.calls != 1 {
		t.Fatalf("the run should end on truncation, made %d requests", fp.calls)
	}
	if countKind(evs, AgentError) != 0 || evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("truncation should close the run normally: %v", kinds(evs))
	}
	var text strings.Builder
	for _, ev := range evs {
		if ev.Kind == AgentDelta {
			text.WriteString(ev.Text)
		}
	}
	if !strings.HasPrefix(text.String(), "写到一半") || !strings.HasSuffix(text.String(), defaultRunStrings.OutputTruncated) {
		t.Fatalf("the answer should keep what was said and say it is unfinished, got %q", text.String())
	}
}

func TestRunLoop_ContextOverflowRetriesTrimmed(t *testing.T) {
	msgs := []Message{
		{Role: RoleUser, Content: "Q"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "t1", Name: "search_echos"}}},
		{Role: RoleTool, ToolCallID: "t1", Content: strings.Repeat("a", 3000)},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "t2", Name: "search_echos"}}},
		{Role: RoleTool, ToolCallID: "t2", Content: "fresh"},
	}
	fp := &fakeProvider{scripts: [][]Event{
		{errEvent(errors.New(`400 Bad Request: {"code":"context_length_exceeded"}`))},
		{textEvent("ok"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{
		Setting:          enabledSetting(),
		Messages:         msgs,
		MaxContextTokens: 100_000,
	})

	if fp.calls != 2 {
		t.Fatalf("an overflow should be retried once, made %d requests", fp.calls)
	}
	if got := fp.gotReqs[1].Messages[2].Content; got != defaultRunStrings.ContextTrimNote {
		t.Fatalf("the retry should clear the old result, got %d bytes", len(got))
	}
	if countKind(evs, AgentError) != 0 || evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("the retry should answer: %v", kinds(evs))
	}
}

func TestRunLoop_ContextOverflowNotRetried(t *testing.T) {
	overflow := errors.New("prompt is too long: 210000 tokens > 200000 maximum")
	cases := []struct {
		name   string
		msgs   []Message
		script []Event
	}{
		{
			name:   "nothing left to trim",
			msgs:   []Message{{Role: RoleSystem, Content: "S"}, {Role: RoleUser, Content: "Q"}},
			script: []Event{errEvent(overflow)},
		},
		{
			name: "text already reached the person",
			msgs: []Message{
				{Role: RoleUser, Content: "Q"},
				{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "t1", Name: "search_echos"}}},
				{Role: RoleTool, ToolCallID: "t1", Content: strings.Repeat("a", 3000)},
				{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "t2", Name: "search_echos"}}},
				{Role: RoleTool, ToolCallID: "t2", Content: "fresh"},
			},
			script: []Event{textEvent("partial"), errEvent(overflow)},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fp := &fakeProvider{scripts: [][]Event{c.script, {textEvent("ok"), doneEvent()}}}
			evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Messages: c.msgs})
			if fp.calls != 1 || countKind(evs, AgentError) != 1 {
				t.Fatalf("want one request and an error, got %d requests: %v", fp.calls, kinds(evs))
			}
		})
	}
}

// Once a reply says what a request really cost, the next round is sized from
// that count, not from the estimate alone.
func TestRunLoop_AnchorsOnReportedUsage(t *testing.T) {
	tool, _ := countingTool("search_echos", ToolOutput{Content: strings.Repeat("x", 1350)}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{}`), doneWith(Event{Usage: Usage{InputTokens: 800}})},
		{textEvent("ok"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{
		Setting:          enabledSetting(),
		Messages:         []Message{{Role: RoleUser, Content: "Q"}},
		Tools:            []Tool{tool},
		MaxContextTokens: 1_000,
	})

	got := fp.gotReqs[1].Messages[2].Content
	if !strings.HasSuffix(got, defaultRunStrings.TruncateNote) {
		t.Fatalf("450 estimated tokens fit 1000, but 800 were already spent before them; the result should be cut, got %d bytes", len(got))
	}
}

func TestWindow(t *testing.T) {
	w := &window{limit: 1_000, fixed: 100}
	if got := w.messageBudget(); got != 900 {
		t.Fatalf("budget before any usage = %d, want 900", got)
	}

	w.observe(500, Usage{InputTokens: 700})
	if got := w.messageBudget(); got != 700 {
		t.Fatalf("budget after a 200-token underestimate = %d, want 700", got)
	}
	w.observe(500, Usage{})
	if got := w.messageBudget(); got != 700 {
		t.Fatalf("a reply without usage should keep the last anchor, got %d", got)
	}

	w.overflowed(800)
	if got := w.messageBudget(); got != 450 {
		t.Fatalf("budget after an overflow at 800 = %d, want 450 (a quarter under 1000 real, less tools and offset)", got)
	}

	if got := (&window{}).messageBudget(); got != 0 {
		t.Fatalf("no limit should mean no trimming, got %d", got)
	}
	unset := &window{}
	unset.overflowed(1_000)
	if got := unset.messageBudget(); got != 750 {
		t.Fatalf("an overflow should bound even a run without a limit, got %d", got)
	}
}

func TestIsContextOverflow(t *testing.T) {
	for _, msg := range []string{
		`error, status code: 400, message: This model's maximum context length is 131072 tokens.`,
		`POST "https://api.openai.com/v1/responses": 400 Bad Request {"code":"context_length_exceeded"}`,
		`prompt is too long: 208310 tokens > 200000 maximum`,
		`Your input exceeds the context window of this model.`,
	} {
		if !isContextOverflow(errors.New(msg)) {
			t.Errorf("not recognised as an overflow: %q", msg)
		}
	}
	for _, err := range []error{nil, errors.New("401 Unauthorized"), errors.New("rate limit exceeded")} {
		if isContextOverflow(err) {
			t.Errorf("wrongly taken for an overflow: %v", err)
		}
	}
}

// Clearing costs the prompt cache from the cleared point on, so once it has to
// happen it clears past the limit and leaves room to append.
func TestTrimContext_ClearsBelowLimitToSpareTheCache(t *testing.T) {
	strs := defaultRunStrings
	old := strings.Repeat("o", 300) // 100 tokens each; clearing one saves ~84
	msgs := []Message{
		{Role: RoleUser, Content: "Q"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "a"}}},
		{Role: RoleTool, ToolCallID: "a", Content: old},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "b"}}},
		{Role: RoleTool, ToolCallID: "b", Content: old},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "c"}}},
		{Role: RoleTool, ToolCallID: "c", Content: old},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "d"}}},
		{Role: RoleTool, ToolCallID: "d", Content: "fresh"},
	}

	trimContext(msgs, 250, strs)

	if msgs[2].Content != strs.ContextTrimNote || msgs[4].Content != strs.ContextTrimNote {
		t.Fatalf("clearing should go below the limit, not just under it")
	}
	if msgs[6].Content != old {
		t.Fatalf("clearing should stop once below the target")
	}
	if got := contextTokens(msgs); got > 250*trimTargetPercent/100 {
		t.Fatalf("context = %d, want <= %d", got, 250*trimTargetPercent/100)
	}
}
