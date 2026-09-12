package fakeanthropic

// ToolServer is the Chapter 2 fake: a Messages API stand-in that can drive a
// tool loop, hold a reply open on demand, and record the block-level structure
// of everything it was sent.
//
// It follows the same three rules as the Chapter 1 Server (record everything,
// forgive at the transport and be strict in the record, stay deterministic)
// and adds one more:
//
//  4. It can BLOCK. A hint is a message that arrives while the program is
//     waiting on an HTTP request it has already sent. The only way to test
//     that honestly is for the server to stop, let the grader type something,
//     and then answer. That is what Hook is for.

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// ToolName is the single tool this chapter's exercise implements.
const ToolName = "agent_status"

// Block is one content block of a message, flattened for inspection.
type Block struct {
	MsgIdx    int    // index into the request's messages array
	Role      string // role of the containing message
	BlockIdx  int    // index within that message's content
	Type      string // "text", "tool_use", "tool_result", ...
	Text      string // text for text blocks; stringified content for tool_result
	Name      string // tool_use: tool name
	ToolUseID string // tool_use: id; tool_result: tool_use_id
	IsError   bool   // tool_result: is_error
}

// blocks parses a message's content into blocks. A bare string becomes a
// single text block, which is what the Chapter 1 shape degrades to.
func blocks(m Message, msgIdx int) []Block {
	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return []Block{{MsgIdx: msgIdx, Role: m.Role, BlockIdx: 0, Type: "text", Text: s}}
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(m.Content, &raw); err != nil {
		return []Block{{MsgIdx: msgIdx, Role: m.Role, BlockIdx: 0, Type: "unknown", Text: string(m.Content)}}
	}
	out := make([]Block, 0, len(raw))
	for i, r := range raw {
		var b struct {
			Type      string          `json:"type"`
			Text      string          `json:"text"`
			Name      string          `json:"name"`
			ID        string          `json:"id"`
			ToolUseID string          `json:"tool_use_id"`
			IsError   bool            `json:"is_error"`
			Content   json.RawMessage `json:"content"`
			Input     json.RawMessage `json:"input"`
		}
		if err := json.Unmarshal(r, &b); err != nil {
			out = append(out, Block{MsgIdx: msgIdx, Role: m.Role, BlockIdx: i, Type: "unparseable", Text: string(r)})
			continue
		}
		blk := Block{
			MsgIdx: msgIdx, Role: m.Role, BlockIdx: i,
			Type: b.Type, Text: b.Text, Name: b.Name, IsError: b.IsError,
		}
		if b.ToolUseID != "" {
			blk.ToolUseID = b.ToolUseID
		} else {
			blk.ToolUseID = b.ID
		}
		if blk.Type == "tool_result" {
			blk.Text = flattenContent(b.Content)
		}
		if blk.Type == "tool_use" && blk.Text == "" {
			blk.Text = string(b.Input)
		}
		out = append(out, blk)
	}
	return out
}

// flattenContent renders a tool_result's content (string or block array) as
// plain text, so the grader can look for a payload without caring which legal
// encoding the student chose.
func flattenContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.Text)
		}
		return b.String()
	}
	return string(raw)
}

// ToolRequest is the subset of the request body Chapter 2 cares about.
type ToolRequest struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	System    json.RawMessage   `json:"system"`
	Messages  []Message         `json:"messages"`
	Tools     []json.RawMessage `json:"tools"`
	Stream    bool              `json:"stream"`
}

// SystemText flattens the system field to plain text.
func (r ToolRequest) SystemText() string {
	if len(r.System) == 0 {
		return ""
	}
	return Message{Content: r.System}.Text()
}

// ToolNames lists the tool names declared in the request.
func (r ToolRequest) ToolNames() []string {
	var out []string
	for _, t := range r.Tools {
		var td struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(t, &td); err == nil && td.Name != "" {
			out = append(out, td.Name)
		}
	}
	return out
}

// ToolRecord is everything the fake saw and did for one Chapter 2 request.
type ToolRecord struct {
	Seq        int
	Header     http.Header
	RawBody    []byte
	Body       ToolRequest
	ParseErr   string
	Violations []string

	Blocks []Block // every content block, in request order

	Turn         string // name of the scripted turn this belonged to
	Continuation bool   // carried tool_result blocks (i.e. mid-turn)
	Iteration    int    // 1-based index within the current tool loop, 0 if not in one

	Reply      string // assistant text served back (may be empty)
	ToolCallID string // id of the tool_use block served back, if any
	StopReason string
	Usage      Usage
}

