# Chapter 2 — One Log, Three Vendors

*The data structures, and the seam they exist to make possible.*

**Status:** Draft 4, 2026-09-12. Outline only; prose not yet written.

---

## Draft 4 — what changed and why

Draft 3 was "the real data structures," with the vendor seam present as an
*idea* and the exercise targeting Anthropic alone. That is the same mistake the
chapter now opens by describing: a seam with one implementation is not a seam,
it is a naming convention. Draft 4 makes the seam the subject, and makes the
student build three renderers and three parsers over one context.

Three structural changes:

1. **The LLM seam is the chapter.** The data structures are motivated *by* the
   seam rather than the seam being one of their benefits. Three vendors —
   Anthropic, OpenAI, Gemini — request rendering **and** response parsing.
2. **Hints and interrupts moved to Chapter 4.** They are about time and
   concurrency, not about vendors. Nothing is retracted; the material is
   preserved verbatim in `book/chapter-04-actors-parking.md`, including review
   findings M1, M2, M4, E4 and E7.
3. **The `agent_status` toy tool is gone.** It existed only so a Chapter 2 turn
   would have a middle for a hint to land in. With tools in Chapter 3 and hints
   in Chapter 4, the constraint that forced it no longer exists. Chapter 2
   *renders* logs containing tool events without executing any.

**The write-once constraint now governs the book.** Chapter 1 is the single
sacrificial chapter. From Chapter 2 on, every chapter is strictly additive —
new events, new tools, new seams, never "delete what you built." This is why
Draft 4 exists at all: under write-once, whatever Chapter 2 gets wrong is
inherited by every chapter after it.

**Claim status:**

- **FALSIFIED and removed** (Draft 3, finding M9): the claim that mid-turn
  `system` messages work on `claude-opus-4-8`+ but not `claude-sonnet-5`.
  Measured live; Sonnet 5 accepts and obeys them. That material now lives in
  the Chapter 4 parking file with the correction applied.
- **STILL UNVERIFIED:** the July 2025 priority claim for mid-turn hints — now
  a Chapter 4 problem, not this chapter's.
- **NEEDS VERIFICATION BEFORE PRINT:** every wire-format detail in §2.6. The
  three request and response shapes are quoted from working knowledge and must
  be checked against current vendor documentation by the coder. Wire formats
  drift, and this chapter is nothing but wire formats.

---

## §2.0 Cold open — the seam I got wrong, and what it cost

Open with the author's own failure, told plainly, because it is the most
expensive mistake in this book and it looks completely reasonable while you are
making it.

The sequence:

1. Build an agent against Anthropic. It works.
2. Extract an interface — `AIClientInterface` — with all the methods needed,
   as they were needed, by `ClaudeClient`.
3. Add a second vendor. The interface does not fit, because it was never
   vendor-shaped; it was Claude-shaped with an interface keyword in front of
   it. So: copy `ClaudeClient`, paste, edit until Gemini works.
4. Add a third. Copy, paste, edit until OpenAI works.
5. Discover, roughly 30,000 lines later, that three near-identical clients
   drift independently, that every bug must be fixed three times, and that
   two of the three fixes will be forgotten.

Then the part most books leave out. **The remedy was worse than the disease.**
The correct seam was eventually designed — one context, one renderer per
vendor, one parser per vendor — and delivered as a big-bang rewrite. A year
later the migration is still not finished. The product works. It is also
semi-broken in ways its author has not finished cataloguing, and some bugs have
not yet been reported to anyone, including himself.

The lesson has two halves and readers usually get taught only the first:

> **The seam was right. Shipping it as a rewrite was the mistake.**
> Cutting a seam late does not cost you one refactor. It costs you a tail —
> and the tail is paid by whoever is using the product while you migrate.

**This is why the chapter charges an hour for something a reader would rather
skip.** Writing three renderers on day one feels like over-engineering. It is
the cheapest hour in the book: it buys the shape of an interface that was
*derived from three implementations* instead of extrapolated from one.

### The demolition (carried forward from Draft 3)

Chapter 1 ended with a working agent and a deliberate attachment to it. This
section takes it apart. The `[]{role, content}` array cannot express: what the
model actually returned versus what we chose to send, tool calls and their
results, a redaction, token accounting, or who said a thing and why. Chapter 1
was "total garbage, but the student learns the basics" — and the student now
knows enough to see why.

