// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	model "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	openai "github.com/sashabaranov/go-openai"
)

func TestRunLoop_FailedCallIsRetriedNotDeduped(t *testing.T) {
	calls := 0
	tool := Tool{
		Def:    ToolDef{Name: "search_echos", Parameters: json.RawMessage(`{"type":"object"}`)},
		Effect: EffectRead,
		Run: func(context.Context, json.RawMessage) (ToolOutput, error) {
			calls++
			if calls == 1 {
				return ToolOutput{}, errors.New("429 rate limited")
			}
			return ToolOutput{Content: "hit"}, nil
		},
	}
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"same"}`), doneEvent()},
		{toolCallEvent("c2", "search_echos", `{"q":"same"}`), doneEvent()},
		{textEvent("done"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if calls != 2 {
		t.Fatalf("tool executed %d times, want 2 (a failed call must not count as searched)", calls)
	}
	msgs := fp.gotReqs[1].Messages
	if last := msgs[len(msgs)-1]; last.Role != RoleTool || !last.IsError {
		t.Fatalf("failed result should be flagged IsError, got %+v", last)
	}
}

func TestRunLoop_TrimmedResultIsSearchedAgain(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{Content: strings.Repeat("x", 600)}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"a"}`), doneEvent()},
		{toolCallEvent("c2", "search_echos", `{"q":"b"}`), doneEvent()},
		{toolCallEvent("c3", "search_echos", `{"q":"a"}`), doneEvent()},
		{textEvent("done"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{
		Setting:          enabledSetting(),
		Tools:            []Tool{tool},
		MaxRounds:        4,
		MaxContextTokens: 300,
	})

	if *execs != 3 {
		t.Fatalf("tool executed %d times, want 3: the first result was trimmed, so pointing at it would be a lie", *execs)
	}
}

func TestRunLoop_SameCallTwiceInOneRoundRunsOnce(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"a"}`), toolCallEvent("c2", "search_echos", `{"q":"a"}`), doneEvent()},
		{textEvent("done"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if *execs != 1 {
		t.Fatalf("tool executed %d times, want 1", *execs)
	}
	msgs := fp.gotReqs[1].Messages
	if msgs[len(msgs)-1].ToolCallID != "c2" || msgs[len(msgs)-1].Content != defaultRunStrings.DedupNote {
		t.Fatalf("the duplicate still needs its own result, got %+v", msgs[len(msgs)-1])
	}
}

func TestRunLoop_NativeTurnCarriedOnAssistantMessage(t *testing.T) {
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	native := openaiTurn{reasoning: "let me search"}
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"a"}`), {Kind: EventDone, Native: native}},
		{textEvent("done"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	var assistant *Message
	for i, m := range fp.gotReqs[1].Messages {
		if m.Role == RoleAssistant {
			assistant = &fp.gotReqs[1].Messages[i]
		}
	}
	if assistant == nil || assistant.Native != native {
		t.Fatalf("assistant turn should carry the provider's native record, got %+v", assistant)
	}
}

func TestTrimContext_DropsImagesThenTruncatesFreshResults(t *testing.T) {
	strs := defaultRunStrings
	msgs := []Message{
		{Role: RoleUser, Content: "Q"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "t1", Name: "search_echos"}}},
		{Role: RoleTool, ToolCallID: "t1", Content: strings.Repeat("r", 3000)},
		{Role: RoleUser, Content: strs.ImageNote, Images: []ImagePart{{MediaType: "image/png", Base64: "x"}}},
	}

	trimContext(msgs, 600, strs)

	if len(msgs[3].Images) != 0 {
		t.Fatalf("images should go before the newest results are cut")
	}
	if msgs[2].Content == strs.ContextTrimNote {
		t.Fatalf("the newest round's result must be cut down, never dropped")
	}
	if !strings.HasSuffix(msgs[2].Content, strs.TruncateNote) {
		t.Fatalf("a cut result should say so, got tail %q", msgs[2].Content[len(msgs[2].Content)-40:])
	}
	if got := contextTokens(msgs); got > 600 {
		t.Fatalf("context = %d tokens, want <= 600", got)
	}
}

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens(strings.Repeat("a", 300)); got != 100 {
		t.Fatalf("latin estimate = %d, want 100", got)
	}
	if got := EstimateTokens("今天天气很好"); got != 6 {
		t.Fatalf("CJK estimate = %d, want 6", got)
	}
	if got := TruncateTokens("今天天气很好，适合出门", 4, ""); !strings.HasPrefix(got, "今天天") {
		t.Fatalf("truncate kept %q", got)
	}
}