// AllText returns every scrap of text in the request, in request order: the
// system field first, then each block.
func (r ToolRecord) AllText() string {
	var b strings.Builder
	b.WriteString(r.Body.SystemText())
	b.WriteString("\n")
	for _, blk := range r.Blocks {
		b.WriteString(blk.Text)
		b.WriteString("\n")
	}
	return b.String()
}

// Count returns how many times needle appears across the request's text.
func (r ToolRecord) Count(needle string) int {
	return strings.Count(r.AllText(), needle)
}

// TurnKind selects what the fake does when a new turn begins.
type TurnKind int

const (
	// TurnText answers with a plain text message and stop_reason end_turn.
	TurnText TurnKind = iota
	// TurnLoop answers with a tool_use block, and keeps answering with
	// tool_use blocks for every continuation until StopWhen is satisfied.
	TurnLoop
)

// Turn is one scripted turn of the Chapter 2 session.
type Turn struct {
	Name string
	Kind TurnKind

	// Prompt is the text the grader types to begin this turn. The fake uses
	// it to tell a new turn from a continuation of the current one.
	//
	// The obvious test — "does this request contain tool_result blocks?" — is
	// wrong, and wrong in an instructive way: once a tool loop has happened,
	// EVERY later request carries those results forever, because history is
	// resent in full. A synthesized "interrupted by user" result makes it
	// worse still. Matching the prompt in the final user message is the only
	// discriminator that survives both.
	Prompt string

	// Text is the assistant reply for TurnText, and the final reply that ends
	// a TurnLoop.
	Text string

	// GateAt, for TurnLoop, is the 1-based continuation index at which the
	// fake calls Hook before answering — the moment the program is provably
	// blocked on the network.
	GateAt int

	// StopWhen, for TurnLoop, ends the loop when it returns true for an
	// incoming continuation request. Nil means "never stop voluntarily".
	StopWhen func(r ToolRecord) bool

	// MaxIterations caps a TurnLoop so a submission that never satisfies
	// StopWhen fails with a diagnosis instead of hanging.
	MaxIterations int
}

// HookInfo describes the moment at which the fake paused.
type HookInfo struct {
	Turn      string
	Iteration int
	Seq       int
}

// ToolServer is the Chapter 2 fake API.
type ToolServer struct {
	Turns []Turn

	// Hook, if set, is called synchronously before the response to a gated
	// request is written. The handler blocks until it returns.
	Hook func(HookInfo)

	mu       sync.Mutex
	records  []ToolRecord
	turnIdx  int
	inLoop   bool
	loopTurn int // index into Turns of the running loop
	loopIter int
	capped   map[string]bool
	postStop map[string]bool
	nextID   int

	httpSrv  *http.Server
	listener net.Listener
	baseURL  string
}

// Start binds an ephemeral localhost port and begins serving.
func (s *ToolServer) Start() (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	s.capped = map[string]bool{}
	s.postStop = map[string]bool{}
	s.listener = ln
	s.baseURL = "http://" + ln.Addr().String()
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)
	s.httpSrv = &http.Server{Handler: mux}
	go func() { _ = s.httpSrv.Serve(ln) }()
	return s.baseURL, nil
}

// Close shuts the server down.
func (s *ToolServer) Close() {
	if s.httpSrv != nil {
		_ = s.httpSrv.Close()
	}
}

// Records returns a snapshot of every request served, in arrival order.
func (s *ToolServer) Records() []ToolRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ToolRecord, len(s.records))
	copy(out, s.records)
	return out
}

// Capped reports whether the named loop hit its iteration cap — meaning the
// submission never satisfied StopWhen.
func (s *ToolServer) Capped(turn string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.capped[turn]
}

// ContinuedAfterStop reports whether the program sent another tool_result for
// a turn the fake had already finished — the signature of a tool call that was
// executed when it should only have been recorded.
func (s *ToolServer) ContinuedAfterStop(turn string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.postStop[turn]
}

// TotalUsage sums the usage reported across every request served.
func (s *ToolServer) TotalUsage() Usage {
	s.mu.Lock()
	defer s.mu.Unlock()
	var t Usage
	for _, r := range s.records {
		t.InputTokens += r.Usage.InputTokens
		t.OutputTokens += r.Usage.OutputTokens
	}
	return t
}