State the promise honestly, because Draft 3 learned this the hard way when
scope changed: **the rewrite is observably identical for everything Chapter 1
could already do.** It then gains something Chapter 1 could not express at any
price — the same conversation, correctly, to three different vendors.

---

## §2.1 History ≠ Context

The distinction the whole book rests on.

- **History** is an append-only event log. What happened, in order, forever.
  It is the truth and it is never edited.
- **Context** is the vendor-independent state you get by replaying that log.
  It is derived, reconstructible, and disposable.
- **The request** is what a *renderer* makes from context for one specific
  vendor. It is disposable and it is a lie by omission — necessarily.

Three consumers, three needs: the renderer reads context; the GUI reads the
log; the auditor reads the log.

**Full re-send is the price of ownership.** Every request carries the entire
conversation. You pay for it in tokens (largely refunded by prefix caching, a
later chapter) and you buy the ability to edit history — which is the core
capability of a coding agent and the reason §2.6 declines vendor stateful
conversation APIs.

> **Sidebar: "Is this request mid-turn?" — a bug from building this chapter.**
>
> Draft 3's grader needed to know whether an incoming request began a new turn
> or continued a tool loop. The obvious test: *does it contain `tool_result`
> blocks?*
>
> Wrong, and wrong in exactly the way this section is about. Once a tool loop
> has happened, **every later request contains those tool results forever**,
> because history is re-sent in full. Four new-turn requests were classified as
> continuations of a loop that had ended three turns earlier.
>
> The general lesson: **a request is not a description of the current moment.**
> It is the entire history, re-sent, with a little new material on the end. Any
> question shaped like "what is happening right now?" must be asked of the
> *end* of the request, or of the log — never of the whole.

---

## §2.2 The event log

- Append-only. Monotonic `Seq`. **Ordering is primary**; wall-clock time is
  metadata and may be wrong, duplicated, or non-monotonic across machines.
- Never edited, never reordered, never deleted in place. A redaction is a new
  event that supersedes, not a mutation of an old one (§2.6).
- Serialized as JSON-lines so it is greppable with ordinary tools — a property
  that becomes load-bearing in Chapter 3, when tool output starts arriving by
  the megabyte.

---

## §2.3 Events

### The self-contained event rule (load-bearing)

Given the current context and just the next event — nothing else — we can
correctly compute the new context. Write the rule as:

```
newContext = Apply(context, event)
```

That notation describes **information flow**. It is the chapter's central
claim: everything needed to advance the context is in the context plus one
event. It is *not* a demand for value semantics. In Go the natural
implementation is a pointer receiver mutating in place, and that is the one to
write — a `Context` full of slices copied by value gives you two contexts
sharing one backing array, and that bug is indistinguishable from renderer
non-determinism (§2.7).

**Classification is the reducer's job, not the capture site's.** The same
arriving bytes mean different things depending on turn state. Decide at capture
time and you are wrong every time the human types quickly. Only the reducer
holds the state that makes the decision correct. (Chapter 4 makes this vivid:
the *same* event is a prompt or a hint depending solely on turn state.)

### The taxonomy for this chapter

`MessageReceived`, `RequestSent`, `ResponseStarted`, `ResponseEnded`,
`ToolCalled`, `ToolReturned`, `Redacted`, `ErrorOccurred`.

Chapter 4 adds `Interrupted`. Chapter 3 adds job events. **Additive, always** —
this is the first place the write-once discipline is visible to the reader.

Two notes:

- **A single `ErrorOccurred`.** Infrastructure errors change turn state;
  semantic errors (a tool that ran and failed) are ordinary tool *content*.
  Conflating them is why agents get stuck retrying a compile error as though it
  were a network outage.
- **Thinking text is log-only.** The context carries opaque replay material —
  a signature, a redacted block, an id — tagged with the exact model that
  produced it, and the renderer decides whether that model wants it back. Never
  reconstruct reasoning as prose and feed it to a different model as though it
  were your own.

### Tool events without a tool loop

Chapter 2 executes no tools. It **renders logs that contain tool events**,
supplied by the exercise. The student therefore writes a reducer that handles
events it cannot yet produce.

