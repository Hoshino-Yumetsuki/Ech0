// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

// anthropicMaxTokens is the output ceiling the API requires on every request.
// It is a ceiling, not a target, and a low one truncates long summaries and
// the thinking current models do by default before answering.
const anthropicMaxTokens = 16_000

var errAnthropicRefusal = errors.New("anthropic: 模型拒绝回答该请求（stop_reason=refusal）")

// anthropicTurn is an assistant turn exactly as the API returned it. Replaying
// it verbatim is the documented way to continue a tool loop: it keeps the
// signed thinking blocks current models produce by default, which the API
// requires back unchanged on the turn that called the tools.
type anthropicTurn struct {
	content []anthropic.ContentBlockParamUnion
}

type anthropicProvider struct {
	setting model.AgentSetting
}

func (p *anthropicProvider) newClient() anthropic.Client {
	opts := []option.RequestOption{option.WithAPIKey(p.setting.ApiKey)}
	if p.setting.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(p.setting.BaseURL))
	}
	return anthropic.NewClient(opts...)
}

func (p *anthropicProvider) buildParams(req Request) anthropic.MessageNewParams {
	systemBlocks, msgs := p.buildMessages(req.Messages)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.setting.Model),
		MaxTokens: anthropicMaxTokens,
		Messages:  msgs,
	}
	if len(systemBlocks) > 0 {
		systemBlocks[len(systemBlocks)-1].CacheControl = anthropic.NewCacheControlEphemeralParam()
		params.System = systemBlocks
	}
	if tools := p.buildTools(req.Tools); len(tools) > 0 {
		params.Tools = tools
		if req.ToolChoice == ToolChoiceNone {
			params.ToolChoice = anthropic.ToolChoiceUnionParam{OfNone: &anthropic.ToolChoiceNoneParam{}}
		}
	}
	if t := p.setting.Temperature; t != nil {
		params.Temperature = param.NewOpt(*t)
	}
	return params
}

// markCacheTail puts a cache breakpoint on the request's last block.
//
// The breakpoint on the system prompt covers only the tools and the prompt. A
// tool loop resends the whole conversation every round, and without a
// breakpoint at its end each round pays full price for everything the round
// before already sent. Placed on the last block, it moves forward with the
// conversation: each round reads the previous round's prefix from the cache and
// writes its own for the next one. Streaming is the loop's path; a one-shot
// Complete has no next round to read it, and would pay the cache write for
// nothing.
func markCacheTail(msgs []anthropic.MessageParam) {
	if len(msgs) == 0 {
		return
	}
	blocks := msgs[len(msgs)-1].Content
	if len(blocks) == 0 {
		return
	}
	if cc := blocks[len(blocks)-1].GetCacheControl(); cc != nil {
		*cc = anthropic.NewCacheControlEphemeralParam()
	}
}

func (p *anthropicProvider) generate(ctx context.Context, req Request) (string, []ToolCall, error) {
	client := p.newClient()
	resp, err := client.Messages.New(ctx, p.buildParams(req))
	if err != nil {
		return "", nil, err
	}
	if resp.StopReason == anthropic.StopReasonRefusal {
		return "", nil, errAnthropicRefusal
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return text.String(), toolCallsFromContent(resp.Content), nil
}

func toolCallsFromContent(content []anthropic.ContentBlockUnion) []ToolCall {
	var calls []ToolCall
	for _, block := range content {
		if block.Type != "tool_use" {
			continue
		}
		args := json.RawMessage(block.Input)
		if len(args) == 0 {
			args = json.RawMessage("{}")
		}
		calls = append(calls, ToolCall{
			ID:   block.ID,
			Name: block.Name,
			Args: args,
		})
	}
	return calls
}

func (p *anthropicProvider) buildMessages(in []Message) ([]anthropic.TextBlockParam, []anthropic.MessageParam) {
	var (
		systemBlocks []anthropic.TextBlockParam
		msgs         []anthropic.MessageParam
		pendingTools []anthropic.ContentBlockParamUnion
	)

	flush := func() {
		if len(pendingTools) > 0 {
			msgs = append(msgs, anthropic.NewUserMessage(pendingTools...))
			pendingTools = nil
		}
	}

	for _, m := range in {
		switch m.Role {
		case RoleSystem:
			flush()
			systemBlocks = append(systemBlocks, anthropic.TextBlockParam{Text: m.Content})
		case RoleAssistant:
			flush()
			if turn, ok := m.Native.(anthropicTurn); ok && len(turn.content) > 0 {
				msgs = append(msgs, anthropic.NewAssistantMessage(turn.content...))
				continue
			}
			var blocks []anthropic.ContentBlockParamUnion
			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, json.RawMessage(tc.Args), tc.Name))
			}
			if len(blocks) > 0 {
				msgs = append(msgs, anthropic.NewAssistantMessage(blocks...))
			}
		case RoleTool:
			pendingTools = append(pendingTools, anthropic.NewToolResultBlock(m.ToolCallID, m.Content, m.IsError))
		default:
			flush()
			msgs = append(msgs, anthropic.NewUserMessage(userBlocks(m)...))
		}
	}
	flush()
	return systemBlocks, msgs
}

