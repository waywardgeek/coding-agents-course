package main

// The engine: an actor with a mailbox.
//
// This is the part Chapter 1 could not have written. There, the program read a
// line, sent a request, waited, printed the answer, and read the next line —
// and during the wait it was deaf. A hint is a message that arrives WHILE a
// request is outstanding, so a program shaped like that cannot receive one at
// any price.
//
// So: one goroutine owns the conversation and does nothing but select on two
// channels. Stdin is read by a second goroutine that only ever posts to the
// mailbox. The HTTP round trip happens on a third and posts its result back.
// Nothing else touches the log or the context, which is why there is not a
// single mutex in this file.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// CodeCommit is stamped into every RequestSent. Set it at build time:
//
//	go build -ldflags "-X main.CodeCommit=$(git rev-parse HEAD)"
var CodeCommit = "unknown"

// MaxTokens is this exercise's fixed output budget.
const MaxTokens = 1024

type inKind string

const (
	inUser      inKind = "user"
	inHint      inKind = "hint"
	inInterrupt inKind = "interrupt"
	inRedact    inKind = "redact"
	inEphemera  inKind = "ephemera"
	inDump      inKind = "dump"
	inJunk      inKind = "junk"
)

type inMsg struct {
	Kind        inKind
	Text        string
	Seq         int
	Instruction string
	Path        string
	Raw         string
}

type respMsg struct {
	Resp *APIResponse
	Err  error
}

// Engine owns one conversation.
type Engine struct {
	log    *Log
	ctx    *Context
	client *Client

	inbox chan inMsg
	resp  chan respMsg

	out  *bufio.Writer
	repl bool

	// owes records that a prompt is waiting for its answer. An interrupt
	// cancels the debt: a killed turn produces no reply.
	owes     bool
	lastText string
}

// NewEngine wires an engine to stdout.
func NewEngine(repl bool) *Engine {
	return &Engine{
		log:    NewLog(),
		ctx:    NewContext(),
		client: NewClient(),
		inbox:  make(chan inMsg, 64),
		resp:   make(chan respMsg, 4),
		out:    bufio.NewWriter(os.Stdout),
		repl:   repl,
	}
}

// record appends an event and folds it into the context. Those two things
// always happen together and in that order, which is what makes the context
// exactly "the state after playing the log".
func (e *Engine) record(ev Event) Event {
	ev = e.log.Append(ev)
	e.ctx.Apply(ev)
	return ev
}

func (e *Engine) emit(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	e.out.Write(b)
	e.out.WriteByte('\n')
	e.out.Flush()
}

func (e *Engine) say(format string, args ...any) {
	if e.repl {
		fmt.Fprintf(e.out, format+"\n", args...)
		e.out.Flush()
	}
}

func (e *Engine) ack() {
	if e.repl {
		return
	}
	e.emit(map[string]any{"ok": true})
}

// --- the mailbox ------------------------------------------------------------

// readStdin is the only thing that touches stdin. It never blocks the engine:
// whatever it reads goes into the mailbox and the engine picks it up whenever
// it next reaches its select, including while a request is in flight.
func (e *Engine) readStdin() {
	defer close(e.inbox)
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if e.repl {
			e.inbox <- inMsg{Kind: inUser, Text: line}
			continue
		}
		e.inbox <- parseDirective(line)
	}
}

// parseDirective turns one grader line into a mailbox message. The Chapter 1
// protocol is the {"user": ...} case; everything else is new this chapter.
func parseDirective(line string) inMsg {
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return inMsg{Kind: inJunk, Raw: line}
	}
	if raw, ok := m["user"]; ok {
		var s string
		_ = json.Unmarshal(raw, &s)
		return inMsg{Kind: inUser, Text: s}
	}
	if raw, ok := m["hint"]; ok {
		var s string
		_ = json.Unmarshal(raw, &s)
		return inMsg{Kind: inHint, Text: s}
	}
	if _, ok := m["interrupt"]; ok {
		return inMsg{Kind: inInterrupt}
	}
	if raw, ok := m["redact"]; ok {
		var n int
		_ = json.Unmarshal(raw, &n)
		return inMsg{Kind: inRedact, Seq: n}
	}
	if raw, ok := m["ephemera"]; ok {
		var eph struct {
			Instruction string `json:"instruction"`
			Text        string `json:"text"`
		}
		_ = json.Unmarshal(raw, &eph)
		return inMsg{Kind: inEphemera, Instruction: eph.Instruction, Text: eph.Text}
	}
	if raw, ok := m["dump"]; ok {
		var s string
		_ = json.Unmarshal(raw, &s)
		return inMsg{Kind: inDump, Path: s}
	}
	return inMsg{Kind: inJunk, Raw: line}
}

// --- the loop ---------------------------------------------------------------

// Run is the engine's whole life: select, fold, decide. It exits when stdin has
// closed and no turn is still open.
func (e *Engine) Run() {
	go e.readStdin()

	inbox := e.inbox
	for {
		if inbox == nil && e.ctx.Turn == Idle {
			break
		}
		select {
		case m, ok := <-inbox:
			if !ok {
				inbox = nil
				continue
			}
			e.handleInput(m)
		case r := <-e.resp:
			e.handleResponse(r)
		}
	}

	if !e.repl {
		e.emit(map[string]any{"usage": map[string]int{
			"input": e.ctx.Usage.Input, "output": e.ctx.Usage.Output,
		}})
	}
}