func TestToolCallAccumulator_MissingIndexSplitsByID(t *testing.T) {
	acc := newToolCallAccumulator()
	acc.add([]openai.ToolCall{{ID: "a", Function: openai.FunctionCall{Name: "search_echos", Arguments: `{"query":`}}})
	acc.add([]openai.ToolCall{{Function: openai.FunctionCall{Arguments: `"x"}`}}})
	acc.add([]openai.ToolCall{{ID: "b", Function: openai.FunctionCall{Name: "stats_overview", Arguments: `{}`}}})

	got := acc.finish()
	if len(got) != 2 {
		t.Fatalf("got %d calls, want 2: %+v", len(got), got)
	}
	if got[0].Name != "search_echos" || string(got[0].Args) != `{"query":"x"}` {
		t.Fatalf("first call = %+v", got[0])
	}
	if got[1].Name != "stats_overview" || got[1].ID != "b" {
		t.Fatalf("second call = %+v", got[1])
	}
}

// Sampling is the operator's call, per model: nothing is sent unless it was
// configured, and what was configured is sent as-is. No output ceiling is sent
// either, so no model has a field name to reject.
func TestOpenAIBuildRequest_SamplingOnlyWhenConfigured(t *testing.T) {
	p := &openaiProvider{setting: model.AgentSetting{Model: "o4-mini"}}
	req := p.buildRequest(Request{
		Messages:   []Message{{Role: RoleUser, Content: "hi"}},
		Tools:      []ToolDef{{Name: "t", Parameters: json.RawMessage(`{"type":"object"}`)}},
		ToolChoice: ToolChoiceNone,
	})
	body := wireBody(t, req)
	for _, k := range []string{"temperature", "max_tokens", "max_completion_tokens"} {
		if _, has := body[k]; has {
			t.Fatalf("%s must not be sent unless configured: %v", k, body)
		}
	}
	if req.ToolChoice != "none" {
		t.Fatalf("tool_choice = %v, want none", req.ToolChoice)
	}
	if err := openai.NewReasoningValidator().Validate(req); err != nil {
		t.Fatalf("go-openai would reject this request client-side: %v", err)
	}

	temp := 0.4
	p.setting = model.AgentSetting{Model: "gpt-4o", Temperature: &temp}
	body = wireBody(t, p.buildRequest(Request{Messages: []Message{{Role: RoleUser, Content: "hi"}}}))
	if got, ok := body["temperature"].(float64); !ok || float32(got) != float32(temp) {
		t.Fatalf("configured temperature should be sent, got %v", body["temperature"])
	}
}

func wireBody(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestOpenAIBuildMessages_ReplaysReasoningContent(t *testing.T) {
	p := &openaiProvider{}
	msgs := p.buildMessages([]Message{
		{Role: RoleUser, Content: "q"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "c1", Name: "s", Args: json.RawMessage(`{}`)}}, Native: openaiTurn{reasoning: "think"}},
		{Role: RoleAssistant, Content: "foreign", Native: anthropicTurn{}},
	})
	if msgs[1].ReasoningContent != "think" {
		t.Fatalf("reasoning_content must be passed back on the tool-call turn, got %q", msgs[1].ReasoningContent)
	}
	if msgs[2].ReasoningContent != "" {
		t.Fatalf("another provider's native record must be ignored")
	}
}

