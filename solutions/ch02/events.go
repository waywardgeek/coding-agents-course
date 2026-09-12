package main

// The event log: what happened, in the order it happened.
//
// Two rules govern everything in this file.
//
//	Append-only. Events are facts in the past tense. Nothing here is ever
//	edited — not even a redaction, which is itself a new event recording the
//	decision to hide something.
//
//	Self-contained. Given the current context and one event, the reducer in
//	context.go can compute the new context with no other information. That is
//	what makes replay possible, and it is why there is no Hint event: the
//	identical bytes are a prompt or a hint depending on the state they arrive
//	in, and only the reducer knows the state.

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// --- actors -----------------------------------------------------------------

// ActorKind is the sort of thing an actor is.
type ActorKind string

const (
	ActorHuman  ActorKind = "human"
	ActorAgent  ActorKind = "agent"
	ActorSystem ActorKind = "system"
	ActorTool   ActorKind = "tool"
)

// Actor is a persistent, stateful entity with a name that outlives any single
// message. Every event has one.
type Actor struct {
	Kind ActorKind `json:"kind"`
	ID   string    `json:"id"`
}

var (
	You   = Actor{Kind: ActorHuman, ID: "you"}
	Model = Actor{Kind: ActorAgent, ID: "assistant"}
	Tool  = Actor{Kind: ActorTool, ID: "agent_status"}
)

// --- content ----------------------------------------------------------------

// Part is one piece of a message's content. New media means a new Part kind,
// not a new event type and not a new method on the engine.
type Part interface{ isPart() }

// TextPart is text.
type TextPart struct{ Text string }

// ImagePart is an image, stored by hash in the blob table.
type ImagePart struct {
	MediaType string
	Hash      string
}

// AudioPart is audio, stored by hash in the blob table.
type AudioPart struct {
	MediaType string
	Hash      string
}

func (TextPart) isPart()  {}
func (ImagePart) isPart() {}
func (AudioPart) isPart() {}

// Parts is a content sequence that knows how to serialize itself.
type Parts []Part

