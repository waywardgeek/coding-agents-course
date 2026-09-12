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

**The escape hatch, stated plainly to the reader.** Write-once binds *us*, not
you. Every chapter's reference solution is public from day one — read it
whenever you like, and **start any chapter from ours instead of your own.**
Failing a check in Chapter 2 must never end your course in Chapter 4. This is
not a grudging concession — it is what makes write-once safe to promise.
Additive means the book never demolishes code you wrote; rebasing means one bad
structural choice never strands you. The graders cooperate: they run your
binary and read what it emits, never your source and never its history. No
check asks whose code it is.

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

### The contract, stated once

This is the last time you will be asked to throw code away.

Chapter 1 was sacrificial on purpose: you had to feel a wrong data structure
fail before a right one could mean anything. **From here on, every chapter is
additive.** New events, new tools, new seams — but nothing in this chapter gets
deleted in the next one, or the one after that.

That is a promise with a practical consequence, and it is the reason to state
it rather than let the reader infer it: **build the simplest thing that
satisfies this chapter.** Do not leave room for the tool loop. Do not
generalize for concurrency. Do not invent a plugin system for skills. Those are
all coming, and the chapters are sequenced so that each arrives *before* the
weight that would have made it painful — which is precisely the lesson §2.0
just paid 30,000 lines to learn.

A reader who does not trust this promise will over-engineer defensively, and
defensive over-engineering is the failure mode this book argues against
everywhere else. So: trust it. If a later chapter makes you delete something
from this one, that is our bug, not yours.

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
calls that were canceled.)

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
ephemera, token accounting, and opaque per-model replay
material carried but never interpreted. Redacted content is not tracked
separately — it is replaced in place by the reducer (§2.4a).

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
    Vendor  Vendor
    Model   string // OPEN set. New models ship weekly. Never switch on it.
    Surface Surface
}

type Vendor uint8
type Surface uint8

// iota + 1 on purpose: the zero value is INVALID, so a Provenance that was
// never populated is detectable instead of silently meaning "Anthropic".
const (
    VendorAnthropic Vendor = iota + 1
    VendorGoogle
    VendorOpenAI
)

const (
    SurfaceMessages Surface = iota + 1
    SurfaceInteractions
    SurfaceResponses
)

type RedactData struct {
    From, To    Seq        // the SPAN superseded, inclusive. Not one event.
    Level       Redaction  // what, within that span, is removed
    Replacement []Part     // RedactSummary only; purges synthesize their stubs
    Reason      string
}

// Redaction levels, weakest first. Ch2 exercises only RedactResult; the rest
// arrive with context engineering. Span says WHERE, level says WHAT.
type Redaction uint8
const (
    RedactResult   Redaction = iota + 1 // tool result content -> stub; the call survives
    RedactTool                          // call and result both go; visible reasoning survives
    RedactDialogue                      // prose and reasoning go; survivors are defined by the compaction policy
    RedactSummary                       // the span is replaced by compressed prose
)
```

### Content

```go
type Part interface{ isPart() }

type TextPart   struct{ Text string }
type BlobPart   struct{ MIME, Path string } // on disk, never inline
type OpaquePart struct{ From Provenance; Data json.RawMessage }

// The RESULT of applying a Redacted event. Replaces the parts it supersedes;
// nothing records "a redaction happened" separately — the log already does.
// The stub is informative on purpose: what it was, and how to get it back.
type RedactedPart struct{ Stub string }

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

### Enum or string? The rule, and why `Model` is the odd one out

Two of `Provenance`'s three fields are enums and one is a string, which looks
inconsistent until you have the rule:

> **Enum when the code must exhaustively handle every case. String when the
> value is only compared for equality and the set is open.**

`Vendor` and `Surface` are closed sets the renderer *switches on* — there is
exactly one renderer and parser per vendor, compiled in. A typo like
`"Messages"` in a string field is a runtime surprise; as an enum it does not
compile. `Model` is the opposite: an open set that gains members weekly, never
switched on, only ever compared — *is this the same model that issued that
signature?* Make it an enum and you need a rebuild to record a model you have
no other opinion about.

Two details that are easy to get wrong:

**Start the constants at `iota + 1`.** The zero value must be invalid.
Otherwise a `Provenance` that nobody populated is indistinguishable from a
genuine Anthropic/messages one — and since provenance must be captured at write
time and can never be reconstructed, "nobody populated it" is precisely the bug
you need to be loud. A zero value that silently means something is a default
wearing a disguise.