func TestAnthropicBuildParams_FinalRoundAndSampling(t *testing.T) {
	tools := []ToolDef{{Name: "t", Parameters: json.RawMessage(`{"type":"object","properties":{}}`)}}

	p := &anthropicProvider{setting: model.AgentSetting{Model: "claude-opus-5"}}
	params := p.buildParams(Request{
		Messages:   []Message{{Role: RoleUser, Content: "hi"}},
		Tools:      tools,
		ToolChoice: ToolChoiceNone,
	})
	if len(params.Tools) != 1 || params.ToolChoice.OfNone == nil {
		t.Fatalf("final round must declare tools with tool_choice none, got tools=%d choice=%+v", len(params.Tools), params.ToolChoice)
	}
	if params.Temperature.Valid() {
		t.Fatalf("temperature must not be sent unless configured")
	}
	if params.MaxTokens != anthropicMaxTokens {
		t.Fatalf("max_tokens = %d", params.MaxTokens)
	}

	temp := 0.3
	p.setting.Temperature = &temp
	params = p.buildParams(Request{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if !params.Temperature.Valid() || params.Temperature.Value != temp {
		t.Fatalf("configured temperature should be sent, got %+v", params.Temperature)
	}
}

func TestAnthropicBuildMessages_ReplaysNativeTurnAndErrors(t *testing.T) {
	native := anthropicTurn{content: []anthropic.ContentBlockParamUnion{
		anthropic.NewThinkingBlock("sig", "thinking"),
		anthropic.NewToolUseBlock("c1", json.RawMessage(`{}`), "search_echos"),
	}}
	p := &anthropicProvider{}
	_, msgs := p.buildMessages([]Message{
		{Role: RoleUser, Content: "q"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "c1", Name: "search_echos"}}, Native: native},
		{Role: RoleTool, ToolCallID: "c1", Content: "boom", IsError: true},
	})
	if len(msgs) != 3 {
		t.Fatalf("got %d messages", len(msgs))
	}
	if msgs[1].Content[0].OfThinking == nil || msgs[1].Content[0].OfThinking.Signature != "sig" {
		t.Fatalf("signed thinking block must be replayed unchanged, got %+v", msgs[1].Content)
	}
	if res := msgs[2].Content[0].OfToolResult; res == nil || !res.IsError.Value {
		t.Fatalf("failed tool result should carry is_error, got %+v", msgs[2].Content[0])
	}
}

func TestResponsesBuildParams_ReasoningModel(t *testing.T) {
	p := &openaiResponsesProvider{setting: model.AgentSetting{Model: "gpt-5"}}
	params, err := p.buildParams(Request{
		Messages: []Message{
			{Role: RoleUser, Content: "q"},
			{
				Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "c1", Name: "s", Args: json.RawMessage(`{}`)}},
				Native: respTurn{reasoning: []responses.ResponseReasoningItemParam{{ID: "rs_1", EncryptedContent: param.NewOpt("enc")}}},
			},
			{Role: RoleTool, ToolCallID: "c1", Content: "hit"},
		},
		Tools:      []ToolDef{{Name: "s", Parameters: json.RawMessage(`{"type":"object"}`)}},
		ToolChoice: ToolChoiceNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(params)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)

	if _, has := body["temperature"]; has {
		t.Fatalf("temperature must not be sent unless configured")
	}
	if body["tool_choice"] != "none" {
		t.Fatalf("tool_choice = %v, want none", body["tool_choice"])
	}
	if inc, _ := body["include"].([]any); len(inc) != 1 || inc[0] != "reasoning.encrypted_content" {
		t.Fatalf("include = %v, want encrypted reasoning", body["include"])
	}
	input := body["input"].([]any)
	if first := input[1].(map[string]any); first["type"] != "reasoning" || first["encrypted_content"] != "enc" {
		t.Fatalf("reasoning item must precede the call it produced, got %v", input[1])
	}
	if call := input[2].(map[string]any); call["type"] != "function_call" {
		t.Fatalf("function_call should follow its reasoning, got %v", input[2])
	}
}
