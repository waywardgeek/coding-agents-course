# Ch2 implementation spec (condensed from chapter-02-outline.md Draft 4)

Working notes for the coder. The outline is authoritative; this is the subset
needed while building. Regenerate from the outline if they disagree.

## What ch2 is

One context. Three renderers, three parsers: Anthropic, OpenAI, Gemini.
Bidirectional seam. **No tool execution** (ch3). **No concurrency/mailbox/hints/
interrupts** (ch4). `chat` from ch1 survives unchanged.

Build order is FIXED and load-bearing: Anthropic → OpenAI → Gemini. Alien vendor
last so "the third was nearly free" can actually be falsified.

## Commands

| command | behaviour |
|---|---|
| `./ch02 chat` | ch1's interactive loop, observably unchanged |
| `./ch02 render LOG` | play LOG → context → render; print vendor request JSON to stdout, **nothing else**, exit 0, **no network** |
| `./ch02 dump` | write event log as JSON-lines |

`render` takes **NO FLAGS**. Vendor/model from env: `LLM_VENDOR=anthropic|openai|gemini`.
Reason: two renders compared byte-for-byte; flags make byte-identity a property of
invocation rather than of the log.

## Checks (sum = 100)

| check | pts | property |
|---|---|---|
| `session` | 0 | stdio protocol; directives acked; request census |
| `ch1parity` | 25 | all seven ch1 checks still pass |
| `logdump` | 5 | dump → render round-trips in a fresh process |
| `replay` | 10 | two renders of one log byte-identical |
| `redaction` | 10 | Redacted event names target; content absent from later renders |
| `ephemera` | 10 | delivered exactly once, then absent; never written to history |
| `usage` | 10 | four token categories normalized from all three vendors into one **disjoint** set summing to billable total |
| `seam-render` | 15 | one log renders correctly to all three request shapes |
| `seam-parse` | 15 | three responses → contexts identical **apart from Provenance**, which must be PRESERVED not normalized away |

**seam-parse: fully-identical contexts FAIL** (provenance thrown away).

## Type contract (§2.4a)

Grader normalizes event-type AND field names case- and punctuation-insensitively.
**Never grade on Go identifiers.**

```go
type Seq uint64

type Event struct {
    Seq  Seq
    Type EventType
    Time time.Time // METADATA ONLY. Ordering from Seq, never this.
    // exactly one non-nil, selected by Type
    Message  *MessageData
    Request  *RequestData
    Response *ResponseData
    Tool     *ToolData
    Redact   *RedactData
    Error    *ErrorData
}

type MessageData  struct { Actor Actor; Parts []Part }
type ResponseData struct { Parts []Part; Usage Usage; From Provenance }

// Recorded at WRITE time by the client that produced it. NEVER inferred.
type Provenance struct {
    Vendor  Vendor
    Model   string  // OPEN set, never switched on, equality only
    Surface Surface
}

type Vendor uint8
type Surface uint8
const ( VendorAnthropic Vendor = iota + 1; VendorGemini; VendorOpenAI )
const ( SurfaceMessages Surface = iota + 1; SurfaceInteractions; SurfaceResponses )
// iota+1: zero value INVALID so unpopulated != "Anthropic".
// Marshal as readable STRINGS (log is greppable JSON-lines).
// Unknown value on read = LOUD REFUSAL, not a default, not a skip.

type RedactData struct {
    From, To    Seq       // the SPAN superseded, inclusive
    Level       Redaction
    Replacement []Part    // RedactSummary ONLY; purges synthesize stubs
    Reason      string
}
type Redaction uint8
const ( RedactResult Redaction = iota + 1; RedactTool; RedactDialogue; RedactSummary )
// Ch2 exercises ONLY RedactResult.
```

### Content

