// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	model "github.com/lin-snow/ech0/internal/model/setting"
)

func TestChatStream_FinishReasonAndUsage(t *testing.T) {
	cases := []struct {
		name          string
		finish        string
		wantTruncated bool
		wantErr       error
	}{
		{"stop", "stop", false, nil},
		{"length", "length", true, nil},
		{"content filter", "content_filter", false, errContentFiltered},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var body string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				body = string(raw)
				w.Header().Set("Content-Type", "text/event-stream")
				for _, f := range []string{
					`{"choices":[{"index":0,"delta":{"content":"hi"}}]}`,
					`{"choices":[{"index":0,"delta":{},"finish_reason":"` + c.finish + `"}]}`,
					`{"choices":[],"usage":{"prompt_tokens":120,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":100}}}`,
				} {
					_, _ = io.WriteString(w, "data: "+f+"\n\n")
				}
				_, _ = io.WriteString(w, "data: [DONE]\n\n")
			}))
			defer srv.Close()

			p := &openaiProvider{setting: model.AgentSetting{Model: "m", ApiKey: "k", BaseURL: srv.URL + "/v1"}}
			ch, _ := p.Stream(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "q"}}})
			evs := drainProviderEvents(t, ch)
			last := evs[len(evs)-1]

			if !strings.Contains(body, `"include_usage":true`) {
				t.Fatalf("a stream should ask for usage, body %s", body)
			}
			if c.wantErr != nil {
				if last.Kind != EventError || !errors.Is(last.Err, c.wantErr) {
					t.Fatalf("want error %v, got %+v", c.wantErr, last)
				}
				return
			}
			if last.Kind != EventDone || last.Truncated != c.wantTruncated {
				t.Fatalf("done = %+v, want truncated=%v", last, c.wantTruncated)
			}
			if last.Usage != (Usage{InputTokens: 120, CachedTokens: 100, OutputTokens: 7}) {
				t.Fatalf("usage = %+v", last.Usage)
			}
		})
	}
}

func TestResponsesStream_IncompleteAndUsage(t *testing.T) {
	cases := []struct {
		name  string
		frame string
		check func(t *testing.T, last Event)
	}{
		{
			name: "completed reports usage",
			frame: `{"type":"response.completed","sequence_number":2,"response":{"id":"r","status":"completed","output":[],` +
				`"usage":{"input_tokens":90,"input_tokens_details":{"cached_tokens":64},"output_tokens":5,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":95}}}`,
			check: func(t *testing.T, last Event) {
				if last.Kind != EventDone || last.Truncated || last.Usage != (Usage{InputTokens: 90, CachedTokens: 64, OutputTokens: 5}) {
					t.Fatalf("done = %+v", last)
				}
			},
		},
		{
			name: "max_output_tokens truncates",
			frame: `{"type":"response.incomplete","sequence_number":2,"response":{"id":"r","status":"incomplete",` +
				`"incomplete_details":{"reason":"max_output_tokens"},"output":[]}}`,
			check: func(t *testing.T, last Event) {
				if last.Kind != EventDone || !last.Truncated {
					t.Fatalf("an incomplete response should end truncated, got %+v", last)
				}
			},
		},
		{
			name: "content_filter fails",
			frame: `{"type":"response.incomplete","sequence_number":2,"response":{"id":"r","status":"incomplete",` +
				`"incomplete_details":{"reason":"content_filter"},"output":[]}}`,
			check: func(t *testing.T, last Event) {
				if last.Kind != EventError || !errors.Is(last.Err, errContentFiltered) {
					t.Fatalf("a filtered response should fail, got %+v", last)
				}
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := newRespServer(t, func(w http.ResponseWriter) {
				writeSSE(w, `{"type":"response.output_text.delta","sequence_number":1,"delta":"hi"}`, c.frame)
			})
			ch, _ := srv.provider().Stream(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "q"}}})
			evs := drainProviderEvents(t, ch)
			c.check(t, evs[len(evs)-1])
		})
	}
}

func TestAnthropicStream_StopReasonUsageAndCacheTail(t *testing.T) {
	cases := []struct {
		stop          string
		wantTruncated bool
	}{
		{"end_turn", false},
		{"max_tokens", true},
		{"model_context_window_exceeded", true},
	}
	for _, c := range cases {
		t.Run(c.stop, func(t *testing.T) {
			var body map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &body)
				writeSSE(w,
					`{"type":"message_start","message":{"id":"m","type":"message","role":"assistant","model":"claude","content":[],`+
						`"usage":{"input_tokens":10,"cache_read_input_tokens":500,"cache_creation_input_tokens":40,"output_tokens":1}}}`,
					`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
					`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hi"}}`,
					`{"type":"content_block_stop","index":0}`,
					`{"type":"message_delta","delta":{"stop_reason":"`+c.stop+`"},"usage":{"output_tokens":9}}`,
					`{"type":"message_stop"}`,
				)
			}))
			defer srv.Close()

			p := &anthropicProvider{setting: model.AgentSetting{Model: "claude", ApiKey: "k", BaseURL: srv.URL}}
			ch, _ := p.Stream(context.Background(), Request{Messages: []Message{
				{Role: RoleSystem, Content: "S"},
				{Role: RoleUser, Content: "q"},
			}})
			evs := drainProviderEvents(t, ch)
			last := evs[len(evs)-1]

			if last.Kind != EventDone || last.Truncated != c.wantTruncated {
				t.Fatalf("done = %+v, want truncated=%v", last, c.wantTruncated)
			}
			if last.Usage != (Usage{InputTokens: 550, CachedTokens: 500, OutputTokens: 9}) {
				t.Fatalf("usage = %+v", last.Usage)
			}

			msgs := body["messages"].([]any)
			tail := msgs[len(msgs)-1].(map[string]any)["content"].([]any)
			if tail[len(tail)-1].(map[string]any)["cache_control"] == nil {
				t.Fatalf("the last block of a streamed request should carry a cache breakpoint: %v", tail)
			}
		})
	}
}

func TestAnthropicComplete_NoCacheTail(t *testing.T) {
	p := &anthropicProvider{setting: model.AgentSetting{Model: "claude"}}
	params := p.buildParams(Request{Messages: []Message{{Role: RoleUser, Content: "q"}}})
	if cc := params.Messages[0].Content[0].GetCacheControl(); cc != nil && cc.Type != "" {
		t.Fatalf("a one-shot request has no next round to read a tail cache")
	}
}