func (s *ToolServer) handle(w http.ResponseWriter, r *http.Request) {
	rec := ToolRecord{Header: r.Header.Clone()}

	buf := make([]byte, 0, 64*1024)
	tmp := make([]byte, 32*1024)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	rec.RawBody = buf

	if r.Method != http.MethodPost {
		rec.Violations = append(rec.Violations,
			fmt.Sprintf("method is %s, the Messages API requires POST", r.Method))
	}
	if !strings.HasSuffix(r.URL.Path, "/messages") {
		rec.Violations = append(rec.Violations,
			fmt.Sprintf("path %q does not end in /messages", r.URL.Path))
	}
	if r.Header.Get("x-api-key") == "" {
		rec.Violations = append(rec.Violations, "missing x-api-key header")
	}
	if r.Header.Get("anthropic-version") == "" {
		rec.Violations = append(rec.Violations, "missing anthropic-version header")
	}

	if err := json.Unmarshal(buf, &rec.Body); err != nil {
		rec.ParseErr = err.Error()
		rec.Violations = append(rec.Violations, "request body is not valid JSON: "+err.Error())
	} else {
		for i, m := range rec.Body.Messages {
			rec.Blocks = append(rec.Blocks, blocks(m, i)...)
		}
		rec.Violations = append(rec.Violations, legality(rec)...)
	}

	// --- decide the response -------------------------------------------
	s.mu.Lock()
	rec.Seq = len(s.records) + 1

	// A new turn is a request whose FINAL user message carries the prompt the
	// grader just typed. Anything else that arrives while a loop is running is
	// a continuation of it.
	isNew := false
	if s.turnIdx < len(s.Turns) {
		if p := s.Turns[s.turnIdx].Prompt; p != "" && strings.Contains(finalUserText(rec.Blocks), p) {
			isNew = true
		}
	}
	rec.Continuation = !isNew

	var (
		turn     Turn
		gate     bool
		kind     TurnKind
		endLoop  bool
		haveTurn bool
	)

	switch {
	case isNew:
		turn = s.Turns[s.turnIdx]
		s.turnIdx++
		rec.Turn = turn.Name
		haveTurn = true
		kind = turn.Kind
		if kind == TurnLoop {
			s.inLoop = true
			s.loopTurn = s.turnIdx - 1
			s.loopIter = 0
			rec.Iteration = 0
		} else {
			s.inLoop = false
		}

	case s.inLoop:
		s.loopIter++
		rec.Iteration = s.loopIter
		turn = s.Turns[s.loopTurn]
		rec.Turn = turn.Name
		haveTurn = true
		kind = TurnLoop
		gate = turn.GateAt > 0 && s.loopIter == turn.GateAt
		switch {
		case turn.StopWhen != nil && turn.StopWhen(rec):
			endLoop = true
		case turn.MaxIterations > 0 && s.loopIter >= turn.MaxIterations:
			endLoop = true
			s.capped[turn.Name] = true
		}

	default:
		// Neither a scripted prompt nor a continuation of a running loop. The
		// usual cause is a tool call that was answered after an interrupt.
		name := "(none)"
		if s.turnIdx > 0 && s.turnIdx <= len(s.Turns) {
			name = s.Turns[s.turnIdx-1].Name
		}
		s.postStop[name] = true
		rec.Turn = name
		rec.Violations = append(rec.Violations,
			"request is neither the next scripted prompt nor a continuation of a running tool loop")
		endLoop = true
		haveTurn = true
		turn = Turn{Text: "(the fake had already finished that turn)"}
		kind = TurnText
	}

	if !haveTurn {
		rec.Violations = append(rec.Violations,
			"request arrived after the scripted session was exhausted")
		turn = Turn{Text: "(the fake server has run out of scripted turns)"}
		kind = TurnText
	}

	if endLoop {
		s.inLoop = false
	}

	var reply, toolID string
	stop := "end_turn"
	if kind == TurnLoop && !endLoop {
		s.nextID++
		toolID = fmt.Sprintf("toolu_fake_%03d", s.nextID)
		stop = "tool_use"
		reply = ""
	} else {
		reply = turn.Text
	}
	rec.Reply, rec.ToolCallID, rec.StopReason = reply, toolID, stop

	in := countTokens(rec.Body.SystemText())
	for _, b := range rec.Blocks {
		in += countTokens(b.Text)
	}
	out := countTokens(reply)
	if toolID != "" {
		out += countTokens(ToolName)
	}
	rec.Usage = Usage{InputTokens: in, OutputTokens: out}

	hookInfo := HookInfo{Turn: rec.Turn, Iteration: rec.Iteration, Seq: rec.Seq}
	s.records = append(s.records, rec)
	s.mu.Unlock()

	// Hold the reply open. The grader uses this window to type at a program
	// that is provably mid-request.
	if gate && s.Hook != nil {
		s.Hook(hookInfo)
	}

	content := []map[string]any{}
	if reply != "" {
		content = append(content, map[string]any{"type": "text", "text": reply})
	}
	if toolID != "" {
		content = append(content, map[string]any{
			"type": "tool_use", "id": toolID, "name": ToolName, "input": map[string]any{},
		})
	}
	if len(content) == 0 {
		content = append(content, map[string]any{"type": "text", "text": ""})
	}

	resp := map[string]any{
		"id":            fmt.Sprintf("msg_fake2_%03d", rec.Seq),
		"type":          "message",
		"role":          "assistant",
		"model":         "claude-fake-course-2",
		"content":       content,
		"stop_reason":   stop,
		"stop_sequence": nil,
		"usage":         rec.Usage,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func hasToolResult(bs []Block) bool {
	for _, b := range bs {
		if b.Type == "tool_result" {
			return true
		}
	}
	return false
}

// finalUserText concatenates the text blocks of the last non-assistant message
// in the request — the place a freshly typed prompt has to be.
func finalUserText(bs []Block) string {
	last := -1
	for _, b := range bs {
		if b.Role != "assistant" && b.MsgIdx > last {
			last = b.MsgIdx
		}
	}
	if last < 0 {
		return ""
	}
	var out strings.Builder
	for _, b := range bs {
		if b.MsgIdx == last && b.Type == "text" {
			out.WriteString(b.Text)
			out.WriteString("\n")
		}
	}
	return out.String()
}

// legality reports the Anthropic-specific structural rules a rendered request
// must obey. These are recorded, never enforced: a malformed request still
// gets a well-formed 200.
func legality(rec ToolRecord) []string {
	var v []string
	if rec.Body.Model == "" {
		v = append(v, "model field is empty")
	}
	if rec.Body.MaxTokens <= 0 {
		v = append(v, "max_tokens is missing or non-positive")
	}
	if rec.Body.Stream {
		v = append(v, "stream:true — this exercise is non-streaming")
	}
	if len(rec.Body.Messages) == 0 {
		v = append(v, "messages array is empty")
		return v
	}

	// Roles. "system" is tolerated as a message role because one of the two
	// legal hint carriages is a mid-turn system message; every other role is
	// a fault.
	var dialogue []Message
	for i, m := range rec.Body.Messages {
		switch m.Role {
		case "user", "assistant":
			dialogue = append(dialogue, m)
		case "system":
			// allowed, and excluded from the alternation rule
		default:
			v = append(v, fmt.Sprintf("messages[%d].role is %q", i, m.Role))
		}
	}
	for i := 1; i < len(dialogue); i++ {
		if dialogue[i].Role == dialogue[i-1].Role {
			v = append(v, fmt.Sprintf(
				"two consecutive %q messages (ignoring system) — user and assistant must alternate",
				dialogue[i].Role))
			break
		}
	}
	if len(dialogue) > 0 {
		if dialogue[0].Role != "user" {
			v = append(v, "conversation must begin with a user message")
		}
		if last := dialogue[len(dialogue)-1]; last.Role != "user" {
			v = append(v, "the final message must be the user's")
		}
	}

	// Tool bookkeeping: every tool_use must be answered by a tool_result, and
	// every tool_result must answer a tool_use. This is the rule that an
	// interrupt makes interesting.
	uses := map[string]int{}
	results := map[string]int{}
	for _, b := range rec.Blocks {
		switch b.Type {
		case "tool_use":
			uses[b.ToolUseID]++
		case "tool_result":
			results[b.ToolUseID]++
		}
	}
	for id := range uses {
		if results[id] == 0 {
			v = append(v, fmt.Sprintf(
				"tool_use %q has no matching tool_result — Anthropic rejects a dangling tool call; "+
					"an interrupted turn still has to be rendered into legal JSON", id))
		}
	}
	for id := range results {
		if uses[id] == 0 {
			v = append(v, fmt.Sprintf("tool_result %q answers no tool_use in this request", id))
		}
	}
	return v
}