```go
type Part interface{ isPart() }
type TextPart     struct{ Text string }
type BlobPart     struct{ MIME, Path string } // on disk, never inline
type OpaquePart   struct{ From Provenance; Data json.RawMessage }
type RedactedPart struct{ Stub string } // SYNTHESIZED by reducer, not stored
type ToolCallPart struct {
    CallID string // id AS ISSUED by the model in From
    From   Provenance
    Name   string
    Args   json.RawMessage
}
type ToolResultPart struct {
    CallID  string
    Parts   []Part
    IsError bool // a tool that ran and failed is CONTENT, not ErrorOccurred
}
```

### Context

```go
type Context struct {
    Turn     TurnState
    Dialogue []Entry
    Ephemera []Part  // pending; delivered once then cleared
    Usage    Usage   // running totals, vendor-normalized
}
type Entry struct { Actor Actor; Parts []Part }

// Counts NEVER money. All four DISJOINT, summing to billable total.
type Usage struct {
    Input      int // neither read from nor written to cache
    CacheWrite int // typically costs MORE than plain input
    CacheRead  int // typically ~an order of magnitude LESS
    Output     int
}
```

**No field in the context may grow without bound.** No `Redacted map[Seq]bool`
(grew forever AND redundant with the log). No system-prompt field — it is
renderer OUTPUT computed from Config.

### The seam

```go
type Renderer interface{ Render(*Context, Config) (*http.Request, error) }
type Parser   interface{ Parse(status int, body []byte) ([]Event, error) }
```

**No vendor types in either signature.** `Parse` returns EVENTS, not a context —
one path in: append events, run reducer.

## Events & reducer

Taxonomy (FROZEN for the exercise): `MessageReceived`, `RequestSent`,
`ResponseStarted`, `ResponseEnded`, `ToolCalled`, `ToolReturned`, `Redacted`,
`ErrorOccurred`. (ch4 adds `Interrupted`; ch3 adds job events.)

Turn states: `Idle`, `InputPending`, `InFlight`, `ToolsPending`.

| transition | result |
|---|---|
| Idle × MessageReceived | InputPending |
| InputPending × RequestSent | InFlight |
| InFlight × ResponseEnded (tool calls) | ToolsPending |
| InFlight × ResponseEnded (no tool calls) | Idle |
| ToolsPending × ToolReturned (last) | InputPending |
| InFlight × ErrorOccurred | Idle |

**Every pair not listed is identity** — the `default` arm, NOT a panic.
Single `ErrorOccurred`: infra errors change turn state; semantic tool failures
are tool CONTENT (`ToolResultPart.IsError`).

## Key rendering rules

- **Tool-call id synthesis**: rendering an Anthropic `toolu_…` id to OpenAI
  requires a synthesized id **derived from Seq**, never random — `replay`
  compares bytes. Seam meets determinism rule at this field.
- **Exhibit A** — one `ToolReturned`, Actor=Tool, three authorships:
  - Anthropic: `tool_result` block in a **user** message
  - OpenAI: own message, **tool** role, `tool_call_id`
  - Gemini: `functionResponse` part in a **user** turn
- **Exhibit B** — Anthropic refuses two consecutive user messages, so a tool
  result + the human's next instruction MERGE into one message. OpenAI/Gemini
  keep them separate.
- **Media asymmetry is a LOUD error** — audio part to a text-only model raises,
  never silently drops.
- **Opaque replay material** carried, never interpreted; handed back only to the
  EXACT model that issued it.
- Four non-determinism sources: clock, random id, Go map iteration order, set
  iteration. Wire types = structs with ordered fields, never `map[string]any`.

## Replay/versioning

Replay with CURRENT code. Log format carries a semantic version.
**Unknown event type = loud refusal to load**, not a skip.

## Grader requirements (from brief)

- Fake endpoints for ALL THREE vendors. One key or none must score 100.
- Record-then-judge; malformed requests still get 200 (all bugs in one run).
- Deterministic token counts (1 per 4 chars) to catch invented usage.
- Request census in the 0-point `session` check.
- Mutation tests asserting the EXACT SET of failing check ids.