**Marshal them as readable strings, and refuse unknown ones on the way in.**
The log is JSON-lines so that ordinary tools can grep it (§2.2); `"vendor":2`
destroys that for no gain. An unrecognized surface on read is a **loud
refusal**, exactly as §2.7 requires for an unknown event type. Not a default,
not a skip — the same discipline, for the same reason: a value you silently
coerce is a value you will debug in production.

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
    Ephemera []Part // pending; delivered once, then cleared
    Usage    Usage  // running totals, vendor-normalized
}

// Token counts, NEVER money. All four fields are DISJOINT — they sum to the
// billable total. Vendors disagree about whether that is true of their own
// reporting; making it true is the parser's job.
type Usage struct {
    Input      int // input that was neither read from nor written to cache
    CacheWrite int // cache creation. Typically costs MORE than plain input.
    CacheRead  int // typically costs an order of magnitude LESS.
    Output     int
}

type Entry struct {
    Actor Actor
    Parts []Part
}
```

**No field in the context may grow without bound.** State it as a rule, because
it is cheap to honour now and very expensive to retrofit.

The context is not a request buffer. It is the current state of an actor that
may run for **years** — memory, identity, recent conversation, everything the
model knows about itself. Anything in it that only ever accumulates is a slow
leak with a long fuse, and the fuse burns in production, on the agent you care
most about, long after the design decision is unrecoverable.

An earlier draft of this chapter had `Redacted map[Seq]bool` in the context, to
remember which events had been superseded. It is a natural thing to write and
it is wrong twice over:

- **It grows forever.** One entry per redaction, retained for the life of the
  actor, and nothing ever removes them.
- **It is redundant.** The log already records every `Redacted` event,
  permanently. The context does not need to remember that a redaction
  *happened*; it needs to hold the content that redaction *produced*.

Deleting it removes a field and a failure mode at the same time, which is
usually the sign of a correct simplification. A redaction is not metadata about
content — **it is content**, and the context holds the result of replaying the
log, exactly as §2.1 promised.

Apply the same lens to every field and one survivor stands out: `Dialogue`
grows too. It is bounded by a *policy* — compaction and retention, a later
chapter — rather than by its shape, and the shape survives compaction unchanged
because compaction replaces entries with a summary entry. That is the
distinction to hold: `Dialogue` grows and has a plan; the map grew and had
none.

### Redaction is a family, not a flag

The shape above — a **span**, a **level**, and an optional replacement — looks
like over-modelling for a chapter that only ever stubs a tool result. It is
here because the alternative is demolishing it later, and because the thing it
grows into is the mechanism that keeps an agent alive past its context window.

The naive design is `Target Seq` plus a boolean: this event was redacted. It
cannot express "remove every tool call and result older than the last
`save_memory`," which is the single most valuable compaction there is.

**Compaction by position versus compaction by category.** The common framework
approach — Google's ADK does this — is to replace the oldest *portion* of
history with an LLM-written summary. That is compaction by **position**: it
discards whatever happens to be old, valuable or not, and what it loses is
unpredictable, because a summary is lossy in ways nobody enumerated.

Compaction by **category** discards a *kind* of content wherever it appears.
And the categories are wildly unequal: in real coding sessions, tool results
and tool-call arguments together are the clear majority of a conversation's
tokens, while carrying almost none of its continuity. The agent's reasoning,
its decisions, its sense of what it is doing — those are cheap and they are the
part you cannot regenerate.

Which yields the rule:

> **Know what you are throwing away.** Purge categories first, summarize last.
> A category purge is lossy in a way you can name and have measured. A summary
> is lossy in a way you will discover later, in production, as a personality
> change.

Hence the levels, weakest first: stub the tool *results* but keep the calls;
remove calls and results entirely but keep visible reasoning; remove prose and
reasoning but **never** the goal stack; and only when compacted records have
themselves piled up, summarize.

**Stubs are synthesized, not stored.** A `RedactResult` stub is computed by the
reducer from the event it supersedes — the tool's name, the size, the path the
output still lives at — which makes it deterministic (so replay is stable),
recoverable (§2.2's greppable log, and Chapter 3's on-disk tool output), and
free of storage that grows. Only `RedactSummary` stores a `Replacement`,
because only there is the new content something an LLM wrote and nobody can
recompute.

**And compaction is an event.** It goes in the log like everything else, which
is what lets both of this chapter's promises hold at once: the log stays
complete and append-only, replay reproduces the *compacted* context exactly,
and the context stays bounded. A framework that compacts by mutating its
in-memory history has quietly given up on replay, and will not notice until it
needs to debug a session it can no longer reconstruct.

The policy — what thresholds trigger which level, and where the boundaries fall
— is context engineering, and it gets its own chapter. Chapter 2 owes it only a
shape it will not have to break.

### Usage is four numbers, not two — and they are not disjoint the same way

`Usage` looks like bookkeeping and is not. It is the **instrument** that makes
everything in the previous section tunable: you cannot set a token threshold
you cannot measure, and you cannot justify a stable prefix without knowing what
a cache read costs relative to a cache write.

Four categories, at genuinely different prices:

| category | rough price relative to plain input |
|---|---|
| `Input` | 1× |
| `CacheWrite` | **more** than 1× — you pay a premium to create the entry |
| `CacheRead` | **far less** — roughly an order of magnitude cheaper |
| `Output` | several× |

That `CacheRead` row is the whole economic argument for the volatility ordering
in the context-engineering chapter. Put your most-changing content at the front
of the prefix and you convert the cheapest category into the most expensive one,
on every single request, forever. It is the most costly one-line mistake in
agent engineering, and it is invisible without this struct.

**The trap: vendors disagree about whether their own categories overlap.**

Some report cached tokens as a **subset** of the prompt total. Others report
them as **disjoint** additions alongside it. Normalize naively — sum everything
you are given — and you double-count on one vendor and undercount on another,
producing a cost figure that is confidently wrong in opposite directions
depending on which model you are talking to.

This is the purest seam bug in the chapter. Nothing crashes. No test fails. The
number is simply not the number, and you will act on it for months.

> **Canonical form: the four fields are disjoint and sum to the billable
> total.** Whatever a vendor reports, the parser converts it. Where a vendor's
> convention is a subset, subtract; where it is already disjoint, pass through.

*(For the coder: the per-vendor conventions must be verified against current
documentation before print — which is subset, which is disjoint, and whether
either has changed. This is exactly the class of claim that was falsified in
Draft 3. State the date of verification in the text.)*

**Record counts, never money.** Prices change; token counts are history. A
dollar amount in the log is wrong the moment a vendor reprices, and it destroys
your ability to re-cost historical sessions under new rates. Pricing is
configuration and belongs beside the model id; usage is a fact and belongs in
the log. Same distinction as `Provenance` versus format vocabulary: record what
happened, compute what it means.

**A known gap, flagged rather than solved.** At least one vendor bills cache
*storage* by duration — a real cost with no token count attached. A struct of
pure counts cannot express it, and cache lifetime is a concept Chapter 2 does
not have. Noted here so that the later caching chapter adds it deliberately,
rather than discovering that `Usage` was the wrong shape all along.

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
| `RedactedPart` | fully exercised | Ch7 |
| `Redaction` levels above `RedactResult` | not exercised | context engineering — the compaction gradient |
| `RedactData.From`/`To` as a span | spans of one | context engineering — "every tool call before the last save_memory" |

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

This is the chapter's falsifiable claim about its own design, and it is the
only place in the book where the reader can run the experiment themselves. So
the **order is fixed, and it is fixed to make the test honest**:

1. **Anthropic** — the baseline. Everything the student already has.
2. **OpenAI** — a moderate difference: a `tool` role of its own, a flat message
   list, `tool_calls` as an array. Enough divergence to force a real
   abstraction rather than a rename.
3. **Gemini** — the genuinely alien one. `contents` rather than `messages`,
   `parts` rather than blocks, `role: "model"`, `systemInstruction` hoisted
   clean out of the message list, `functionCall`/`functionResponse`.

**The hardest vendor goes last on purpose.** The tempting order puts the most
different one second and the most familiar one third — and then "the third was
nearly free" is true because the third was *easy*, not because the seam was
right. The claim would pass for the wrong reason. Put the alien one last and
the prediction is tested in the direction that can actually falsify it.

Tell the reader to notice what each one costs them. If the third takes twenty
minutes after the second took two hours, they have felt the thing this chapter
is about in a way no paragraph delivers. And if it does *not* — if Gemini
forces them back into the context to add a field — then their seam is wrong,
they have learned it on day one, and the chapter has done its job by losing its
own bet.

Two points always fit a line. A student can shape the interface around vendor
A, bend vendor B to fit it, and call the result a seam. The third
implementation is what separates an abstraction from a bridge between two
specific things.

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

| command | behavior |
|---|---|
| `./ch02 chat` | Chapter 1's interactive loop, unchanged in observable behavior |
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

**Build them in this order: Anthropic, then OpenAI, then Gemini.** The reason is
in §2.6 — the order is what makes the chapter's prediction a real test rather
than a flattering one. Note how long each takes you. That number is the
chapter's actual lesson, and it is yours rather than ours.

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
| `session` | 0 | stdio protocol honored; directives acknowledged; request census |
| `ch1parity` | 25 | all seven Chapter 1 checks still pass, unchanged |
| `logdump` | 5 | log round-trips: `dump` → `render` in a fresh process |
| `replay` | 10 | two renders of one log are byte-identical |
| `redaction` | 10 | a `Redacted` event names its target; content absent from later renders |
| `ephemera` | 10 | delivered exactly once, then absent — and never written to history |
| `usage` | 10 | all four token categories normalized from all three vendors into one **disjoint** set — cache reads and writes separated from plain input, summing to the billable total |
| `seam-render` | 15 | one log renders correctly to all three vendor request shapes |
| `seam-parse` | 15 | three vendor responses produce contexts identical apart from `Provenance` — which must be preserved, not normalized away |

**Sum: 100.**

Notes on the weighting:

- **The parse side still outweighs the render side, 25 to 15** — it is just
  itemized now. `usage` is parsing work: normalizing four token categories
  across three vendors that disagree about whether their own categories
  overlap. Pulling it out of `seam-parse` and naming it separately means a
  student who gets the message shapes right but the accounting wrong is told
  *which* half failed, instead of losing a large undifferentiated block.
  Parsing is where vendor shape hides, and where the author's own seam failed.
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

1. ~~**Three vendors, or two required plus one as payoff?**~~ **RULED
   (2026-09-12): three.** Anthropic, OpenAI, Gemini — all three graded, in that
   order. Three matches the war story (the disaster was three clients) and is
   the only count that makes the "third is nearly free" prediction testable at
   all. The order is load-bearing and is fixed in §2.6: the alien vendor goes
   **last**, so the prediction is tested where it can actually fail.
   *Retreat position if it proves too heavy when the code is built:* grade two,
   ship Gemini as an ungraded exercise with the prediction attached. Retreat on
   evidence, not in advance.
2. ~~**Which two, if two?**~~ **MOOT** — resolved by the ruling above. Exhibit A
   needs all three to land in any case, since its whole point is that three
   vendors cannot agree on who authored a tool result.
3. ~~**Does `chat` have to work against all three vendors live?**~~ **RULED
   (2026-09-12): no live testing is required to score 100.** The grader's fakes
   are the arbiter: a solution that works against them is accepted. For readers
   who want to run live, a cheap proxy is offered so nobody has to sign up for
   three vendor accounts — bring your own key if you prefer, identical either
   way. But if you want it to work live, you have to test it live; passing
   against a fake is not a claim about production.

   **Say that last part in the prose, because this book has already been caught
   by it.** Draft 3 of this chapter asserted a vendor capability that our own
   fakes happily accepted and that turned out to be false the moment it was run
   against the real API. A fake is a *model* of a vendor, and a model is wrong
   in exactly the places you did not think to model. That is not an argument
   against fakes — they make this exercise affordable and they catch the bugs
   that matter here — it is an argument for knowing what a green grader does
   and does not prove.
4. ~~**Is `ch1parity` at 25 still right?**~~ **RULED (2026-09-12): yes, 25
   stays.** A quarter of the grade for "you did not break what you already
   had" is defensible even with the seam at 35, and it honours the standing
   guard from Chapter 1's review.
5. ~~**Should Chapter 2 state the roadmap?**~~ **RULED (2026-09-12): yes —
   as a contract, not a table of contents.** See §2.0, "The contract, stated
   once." The valuable part is not the list of coming chapters, which may be
   reordered; it is the promise that nothing here gets deleted later, and the
   instruction that follows from it: build the simplest thing that satisfies
   this chapter. A reader who distrusts the promise over-engineers
   defensively, which is the failure mode the book argues against everywhere
   else.
6. ~~**Does the goal stack belong in Chapter 2's `Context`?**~~ **RULED
   (2026-09-12): no goal stack.** It arrives with context engineering, which
   is also where the policy that needs it is defined. Deferral is cheap here
   for a structural reason worth remembering: adding a *new* field later is
   additive, while reshaping an existing one is not — which is exactly why
   `RedactData` had to be fixed now and this does not. The `RedactDialogue`
   comment no longer forward-references it.
7. **Where does the context-engineering chapter go? — TBD.** It depends on
   tool results existing (Ch3) and on the system prompt existing (Ch6 skills),
   because one of its central claims is that memory belongs in the message
   history rather than the system prompt. That puts it at Ch7 or later. Its
   *hook* — the `Redacted` event — is established here, so placement is
   genuinely flexible and need not be settled now. Ideas captured in
   `book/chapter-context-engineering-notes.md`.
8. ~~**Is `usage` still worth only 5 points?**~~ **RULED (2026-09-12): raised
   to 10, taking 5 from `seam-parse`.** It was priced when it meant "record two
   numbers." It now means normalizing four categories across three vendors that
   disagree about whether their own categories overlap — a silent bug, wrong in
   opposite directions depending on the vendor. Total seam weight is unchanged
   (the parse side is still 25 against the render side's 15); the hard part is
   simply named now, so a failing student learns *which* half broke.

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