// Text flattens the text parts. Non-text parts are invisible here, which is
// intentional: code that wants them must ask for them.
func (ps Parts) Text() string {
	var b strings.Builder
	for _, p := range ps {
		if t, ok := p.(TextPart); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

// TextParts builds a single-text-part content sequence.
func TextParts(s string) Parts { return Parts{TextPart{Text: s}} }

type partWire struct {
	Kind      string `json:"kind"`
	Text      string `json:"text,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Hash      string `json:"hash,omitempty"`
}

// MarshalJSON writes parts as tagged objects.
func (ps Parts) MarshalJSON() ([]byte, error) {
	out := make([]partWire, 0, len(ps))
	for _, p := range ps {
		switch v := p.(type) {
		case TextPart:
			out = append(out, partWire{Kind: "text", Text: v.Text})
		case ImagePart:
			out = append(out, partWire{Kind: "image", MediaType: v.MediaType, Hash: v.Hash})
		case AudioPart:
			out = append(out, partWire{Kind: "audio", MediaType: v.MediaType, Hash: v.Hash})
		default:
			return nil, fmt.Errorf("unknown part type %T", p)
		}
	}
	return json.Marshal(out)
}

// UnmarshalJSON reads parts back, refusing kinds it does not know.
func (ps *Parts) UnmarshalJSON(b []byte) error {
	var in []partWire
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	out := make(Parts, 0, len(in))
	for _, w := range in {
		switch w.Kind {
		case "text":
			out = append(out, TextPart{Text: w.Text})
		case "image":
			out = append(out, ImagePart{MediaType: w.MediaType, Hash: w.Hash})
		case "audio":
			out = append(out, AudioPart{MediaType: w.MediaType, Hash: w.Hash})
		default:
			return fmt.Errorf("unknown part kind %q", w.Kind)
		}
	}
	*ps = out
	return nil
}

// --- events -----------------------------------------------------------------

// EventType is the taxonomy. Adding a name here is additive; changing what an
// existing name MEANS is a major version bump, because it breaks replay.
type EventType string

const (
	// Dialogue events accumulate into the context.
	EvMessageReceived EventType = "MessageReceived"
	EvAssistantMsg    EventType = "AssistantMessage"
	EvAssistantThink  EventType = "AssistantThought"
	EvToolCalled      EventType = "ToolCalled"
	EvToolReturned    EventType = "ToolReturned"

	// Control events change what the context IS.
	EvInterrupted   EventType = "Interrupted"
	EvRequestSent   EventType = "RequestSent"
	EvResponseEnded EventType = "ResponseEnded"
	EvErrorOccurred EventType = "ErrorOccurred"
	EvEphemeraSet   EventType = "EphemeraSet"
	EvRedacted      EventType = "Redacted"
	EvModelChanged  EventType = "ModelChanged"
)

var knownTypes = map[EventType]bool{
	EvMessageReceived: true, EvAssistantMsg: true, EvAssistantThink: true,
	EvToolCalled: true, EvToolReturned: true, EvInterrupted: true,
	EvRequestSent: true, EvResponseEnded: true, EvErrorOccurred: true,
	EvEphemeraSet: true, EvRedacted: true, EvModelChanged: true,
}

// Usage is what a round trip cost.
type Usage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

// Event is one fact. The fields are a union: which ones are meaningful depends
// on Type, and the reducer is the only thing that needs to know.
//
// Seq orders the log and determines the state of the context. Time is metadata
// for humans — the GUI reads it, the reducer never does.
type Event struct {
	Seq  int       `json:"seq"`
	Type EventType `json:"type"`
	Time time.Time `json:"time"`

	Actor *Actor `json:"actor,omitempty"`
	Parts Parts  `json:"parts,omitempty"`

	// ToolCalled / ToolReturned
	CallID   string          `json:"call_id,omitempty"`
	ToolName string          `json:"tool_name,omitempty"`
	Args     json.RawMessage `json:"args,omitempty"`
	IsError  bool            `json:"is_error,omitempty"`

	// AssistantThought
	Signature string `json:"signature,omitempty"`

	// RequestSent
	CodeCommit  string `json:"code_commit,omitempty"`
	RequestHash string `json:"request_hash,omitempty"`

	// ResponseEnded
	StopReason string `json:"stop_reason,omitempty"`
	Usage      *Usage `json:"usage,omitempty"`

	// ErrorOccurred
	Source     string `json:"source,omitempty"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	RelatedSeq int    `json:"related_seq,omitempty"`

	// EphemeraSet
	Instruction string `json:"instruction,omitempty"`

	// Redacted
	TargetSeq int `json:"target_seq,omitempty"`

	// ModelChanged
	ModelName string `json:"model,omitempty"`
}

// --- the log ----------------------------------------------------------------

// Blobs is a content-addressed store. The same screenshot attached fifty times
// is stored once.
type Blobs struct{ data map[string][]byte }

// NewBlobs makes an empty blob table.
func NewBlobs() *Blobs { return &Blobs{data: map[string][]byte{}} }

// Put stores bytes and returns their hash.
func (b *Blobs) Put(hash string, data []byte) { b.data[hash] = data }

// Get retrieves bytes by hash.
func (b *Blobs) Get(hash string) ([]byte, bool) { v, ok := b.data[hash]; return v, ok }

// Len reports how many distinct blobs are held.
func (b *Blobs) Len() int { return len(b.data) }

// Log is the append-only record. It plus the blob table are the whole
// conversation.
type Log struct {
	Events []Event
	Blobs  *Blobs
	next   int
}

// NewLog makes an empty log whose first Seq is 1.
func NewLog() *Log { return &Log{Blobs: NewBlobs(), next: 1} }

// Append stamps an event with the next Seq and the wall clock, and records it.
// Seq is monotonic and never reused.
func (l *Log) Append(e Event) Event {
	e.Seq = l.next
	l.next++
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	l.Events = append(l.Events, e)
	return e
}

// Find returns the event with the given Seq.
func (l *Log) Find(seq int) (Event, bool) {
	for _, e := range l.Events {
		if e.Seq == seq {
			return e, true
		}
	}
	return Event{}, false
}

// Dump writes the log as JSON-lines, one event per line, ascending Seq.
func (l *Log) Dump() ([]byte, error) {
	events := append([]Event(nil), l.Events...)
	sort.SliceStable(events, func(i, j int) bool { return events[i].Seq < events[j].Seq })
	var b strings.Builder
	for _, e := range events {
		line, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return []byte(b.String()), nil
}

// LoadLog parses a JSON-lines log.
//
// An unknown event type is a hard failure, named loudly. Skipping it would
// produce a context that is subtly wrong while claiming to be an exact replay,
// which is worse than not replaying at all.
func LoadLog(data []byte) (*Log, error) {
	l := NewLog()
	for i, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("log line %d: %w", i+1, err)
		}
		if !knownTypes[e.Type] {
			return nil, fmt.Errorf(
				"log line %d: unknown event type %q. This log was written by a newer binary; "+
					"refusing to replay it rather than silently dropping an event", i+1, e.Type)
		}
		if e.Seq <= 0 {
			return nil, fmt.Errorf("log line %d: missing or non-positive seq", i+1)
		}
		l.Events = append(l.Events, e)
		if e.Seq >= l.next {
			l.next = e.Seq + 1
		}
	}
	sort.SliceStable(l.Events, func(i, j int) bool { return l.Events[i].Seq < l.Events[j].Seq })
	return l, nil
}