That is not an accident of sequencing; it is the write-once discipline in
miniature, and it is worth saying so. You are building the shape before the
capability, because the shape is what determines whether the capability can be
added without a rewrite.

### Turn states

`Idle`, `InputPending`, `InFlight`, `ToolsPending`. (`Interrupted` arrives in
Chapter 4 — and it must be a *state*, not a flag, or replay re-executes tool
calls that were cancelled.)

| transition | result | note |
|---|---|---|
| Idle × MessageReceived | InputPending | ordinary prompt |
| InputPending × RequestSent | InFlight | |
| InFlight × ResponseEnded (tool calls) | ToolsPending | |
| InFlight × ResponseEnded (no tool calls) | Idle | turn complete |
| ToolsPending × ToolReturned (last) | InputPending | loop continues |
| InFlight × ErrorOccurred | Idle | infrastructure failure ends the turn |

**Every pair not listed is identity.** That sentence, not the length of the
table, is what makes the reducer total. A table enumerates the transitions we
thought of; the default covers the ones we did not. Write it as the `default`
arm of the switch, **not** as a `panic`.

---

## §2.4 The context

The context is **vendor-independent by construction**, and this chapter is the
only one that can prove it.

Contents: the dialogue (ordered, actor-attributed, parts-structured), pending
ephemera, redaction state, token accounting, and opaque per-model replay
material carried but never interpreted.

**Content is Parts, not a string.** Text, tool calls, tool results, images,
audio, and vendor-opaque blobs. A string is the Chapter 1 mistake wearing a
struct.

### The system prompt is rendered, not stored

*The inoculation. It teaches no prompt content and prevents the most common
architectural mess in the field.*

The system prompt is **output of the renderer**, computed from context plus
configuration. It is not a blob of text living in the log, and it is not a
field on the context that someone appends to.

For now, a constant string is a perfectly good renderer. The rule is only about
**where it comes from**.

Why this matters enough to state before we need it: the system prompt is the
most abused surface in agent engineering, and the abuse has a predictable
shape. First someone describes the tools in it by hand. Then the descriptions
drift from the actual tools. Then part of it is generated and part is
hand-written, and no one can say which. By the time it is 400 lines nobody will
delete a word, because nobody can prove which words are load-bearing.

Chapter 6 replaces the constant with generation from skills. Under write-once
that must be a pure addition — and it is, **provided the system prompt was
never a stored value in the first place.**

The three vendors make the point concrete before the reader can form a bad
habit: Anthropic takes a top-level `system` parameter, OpenAI takes a `system`
(or `developer`) message inside the array, Gemini takes a separate
`systemInstruction` object. One fact; three placements; a renderer's problem.
Store it and you have just picked a vendor.

---

## §2.4a The types, in one place

*Gathered here so they can be reviewed as a set. In the prose each type is
introduced where it is motivated — this section is the contract with the
grader and the reference solution.*

Shapes, not implementations. Field names are illustrative; the grader
normalizes names case- and punctuation-insensitively, and must never grade on
Go identifiers.

### The log

```go
type Seq uint64

type Event struct {
    Seq  Seq
    Type EventType
    Time time.Time // METADATA. Ordering comes from Seq, never from this.

    // Exactly one is non-nil, selected by Type. Verbose on purpose:
    // it round-trips as JSON with no registry, and it makes the
    // reducer's switch exhaustive by construction.
    Message  *MessageData
    Request  *RequestData
    Response *ResponseData
    Tool     *ToolData
    Redact   *RedactData
    Error    *ErrorData
}

type MessageData struct {
    Actor Actor
    Parts []Part
}

type ResponseData struct {
    Parts []Part
    Usage Usage
    From  Provenance // NOT a vendor string. See below — this is load-bearing.
}

// Provenance records who produced content. Recorded at WRITE time by the
// client that produced it, NEVER inferred later.
type Provenance struct {
    Vendor  string // "anthropic" | "google" | "openai"
    Model   string // "claude-opus-5" — the exact model, not the vendor
    Surface string // "messages" | "interactions" | "responses"
}

type RedactData struct {
    Target Seq    // the event superseded. Never mutate the original.
    Reason string
}
```

### Content