func (e *Engine) handleInput(m inMsg) {
	switch m.Kind {

	case inUser, inHint:
		// The same event either way. Whether it is a prompt or a hint is not
		// decided here — it is decided by the reducer, from the turn state it
		// arrives in. Classifying at capture time is wrong every time the
		// human types fast.
		before := e.ctx.Turn
		e.record(Event{Type: EvMessageReceived, Actor: &You, Parts: TextParts(m.Text)})
		switch before {
		case InFlight, ToolsPending:
			e.say("  (hint delivered mid-turn)")
			e.ack()
		default:
			if m.Kind == inHint {
				e.ack()
			}
			e.owes = !e.repl
			e.sendRequest()
		}

	case inInterrupt:
		e.record(Event{Type: EvInterrupted, Actor: &You})
		e.owes = false
		e.say("  (interrupted)")
		e.ack()

	case inRedact:
		e.record(Event{Type: EvRedacted, Actor: &You, TargetSeq: m.Seq})
		e.ack()

	case inEphemera:
		e.record(Event{
			Type: EvEphemeraSet, Actor: &Actor{Kind: ActorSystem, ID: "engine"},
			Instruction: m.Instruction, Parts: TextParts(m.Text),
		})
		e.ack()

	case inDump:
		data, err := e.log.Dump()
		if err == nil {
			err = os.WriteFile(m.Path, data, 0o644)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "dump: %v\n", err)
		}
		e.ack()

	default:
		fmt.Fprintf(os.Stderr, "ignoring unrecognized line: %s\n", m.Raw)
		e.ack()
	}
}

// sendRequest renders the current context and puts it on the wire.
//
// Order matters and is load-bearing: RENDER FIRST, then record RequestSent.
// The render is what carries a pending hint and the pending ephemera; the
// RequestSent event is what consumes them. Do it the other way round and the
// hint is delivered a round late, which is exactly the bug this chapter is
// about.
func (e *Engine) sendRequest() {
	req := Render(e.ctx, e.client.Model, MaxTokens)
	body, err := Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		return
	}
	e.record(Event{
		Type: EvRequestSent, Actor: &Actor{Kind: ActorSystem, ID: "engine"},
		CodeCommit: CodeCommit, RequestHash: HashRequest(body),
	})
	go func() {
		resp, err := e.client.Send(body)
		e.resp <- respMsg{Resp: resp, Err: err}
	}()
}

func (e *Engine) handleResponse(r respMsg) {
	if r.Err != nil {
		// Infrastructure failure: it changes state and contributes no content.
		// The model never hears about a 429 that was cured by a retry.
		e.record(Event{
			Type: EvErrorOccurred, Actor: &Actor{Kind: ActorSystem, ID: "engine"},
			Source: "Provider", Code: "request_failed", Message: r.Err.Error(),
			RelatedSeq: e.log.next - 1,
		})
		fmt.Fprintf(os.Stderr, "provider error: %v\n", r.Err)
		e.owes = false
		e.ctx.Turn = Idle
		return
	}

	var text strings.Builder
	for _, b := range r.Resp.Content {
		switch b.Type {
		case "text":
			text.WriteString(b.Text)
		case "thinking", "redacted_thinking":
			e.record(Event{
				Type: EvAssistantThink, Actor: &Model,
				Parts: TextParts(b.Thinking), Signature: b.Signature,
			})
		}
	}
	if text.Len() > 0 {
		e.lastText = text.String()
		e.record(Event{Type: EvAssistantMsg, Actor: &Model, Parts: TextParts(e.lastText)})
	}
	for _, b := range r.Resp.Content {
		if b.Type != "tool_use" {
			continue
		}
		args := b.Input
		if len(args) == 0 {
			args = json.RawMessage(`{}`)
		}
		e.record(Event{
			Type: EvToolCalled, Actor: &Model,
			CallID: b.ID, ToolName: b.Name, Args: args,
		})
	}

	// Read the state BEFORE folding in ResponseEnded: that is what tells us
	// whether this turn died while the response was in the air.
	wasInterrupted := e.ctx.Turn == Interrupted

	e.record(Event{
		Type: EvResponseEnded, Actor: &Model, StopReason: r.Resp.StopReason,
		Usage: &Usage{Input: r.Resp.Usage.InputTokens, Output: r.Resp.Usage.OutputTokens},
	})

	if wasInterrupted {
		// The calls that arrived are in the log. None of them is in
		// PendingTools, so none of them runs. The turn is over.
		calls := 0
		for _, b := range r.Resp.Content {
			if b.Type == "tool_use" {
				calls++
			}
		}
		e.say("  (turn ended; %d tool call(s) recorded but not executed)", calls)
		e.owes = false
		return
	}

	if e.ctx.Turn == ToolsPending {
		for _, id := range append([]string(nil), e.ctx.PendingTools...) {
			e.record(Event{
				Type: EvToolReturned, Actor: &Tool, CallID: id,
				Parts: TextParts(e.agentStatus()),
			})
		}
		e.sendRequest()
		return
	}

	// end_turn.
	if e.repl {
		e.say("assistant: %s", e.lastText)
	} else if e.owes {
		e.emit(map[string]any{"assistant": e.lastText})
		e.owes = false
	}
}

// agentStatus is the chapter's one tool, and it is not a tool system: no
// registry, no schema validation, no dispatch table. It reports the
// conversation's own state, which makes it deliberately self-referential —
// the agent's single capability is to look at the structure you just built.
//
// Its value must be a deterministic function of the log, because `render`
// replays the log and the two must agree byte for byte.
func (e *Engine) agentStatus() string {
	highest := 0
	for _, ev := range e.log.Events {
		if ev.Seq > highest {
			highest = ev.Seq
		}
	}
	b, _ := json.Marshal(map[string]any{
		"highest_seq": highest,
		"turn_state":  string(e.ctx.Turn),
	})
	return string(b)
}