func userBlocks(m Message) []anthropic.ContentBlockParamUnion {
	var blocks []anthropic.ContentBlockParamUnion
	if m.Content != "" {
		blocks = append(blocks, anthropic.NewTextBlock(m.Content))
	}
	for _, img := range m.Images {
		switch {
		case img.Base64 != "":
			blocks = append(blocks, anthropic.NewImageBlockBase64(img.MediaType, img.Base64))
		case img.URL != "":
			blocks = append(blocks, anthropic.NewImageBlock(anthropic.URLImageSourceParam{URL: img.URL}))
		}
	}
	if len(blocks) == 0 {
		blocks = append(blocks, anthropic.NewTextBlock(m.Content))
	}
	return blocks
}

func (p *anthropicProvider) buildTools(defs []ToolDef) []anthropic.ToolUnionParam {
	if len(defs) == 0 {
		return nil
	}
	tools := make([]anthropic.ToolUnionParam, 0, len(defs))
	for _, d := range defs {
		var parsed struct {
			Properties any      `json:"properties"`
			Required   []string `json:"required"`
		}
		_ = json.Unmarshal(d.Parameters, &parsed)

		t := anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Properties: parsed.Properties,
				Required:   parsed.Required,
			},
			d.Name,
		)
		if t.OfTool != nil && d.Description != "" {
			t.OfTool.Description = param.NewOpt(d.Description)
		}
		tools = append(tools, t)
	}
	return tools
}

func (p *anthropicProvider) Complete(ctx context.Context, req Request) (Response, error) {
	text, _, err := p.generate(ctx, req)
	if err != nil {
		return Response{}, err
	}
	if text == "" {
		return Response{}, errors.New("anthropic: empty text response")
	}
	return Response{Text: text}, nil
}

func (p *anthropicProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event)
	go p.stream(ctx, req, ch)
	return ch, nil
}

func (p *anthropicProvider) stream(ctx context.Context, req Request, ch chan<- Event) {
	defer close(ch)

	params := p.buildParams(req)
	markCacheTail(params.Messages)
	client := p.newClient()
	stream := client.Messages.NewStreaming(ctx, params)

	var acc anthropic.Message
	for stream.Next() {
		event := stream.Current()
		if err := acc.Accumulate(event); err != nil {
			send(ctx, ch, Event{Kind: EventError, Err: err})
			return
		}

		if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			switch d := delta.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				if d.Text != "" && !send(ctx, ch, Event{Kind: EventTextDelta, Text: d.Text}) {
					return
				}
			case anthropic.ThinkingDelta:
				if d.Thinking != "" && !send(ctx, ch, Event{Kind: EventReasoningDelta, Text: d.Thinking}) {
					return
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		send(ctx, ch, Event{Kind: EventError, Err: err})
		return
	}
	if acc.StopReason == anthropic.StopReasonRefusal {
		send(ctx, ch, Event{Kind: EventError, Err: errAnthropicRefusal})
		return
	}
	for _, tc := range toolCallsFromContent(acc.Content) {
		if !send(ctx, ch, Event{Kind: EventToolCall, ToolCall: tc}) {
			return
		}
	}
	send(ctx, ch, Event{
		Kind:      EventDone,
		Native:    anthropicTurn{content: acc.ToParam().Content},
		Truncated: acc.StopReason == anthropic.StopReasonMaxTokens || acc.StopReason == anthropic.StopReasonModelContextWindowExceeded,
		Usage:     anthropicUsage(acc.Usage),
	})
}

// anthropicUsage totals the input the way the other protocols report it.
// Anthropic splits it three ways, and input_tokens alone is only the part
// after the last cache breakpoint.
func anthropicUsage(u anthropic.Usage) Usage {
	return Usage{
		InputTokens:  int(u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens),
		CachedTokens: int(u.CacheReadInputTokens),
		OutputTokens: int(u.OutputTokens),
	}
}