```go
type Part interface{ isPart() }

type TextPart   struct{ Text string }
type BlobPart   struct{ MIME, Path string } // on disk, never inline
type OpaquePart struct{ From Provenance; Data json.RawMessage }

type ToolCallPart struct {
    CallID string // the ID AS ISSUED, by the model named in From
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

**`Provenance` is the subtle one, and it is where a seam that looks finished
turns out not to be.**

The naive version of this field is `Vendor string`. That is wrong, and it is
wrong in a way you will not discover until a user switches models mid
conversation.

**Thinking signatures — the encrypted reasoning material — are bound to the
model, not the vendor.** Replay a signature produced by one model while talking
to another and the two vendors we have both fail, differently:

| vendor | replaying another model's thinking |
|---|---|
| Google | returns an **error** |
| Anthropic | **silently drops it** |

Stop on that table, because it is a rule of this book in the wild: **the loud
failure is the good one.** Google's error costs you an afternoon. Anthropic's
silent drop costs you a subtly worse agent that still passes every test —
reasoning quietly discarded, nothing in the logs, no way to tell from the
outside. Exactly the shape §2.6 forbids when it insists that media asymmetry
must raise rather than drop.

And `Surface` earns its place for the same reason. Signature validity is scoped
to *(vendor, model, surface)* — not to vendor:

- Gemini **Interactions** attaches signatures to thought steps and built-in
  tool steps, but never to standard function calls.
- Gemini legacy **generateContent** returns a 400 if you replay a
  `functionCall` *without* its signature.

So a coarse vendor tag cannot even decide whether to *include* the material,
let alone whether it is replayable. **Without the exact model, a Gemini
renderer cannot construct a valid request at all.**

Hence the rule, which is about capture rather than rendering:

> Provenance is recorded at **write time**, by the client that produced the
> content, and is never inferred afterwards. A renderer may read it. Nothing
> may reconstruct it.

Inference is impossible in principle here — by the time you are rendering, the
model that produced a signature three turns ago is simply not derivable from
anything else in the context. Miss it at capture and the information is gone.

### The tool-call id, and where it collides with replay

The same field carries a second load. A tool-call id is the one piece of vendor
vocabulary that legitimately enters the context — you cannot answer a call
without quoting the id that made it.

Render a conversation holding an Anthropic `toolu_…` id to OpenAI and it is
meaningless, so the renderer must synthesize one. It must be **derived from
`Seq`**, not generated randomly, because `replay` compares bytes and a random
id is one of the four non-determinism sources named in §2.7. The seam and the
determinism rule meet at this field, and students who wire them up
independently will collide.

### The context

```go
type Context struct {
    Turn     TurnState
    Dialogue []Entry
    Ephemera []Part        // pending; delivered once, then cleared
    Redacted map[Seq]bool  // supersessions already applied
    Usage    Usage         // running totals, vendor-normalized
}

type Entry struct {
    Actor Actor
    Parts []Part
}
```

**Note what is absent: there is nowhere to put system prompt text.** That is
deliberate and structural. The system prompt is an *output*, computed by the
renderer from `Config` (§2.4). If a reader wants to store it, they will have to
add a field — and adding it is the moment to stop and re-read §2.4.

Equally absent: `role`, `content`, `tool_use_id`, `assistant`.

**But note the distinction `Provenance` forces, because it looks like a
violation and is not.** Storing `"anthropic"` or `"claude-opus-5"` is recording
a *fact about where bytes came from*. Storing `role` or `tool_use_id` would be
adopting a vendor's *description of what the bytes are*. The first is history —
it happened, it is not re-derivable, and throwing it away is lossy. The second
is a format decision, and format decisions belong in the renderer.

The test to apply to any field you are tempted to add: **could this have been
different if the same conversation had happened against another vendor?** If
yes, it is provenance and it belongs. If it is just that vendor's word for
something you already model, it has leaked.

### The seam

```go
type Renderer interface {
    Render(*Context, Config) (*http.Request, error)
}

