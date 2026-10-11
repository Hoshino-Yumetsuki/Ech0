// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"time"

	model "github.com/lin-snow/ech0/internal/model/setting"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
	Images     []ImagePart

	// IsError marks a tool result as a failure, so providers that can say so on
	// the wire (Anthropic's is_error) do, instead of the model having to infer
	// it from the wording.
	IsError bool

	// Native is the provider's own record of an assistant turn, carried back
	// verbatim on the next request of the same run. Some protocols refuse a
	// continuation that drops part of what the model produced — DeepSeek's
	// reasoning_content, Claude's signed thinking blocks, OpenAI's encrypted
	// reasoning items — and none of that belongs in a portable Message. The loop
	// never looks inside; a provider only reads back values of its own type, so
	// a value from any other provider is ignored rather than misread.
	Native any
}

type ImagePart struct {
	MediaType string
	Base64    string
	URL       string
}

type ToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
}

type ToolDef struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolEffect is how much a tool changes, and therefore what the loop has to do
// before running it. It is the tool's own declaration; a caller cannot name one
// and cannot widen the one a tool declares.
//
// EffectUnset is the zero value, and that ordering is the whole point. A tool
// added without a declared effect lands in the strictest branch, not the
// widest: it does not run. The mechanism is not what protects anyone here — the
// default is.
type ToolEffect uint8

const (
	EffectUnset ToolEffect = iota
	EffectRead
	EffectMutate
)

// Tool is one callable the model may reach.
//
// Exactly one body is set, and which one is decided by Effect: a read fills
// Run, a change fills Mutation. That is a deliberate asymmetry rather than one
// function with a flag — a mutation's body is unreachable except through the
// confirmation, and the type is what says so.
type Tool struct {
	Def    ToolDef
	Effect ToolEffect

	Run func(ctx context.Context, args json.RawMessage) (ToolOutput, error)

	Mutation *Mutation

	Interactive bool
}

// Mutation is a change split at the only place it can safely be split: after it
// is known and described, before it is made.
//
// Plan and Confirm are both supplied by the caller and both invoked by the
// loop, in that order, with Apply reachable only from between them. The tool
// never calls Confirm, which is why it cannot forget to — there is no code path
// through this type that reaches a change without a decision, and no prompt,
// context length or model behaviour can produce one.
type Mutation struct {
	Plan func(ctx context.Context, args json.RawMessage) (Plan, error)

	Confirm func(ctx context.Context, p Plan) (Decision, error)
}

// Plan is one worked-out change: what to show, and what to do if it is allowed.
type Plan struct {
	Prompt any

	Apply func(ctx context.Context) (ToolOutput, error)
}

// Decision is what a person said. Anything other than Approved is a no,
// including a Decision nobody filled in.
type Decision struct {
	Approved bool
	Refusal  string
}

type ToolOutput struct {
	Content string
	Meta    any
	Images  []ImagePart
}

// ToolChoice is how the model may use the tools a request declares.
type ToolChoice uint8

const (
	// ToolChoiceAuto lets the model decide. The zero value.
	ToolChoiceAuto ToolChoice = iota
	// ToolChoiceNone keeps the tools declared but forbids calling them. It is
	// how a run asks for a final answer once the history already holds tool
	// calls: dropping the declarations instead is rejected by Anthropic
	// ("requests which include tool_use or tool_result blocks must define
	// tools") and is off the documented path for the other protocols.
	ToolChoiceNone
)

type Request struct {
	Messages   []Message
	Tools      []ToolDef
	ToolChoice ToolChoice
}

type Response struct {
	Text string
}

type EventKind int

const (
	EventTextDelta EventKind = iota
	EventReasoningDelta
	EventToolCall
	EventDone
	EventError
)

type Event struct {
	Kind     EventKind
	Text     string
	ToolCall ToolCall
	Err      error
	// Native rides on EventDone: the provider's own record of the turn, to be
	// stored on the assistant Message the loop appends (see Message.Native).
	Native any
	// Truncated rides on EventDone: the output stopped at a length limit
	// instead of finishing, so its text is partial and any tool call in it may
	// be cut mid-argument.
	Truncated bool
	// Usage rides on EventDone: what the request cost, as the provider
	// reported it. Zero when the provider reported nothing.
	Usage Usage
}

// Usage is one request's token counts in the provider's own tokens.
type Usage struct {
	// InputTokens is everything the request carried, cached or not.
	InputTokens int
	// CachedTokens is the part of InputTokens read from the prompt cache.
	CachedTokens int
	OutputTokens int
}

func (u *Usage) add(o Usage) {
	u.InputTokens += o.InputTokens
	u.CachedTokens += o.CachedTokens
	u.OutputTokens += o.OutputTokens
}

type RunStrings struct {
	DedupNote       string
	UnknownTool     string
	ToolError       string
	ImageNote       string
	ContextTrimNote string
	TruncateNote    string
	Malformed       string
	// OutputTruncated closes an answer the model could not finish because it
	// ran into its output limit.
	OutputTruncated string
}

type RunRequest struct {
	Setting          model.AgentSetting
	Messages         []Message
	Tools            []Tool
	MaxRounds        int
	Strings          RunStrings
	Timeout          time.Duration
	MaxContextTokens int
}

type AgentEventKind int

const (
	AgentDelta AgentEventKind = iota
	AgentReasoning
	AgentSearching
	AgentToolResult
	AgentDone
	AgentError
)

type AgentEvent struct {
	Kind     AgentEventKind
	Text     string
	ToolName string
	ToolArgs json.RawMessage
	Meta     any
	Err      error
}
