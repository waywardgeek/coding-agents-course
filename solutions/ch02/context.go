package main

// The context: vendor-independent state, defined as the result of playing
// every event in the log.
//
// No Anthropic rule appears in this file. Roles, alternation, block shapes and
// the handling of a dangling tool call all live in render.go, because they are
// one vendor's quirks and not facts about the conversation.
//
// The reducer is TOTAL: every (state, event) pair has a defined outcome, and
// identity is a legal outcome. Two interrupts in a row is a no-op, not an
// error. A race is an ordering, not corruption.

import "fmt"

// TurnState is where the conversation is in its current turn.
type TurnState string

const (
	// Idle — no turn open.
	Idle TurnState = "Idle"
	// InputPending — input has arrived and no request has been sent. NOT a
	// commitment to respond: in a room, the engine decides whether to speak.
	InputPending TurnState = "InputPending"
	// InFlight — a request is out, awaiting response events.
	InFlight TurnState = "InFlight"
	// ToolsPending — the response ended in tool_use; results are accumulating.
	ToolsPending TurnState = "ToolsPending"
	// Interrupted — the turn was killed. Late events APPEND but execute
	// nothing. This must be a state: without it, a tool call arriving after an
	// interrupt is indistinguishable from a live turn's, and a self-contained
	// reducer would start executing it.
	Interrupted TurnState = "Interrupted"
)

// EntrySide is which side of the dialogue an entry belongs to. It is not a
// vendor role — the renderer decides what a side is called on the wire.
type EntrySide string

const (
	SideUser      EntrySide = "user"
	SideAssistant EntrySide = "assistant"
)

// EntryKind is what an entry carries.
type EntryKind string

const (
	KindText       EntryKind = "text"
	KindToolUse    EntryKind = "tool_use"
	KindToolResult EntryKind = "tool_result"
)

// Entry is one unit of dialogue in the context, tagged with the Seq of the
// event that produced it so a Redacted event can find it again.
type Entry struct {
	Seq      int
	Side     EntrySide
	Kind     EntryKind
	Actor    Actor
	Parts    Parts
	CallID   string
	ToolName string
	IsError  bool
	Redacted bool

	// WasHint records that this message arrived mid-turn. The renderer does
	// not need it and does not read it; it exists so the GUI and the engine
	// can tell the story accurately.
	WasHint bool
}

// Ephemera is volatile data attached to exactly the next request, then gone.
// A stale timestamp resent forever is a lie, so ephemera are never history.
type Ephemera struct {
	Instruction string
	Parts       Parts
}

// Context is the state after playing the log.
type Context struct {
	Turn     TurnState
	Dialogue []Entry

	// PendingHints are messages that arrived mid-turn and have not yet been
	// carried to the model. They are NOT in the dialogue yet: they join it at
	// the moment they are delivered, which is what keeps the rendered history
	// stable from that point on.
	PendingHints []Entry

	// PendingTools are tool calls the engine is expected to execute. A call
	// recorded on an interrupted turn never lands here.
	PendingTools []string

	Ephemera  *Ephemera
	Usage     Usage
	ModelName string
	Signature string // last opaque vendor replay material for thinking
}

// NewContext is the state of a conversation that has not happened yet.
func NewContext() *Context { return &Context{Turn: Idle} }