type Parser interface {
    Parse(status int, body []byte) ([]Event, error)
}
```

**Two methods. No vendor types in either signature.** That is the whole
remedy, and its smallness is the point.

`Parse` returns **events**, not a message or a context. There is exactly one
path into the context — append events, run the reducer — so a vendor response
and a human keystroke enter by the same door. Give the parser the power to
mutate context directly and you have quietly created a second reducer that
nobody will remember to keep total.

Contrast the shape that cost 30,000 lines:

```go
// DON'T. This is the mistake in §2.0, in its natural habitat.
type AIClientInterface interface {
    SendMessage(msgs []ClaudeMessage) (*ClaudeResponse, error)
    CountTokens(msgs []ClaudeMessage) (int, error)
    // ...eleven more methods, each shaped by what ClaudeClient
    //    already happened to do
}
```

The tell is visible without knowing the story: **vendor types in the
signature.** `ClaudeMessage` in the interface means the interface *is* the
Claude client, and the second implementation can only be a copy-paste. An
interface extracted from one implementation records that implementation's
accidents as though they were requirements.

### What exists now for a later chapter

Under write-once, readers will reasonably ask why a field exists that nothing
uses yet. Answer honestly, in the text:

| element | used in Ch2 | exists for |
|---|---|---|
| `ToolCallPart` / `ToolResultPart` | rendering supplied logs | Ch3, where tools are executed |
| `BlobPart.Path` | not exercised | Ch3, when tool output arrives by the megabyte |
| `OpaquePart` | thinking signatures | every chapter; never interpreted |
| `TurnState.ToolsPending` | replaying supplied logs | Ch3 |
| `Redacted` | fully exercised | Ch7 |

This table is the write-once discipline made visible, and it is the chapter's
answer to "am I over-engineering?" — no, and here is the receipt.

---

## §2.5 Actors and rooms

Actors: `Human`, `Agent`, `System`, `Tool`. Rooms group a conversation. There
is deliberately **no `To` field** — addressing is a property of the room, not
of the message, and adding `To` invites a routing layer the book does not want.

The `Tool` actor looks like over-modelling until §2.6, where it becomes the
single sharpest demonstration in the chapter. Hold that thought.

---

## §2.6 The seam — renderers and parsers

**The centerpiece.** Everything before this exists to make this section
possible.

> The context is the truth. A **renderer** turns truth into one vendor's
> request. A **parser** turns one vendor's response back into truth.
> Distortion lives in those two places and nowhere else.

The seam is **bidirectional**, and rendering is the easy half. `AIClientInterface`
did not fail because request formatting was hard; it failed because
vendor-shaped thinking hid in the response path, in retries, in errors, in
streaming, in token accounting.

### Exhibit A — one tool result, three authorships

The demonstration the chapter is built around. A single `ToolReturned` event,
`Actor: Tool`, rendered three ways.

**Anthropic** — a `tool_result` block inside a **user** message:

```json
{ "role": "user",
  "content": [ { "type": "tool_result", "tool_use_id": "toolu_…",
                 "content": "ok" } ] }
```

**OpenAI** — its own message with a **tool** role:

```json
{ "role": "tool", "tool_call_id": "call_…", "content": "ok" }
```

**Gemini** — a `functionResponse` part in a **user** turn:

```json
{ "role": "user",
  "parts": [ { "functionResponse": { "name": "…", "response": { … } } } ] }
```

Three vendors cannot agree on who said it. Anthropic says the human did — which
is false, and is the tidiest available lie under a format that demands strict
user/assistant alternation. OpenAI invents a role. Gemini splits the
difference.

**The context is right and all three wire formats are compromises, in different
directions.** That is the entire argument for the seam, and it is not an
analogy — the student will watch one `Actor: Tool` event become three different
claims about authorship, and none of the three is worth storing.

Corollary, and the reason §2.5 models a `Tool` actor at all: **authorship is a
rendering decision.** If your context stores `role: "user"` for a tool result
because that is what Anthropic wanted, you have already lost, and you will
discover it in the copy-paste.

### Exhibit B — the merged message

Anthropic will not accept two consecutive user messages. So a tool result plus
the human's next instruction merge into one message:

```json
{ "role": "user",
  "content": [
    { "type": "tool_result", "tool_use_id": "toolu_…", "content": "ok" },
    { "type": "text", "text": "now check the config instead" }
  ] }
