package main

// The renderer: Context -> Anthropic Messages API request. One way only.
//
// Everything Chapter 1 thought was architecture lives here, demoted to one
// backend's quirks: roles, strict alternation, block shapes, the rule that a
// tool_use must be answered by a tool_result, and the decision about what to do
// with a tool call that was interrupted before anyone could answer it.
//
// The context above does not know any of this, and that is the point. Put the
// hint logic in your HTTP code and you can serve exactly one model family.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
)

// Wire types are structs, never maps. Go randomizes map iteration order, and a
// renderer that emits keys in a different order every run is not a pure
// function of the log — it breaks replay, and later it will quietly destroy
// prefix caching.
type wireRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system,omitempty"`
	Tools     []wireTool    `json:"tools,omitempty"`
	Messages  []wireMessage `json:"messages"`
}

type wireTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type wireMessage struct {
	Role    string      `json:"role"`
	Content []wireBlock `json:"content"`
}

type wireBlock struct {
	Type string `json:"type"`

	Text string `json:"text,omitempty"`

	// tool_use
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// tool_result
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// RedactedStub is what the model sees in place of content that has been
// redacted. The log still has the original; this is the context's copy.
const RedactedStub = "[redacted]"

// InterruptedResult is what the renderer puts in place of the answer to a tool
// call that never ran.
//
// Two policies are defensible here: strip the dangling tool_use blocks, or
// synthesize a result saying what happened. We synthesize, because a model
// that knows it was cut off behaves better than one whose last action silently
// vanished from its own history.
const InterruptedResult = "interrupted by user"

func toolSchema() []wireTool {
	return []wireTool{{
		Name:        "agent_status",
		Description: "Report the conversation's own state: the highest event sequence number in the log, and the current turn state.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
	}}
}

type group struct {
	side    EntrySide
	entries []Entry
}

// groupEntries merges runs of same-side entries. Anthropic has two roles and
// requires them to alternate, so several of our entries become one message.
// That batching is renderer output, not truth.
func groupEntries(entries []Entry) []group {
	var gs []group
	for _, e := range entries {
		if n := len(gs); n > 0 && gs[n-1].side == e.Side {
			gs[n-1].entries = append(gs[n-1].entries, e)
			continue
		}
		gs = append(gs, group{side: e.Side, entries: []Entry{e}})
	}
	return gs
}

func entryBlocks(e Entry) []wireBlock {
	switch e.Kind {
	case KindToolUse:
		return []wireBlock{{
			Type: "tool_use", ID: e.CallID, Name: e.ToolName,
			Input: json.RawMessage(`{}`),
		}}
	case KindToolResult:
		content := e.Parts.Text()
		if e.Redacted {
			content = RedactedStub
		}
		return []wireBlock{{
			Type: "tool_result", ToolUseID: e.CallID, Content: content, IsError: e.IsError,
		}}
	default:
		text := e.Parts.Text()
		if e.Redacted {
			text = RedactedStub
		}
		return []wireBlock{{Type: "text", Text: text}}
	}
}

// Render turns a context into the request that would be sent right now.
//
// Two things that are NOT in the dialogue get attached here, and only here:
// the pending ephemera, and a hint that has not yet been carried. Both go at
// the end of the final user message, and the hint goes last — after the tool
// results it is meant to redirect. A hint placed ahead of those results reads
// as a comment on nothing, because the model has not yet seen what it is being
// steered away from.
func Render(c *Context, model string, maxTokens int) wireRequest {
	req := wireRequest{Model: model, MaxTokens: maxTokens, Tools: toolSchema()}

	answered := map[string]bool{}
	for _, e := range c.Dialogue {
		if e.Kind == KindToolResult {
			answered[e.CallID] = true
		}
	}

	var msgs []wireMessage
	appendUserBlocks := func(blocks []wireBlock, front bool) {
		if n := len(msgs); n > 0 && msgs[n-1].Role == "user" {
			if front {
				msgs[n-1].Content = append(append([]wireBlock{}, blocks...), msgs[n-1].Content...)
			} else {
				msgs[n-1].Content = append(msgs[n-1].Content, blocks...)
			}
			return
		}
		msgs = append(msgs, wireMessage{Role: "user", Content: blocks})
	}

	for _, g := range groupEntries(c.Dialogue) {
		var blocks []wireBlock
		var dangling []string
		for _, e := range g.entries {
			blocks = append(blocks, entryBlocks(e)...)
			if e.Kind == KindToolUse && !answered[e.CallID] {
				dangling = append(dangling, e.CallID)
			}
		}
		if g.side == SideUser {
			appendUserBlocks(blocks, false)
			continue
		}
		msgs = append(msgs, wireMessage{Role: "assistant", Content: blocks})

		// An interrupted turn leaves a tool_use with no answer. The context
		// records that truthfully; Anthropic will not accept it. That is
		// Anthropic's problem, and this is where it gets eaten.
		if len(dangling) > 0 {
			var synth []wireBlock
			for _, id := range dangling {
				synth = append(synth, wireBlock{
					Type: "tool_result", ToolUseID: id, Content: InterruptedResult,
				})
			}
			msgs = append(msgs, wireMessage{Role: "user", Content: synth})
		}
	}

	// Volatile data for this request only. It is never history, so it is
	// attached at render time and vanishes when the request is sent.
	if c.Ephemera != nil {
		text := c.Ephemera.Parts.Text()
		if c.Ephemera.Instruction != "" {
			text = c.Ephemera.Instruction + "\n" + text
		}
		appendUserBlocks([]wireBlock{{Type: "text", Text: text}}, false)
	}

	// The hint, carried exactly once, in the last position. After this request
	// the RequestSent event moves it into the dialogue, where it stays
	// forever: delivered once, remembered always.
	//
	// There are two legal carriages and which one is correct depends on the
	// model, which is the whole argument for this code living in the renderer.
	// HINT_CARRIAGE=system uses the API's first-class mid-turn system message;
	// anything else uses the original hack, a text block appended after the
	// tool results. The context above knows nothing about either.
	for _, h := range c.PendingHints {
		if hintCarriage() == "system" {
			msgs = append(msgs, wireMessage{
				Role:    "system",
				Content: []wireBlock{{Type: "text", Text: h.Parts.Text()}},
			})
			continue
		}
		appendUserBlocks([]wireBlock{{Type: "text", Text: h.Parts.Text()}}, false)
	}

	req.Messages = msgs
	return req
}

// hintCarriage selects how a pending hint reaches the model. Default is the
// hack, because it works everywhere.
func hintCarriage() string {
	if v := os.Getenv("HINT_CARRIAGE"); v != "" {
		return v
	}
	return "append"
}

// Marshal renders the request to the exact bytes that go on the wire.
func Marshal(req wireRequest) ([]byte, error) { return json.Marshal(req) }

// MarshalPretty renders the request for human eyes. `render` uses this: it is
// still a pure function of the log, so two runs must agree byte for byte.
func MarshalPretty(req wireRequest) ([]byte, error) { return json.MarshalIndent(req, "", "  ") }

// HashRequest is the evidence stored in RequestSent. Drift between what the
// code renders today and what it rendered then becomes detectable, and
// attributable, instead of being a story someone tells.
func HashRequest(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