// Apply folds one event into the context. Given the context and this event —
// nothing else — the new state is determined.
func (c *Context) Apply(ev Event) {
	switch ev.Type {

	case EvMessageReceived:
		entry := Entry{
			Seq: ev.Seq, Side: SideUser, Kind: KindText,
			Parts: ev.Parts,
		}
		if ev.Actor != nil {
			entry.Actor = *ev.Actor
		}
		switch c.Turn {
		case InFlight, ToolsPending:
			// The same event, arriving mid-turn, is a hint. There is no Hint
			// event type and there must not be one: classification is a
			// function of state, and only the reducer holds the state.
			entry.WasHint = true
			c.PendingHints = append(c.PendingHints, entry)
		case Idle:
			c.Dialogue = append(c.Dialogue, entry)
			c.Turn = InputPending
		default:
			// InputPending: input accumulates — a room may have several
			// people talking before anyone is answered.
			// Interrupted: a prompt for the next turn, once this one drains.
			c.Dialogue = append(c.Dialogue, entry)
		}

	case EvAssistantMsg:
		if ev.Parts.Text() != "" || len(ev.Parts) > 0 {
			c.Dialogue = append(c.Dialogue, Entry{
				Seq: ev.Seq, Side: SideAssistant, Kind: KindText,
				Actor: Model, Parts: ev.Parts,
			})
		}

	case EvAssistantThink:
		// The thinking TEXT stays in the log, for the GUI. Only the vendor's
		// opaque replay material enters the context.
		c.Signature = ev.Signature

	case EvToolCalled:
		c.Dialogue = append(c.Dialogue, Entry{
			Seq: ev.Seq, Side: SideAssistant, Kind: KindToolUse,
			Actor: Model, CallID: ev.CallID, ToolName: ev.ToolName,
		})
		if c.Turn != Interrupted {
			c.PendingTools = append(c.PendingTools, ev.CallID)
		}
		// else: recorded, never executed. Truth without action.

	case EvToolReturned:
		c.Dialogue = append(c.Dialogue, Entry{
			Seq: ev.Seq, Side: SideUser, Kind: KindToolResult,
			Actor: Tool, CallID: ev.CallID, Parts: ev.Parts, IsError: ev.IsError,
		})
		c.PendingTools = remove(c.PendingTools, ev.CallID)

	case EvInterrupted:
		if c.Turn == Idle {
			break // nothing to interrupt; defined, and a no-op
		}
		c.Turn = Interrupted
		c.PendingTools = nil

	case EvRequestSent:
		// The round-trip boundary. Ephemera are consumed here, and a pending
		// hint joins the dialogue at the end — exactly where the renderer put
		// it — so that every later request carries it in the same place.
		c.Dialogue = append(c.Dialogue, c.PendingHints...)
		c.PendingHints = nil
		c.Ephemera = nil
		if c.Turn != Interrupted {
			c.Turn = InFlight
		}

	case EvResponseEnded:
		if ev.Usage != nil {
			c.Usage.Input += ev.Usage.Input
			c.Usage.Output += ev.Usage.Output
		}
		if c.Turn == Interrupted {
			c.Turn = Idle // the dead turn has fully drained
			break
		}
		if ev.StopReason == "tool_use" {
			c.Turn = ToolsPending
		} else {
			c.Turn = Idle
		}

	case EvErrorOccurred:
		// Infrastructure errors change state and contribute no content. A
		// cured 429 is something the model never hears about.
		if ev.Source == "Provider" && (c.Turn == InFlight || c.Turn == ToolsPending) {
			c.Turn = InputPending
		}

	case EvEphemeraSet:
		c.Ephemera = &Ephemera{Instruction: ev.Instruction, Parts: ev.Parts}

	case EvRedacted:
		for i := range c.Dialogue {
			if c.Dialogue[i].Seq == ev.TargetSeq {
				c.Dialogue[i].Redacted = true
			}
		}

	case EvModelChanged:
		c.ModelName = ev.ModelName

	default:
		// Unreachable: LoadLog refuses unknown types before we get here.
		panic(fmt.Sprintf("reducer has no transition for event type %q", ev.Type))
	}
}

// Play folds a whole log into a fresh context. This is the definition of
// "context": the state after playing the event log.
func Play(l *Log) *Context {
	c := NewContext()
	for _, ev := range l.Events {
		c.Apply(ev)
	}
	return c
}

func remove(ss []string, s string) []string {
	out := ss[:0]
	for _, v := range ss {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