```

Nothing in the context looks like this. Two honest, separate, ordered facts,
fused because one vendor demands alternation. Render the same log for OpenAI
and they stay separate. The merge is not a fact about the conversation; it is a
fact about a wire format, and it belongs in exactly one function.

### Exhibit C — parsing back

Three response shapes normalize to identical context:

| vendor | assistant text at | tool calls at | stop signal |
|---|---|---|---|
| Anthropic | `content[]` blocks | `tool_use` blocks | `stop_reason` |
| OpenAI | `choices[0].message.content` | `.tool_calls[]` | `finish_reason` |
| Gemini | `candidates[0].content.parts[]` | `functionCall` parts | `finishReason` |

The grader's real question: **feed all three responses, get contexts that are
byte-identical apart from `Provenance`.** Everything the model *said* must
normalize; exactly one thing must survive — the record of who said it, and with
which model, on which surface.

That "apart from" is not a loophole, it is the whole distinction of §2.4a: the
content normalizes because content is ours, the provenance persists because it
is history and is not re-derivable. A submission whose three contexts are
*fully* identical has thrown provenance away and will be unable to render a
valid Gemini request later. A submission whose contexts differ anywhere else
has leaked vendor shape past the parser — and leaked vendor shape is precisely
what makes the second implementation a copy-paste.

Also normalized here: token accounting (`usage.input_tokens` /
`prompt_tokens` / `usageMetadata.promptTokenCount`), and errors — an HTTP 429
is an `ErrorOccurred`, not a response.

### Rules the seam has to hold

- **Decline vendor stateful conversation APIs.** Server-side threads (or
  `previous_response_id`-style continuations) trade away the ability to edit
  history. Editing history is a coding agent's core tool: redaction, replay,
  context surgery. Own the history or you cannot build the product.
- **Media asymmetry is a LOUD error.** An audio part rendered for a text-only
  model raises; it never silently drops. Fallbacks convert an invariant
  violation into silently-wrong output.
- **Opaque replay material is carried, never interpreted.** Thinking
  signatures, tool-use ids, cache markers: store them, hand them back to the
  **exact model** that issued them, and never to a different one. Vendor is not
  a fine enough grain — see `Provenance` in §2.4a.
- **The context never learns a vendor's vocabulary.** If the word `assistant`,
  `toolu_`, or `functionCall` appears in your context types, the seam has
  already leaked.

### The prediction the chapter makes out loud

The second renderer costs real work. **The third should be nearly free.** If it
is not, the seam is wrong — and the reader will discover that in an hour
instead of in 30,000 lines.

This is the chapter's falsifiable claim about its own design, and students
should be told to notice whether it holds for them.

---

## §2.7 Versioning and replay

- Replay with **current code**, not with historical code. Log format carries a
  semantic version.
- **An unknown event type is a refusal to load, loudly.** Not a skip. Skipping
  an unknown event silently produces a context that is wrong in a way nothing
  downstream can detect.
- Retention is **policy**, not architecture: the log is complete; what you keep
  is a separate decision.

**Four ways non-determinism gets into a renderer**, named because the failure
message only tells you *that* two renders differed: the clock, a randomly
generated id, Go's deliberately randomized map iteration order, and iteration
over a set. The last two are the same bug in different hats, and they are why
wire types should be structs with ordered fields rather than `map[string]any` —
the map serializes differently on some future run, on some future machine, and
never on the one where you tested.

---

## §2.8 The exercise

### Commands

| command | behaviour |
|---|---|
| `./ch02 chat` | Chapter 1's interactive loop, unchanged in observable behaviour |
| `./ch02 render LOG` | play `LOG` → context → render; print the vendor request JSON that *would* be sent to stdout, and **nothing else on stdout**; exit `0`. **Makes no network call.** |
| `./ch02 dump` | write the event log as JSON-lines |

`render` is the centerpiece: it makes replay, redaction, ephemera and the seam
into **byte comparisons**, and it proves context is separable from transport.
If your architecture cannot offer it cheaply, your context is not actually
separate from your transport — which is the finding the exercise exists to
surface.

`render` takes **no flags**. Vendor target, model id and every other request
parameter come from the environment (`LLM_VENDOR=anthropic|openai|gemini`),
exactly as in grader mode. Two renders are compared byte for byte, so the
moment rendering accepts `--model`, byte-identity becomes a property of how you
invoked the command rather than of the log.

### Log serialization

JSON-lines, one event per line, ascending `Seq`. Each line carries at minimum
`seq`, `type`, and the event's own fields; dialogue events carry `actor`. The
format must round-trip: `dump` → `render` must work in a fresh process with no
other state.

**The event-type vocabulary is frozen for this exercise.** Three checks assert
on the *contents* of your dumped log, which is impossible unless we agree on
names. Use the §2.3 set. Spell them as you like: the grader compares type names
lowercased with punctuation stripped, so `ToolCalled`, `tool_called` and
`TOOL-CALLED` are the same event, and it does the same for field names. What it
cannot do is guess that you called it `Halted`.

### No network calls for two of the three vendors

The grader serves **fake endpoints for all three vendors**, so a full seam can
be built and graded with a single API key — or none. Students with one vendor
account are not second-class citizens, and nobody pays three subscriptions to
finish Chapter 2.

### Checks

| check | pts | property |
|---|---|---|
| `session` | 0 | stdio protocol honoured; directives acknowledged; request census |
| `ch1parity` | 25 | all seven Chapter 1 checks still pass, unchanged |
| `logdump` | 5 | log round-trips: `dump` → `render` in a fresh process |
| `replay` | 10 | two renders of one log are byte-identical |
| `redaction` | 10 | a `Redacted` event names its target; content absent from later renders |
| `ephemera` | 10 | delivered exactly once, then absent — and never written to history |
| `usage` | 5 | token accounting normalized from all three vendors |
| `seam-render` | 15 | one log renders correctly to all three vendor request shapes |
| `seam-parse` | 20 | three vendor responses produce contexts identical apart from `Provenance` — which must be preserved, not normalized away |

**Sum: 100.**

Notes on the weighting:

- **`seam-parse` outscores `seam-render`** because parsing is the half where
  vendor shape actually hides, and the half the author got wrong.
- **`ch1parity` stays at 25**, honouring the standing guard from Chapter 1's
  review. Below that, a rewrite that silently breaks Chapter 1's contract
  starts to look survivable.
- **`session` is worth zero and can still sink a submission.** Without it, one
  unacknowledged directive fails four checks at once and the student gets four
  mysteries instead of one cause.

### What you are not building

No tool loop — Chapter 3. No mailbox, hints, or interrupts — Chapter 4. No
streaming, no retries, no skills, no sub-agents. **You are building one context
and three ways in and out of it.**

---

## §2.9 Open questions for Bill

1. **Three vendors, or two required plus one as payoff?** Three matches the war
   story exactly (the disaster was three clients) and makes the "third is
   nearly free" prediction testable. Two is a smaller exercise. Leaning three.
2. **Which two, if two?** Anthropic + Gemini are the most structurally
   different (`systemInstruction` hoisted, `role: "model"`, parts not blocks),
   so they prove more. Anthropic + OpenAI gives the sharper authorship lesson
   via the `tool` role. Exhibit A needs all three to land fully.
3. **Does `chat` have to work against all three vendors live**, or is live
   Anthropic plus faked others acceptable? Leaning the latter — cost and key
   availability are real student barriers and the seam is fully provable
   against fakes.
4. **Is `ch1parity` at 25 still right** when the seam is worth 35? It is a
   quarter of the grade for "you didn't break what you had." Defensible, but
   worth a ruling now rather than after the grader is built.
5. **Should Chapter 2 state the three-chapter roadmap** (tools → actors →
   seam-for-capabilities) so the reader knows hints are coming and does not
   design for them prematurely? Leaning yes, briefly — write-once means readers
   will reasonably ask "should I leave room for X?", and the honest answer is
   "no, we sequenced it so you don't have to."

---

## Appendix — material relocated from Draft 3

`book/chapter-04-actors-parking.md` holds, verbatim, the hint and interrupt
sections and review findings M1, M2, M4, E4, E7. Of particular value when
Chapter 4 is outlined:

- **M1**: a pending hint is not in the dialogue; `RequestSent` is what moves it
  there. The one place a careful student still fails.
- **E4**: render *before* recording `RequestSent`, or the hint is delivered one
  round late and gets misdiagnosed as vendor lag.
- **M9's replacement**: hint responsiveness is an *engine* property, not a
  vendor property — measured at an 11.4s mailbox blackout with a slow tool,
  identical across carriages.
