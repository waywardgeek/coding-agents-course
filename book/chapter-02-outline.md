# Chapter 2 — The Real Data Structures (outline)

**Building Advanced AI Coding Agents** — working outline, chapter 2 of N.
Status: outline agreed in discussion 2026-09-11. Prose not yet written.
Revised 2026-09-11 to apply the lessons of `book/review.md`: the exercise
section is now specified to the level a grader can actually be built from.
Open questions 6–9 ruled (agreed as proposed). Real-time hints added as
§2.6a and as a graded requirement, on Bill's directive. Questions 10–11 also
ruled: the exercise gains **exactly one** trivial tool (`agent_status`,
because a turn with no tool call has no middle to steer into) and **requires
concurrent stdin** (because a hint must be received while blocked).

**Draft 3, 2026-09-11:** applies `book/review-ch02.md` — the coder's empirical
review, written from having built the grader and reference solution against
Draft 2. Nine must-fixes (M1–M9), seven enrichments (E1–E7), six standing
rulings guarded.

**Claim status after the review pass:**

- **FALSIFIED and removed:** the model-support split. Draft 2 stated that
  mid-turn `system` messages work on `claude-opus-4-8` and newer but **not**
  on `claude-sonnet-5`. Measured live, `claude-sonnet-5` accepts a mid-turn
  `{"role":"system"}` entry and obeys an instruction appearing nowhere else
  in the request — confirmed through the full rig and again through a bare
  curl returning HTTP 200. The table is replaced by a mechanism (§2.6a).
- **RELOCATED, not retracted:** the one-round lag. It is real, but it is a
  property of the *engine*, not of the carriage or the vendor. See §2.6a and
  the known boundary in the exercise.
- **STILL UNVERIFIED, and outside what any experiment here can settle:** the
  July 2025 priority claim for mid-turn hints. It is the author's, dated and
  falsifiable. Verify before print; "as far as we can determine" stays either
  way.

Predecessor: `chapter-01-outline.md` (naive Anthropic-specific chatbot,
`[]{role, content string}`, ends clean and confident — the ambush ruling).

---

## §2.0 Cold open: the demolition (no warning, per ruling)

The Chapter 1 chatbot works. Now ask it five questions it cannot answer:

1. What exactly did the model see on request #3? (Unknown — system line and
   volatile data aren't in your array.)
2. Where does a tool result go? (`role:"user"` — a lie you will tell forever.)
3. Where does a mid-turn human interjection go? (Nowhere; the format has no
   concept of "mid-turn.")
4. How do you drop a 50KB output from future requests without destroying
   your audit record? (You can't — the record and the request are the same
   object.)
5. Who said each message? (Two roles, for a world of humans, tools, other
   agents, and automation.)

**Diagnosis: the array conflates the record with the request.** Every
failure above is that one sin in different clothes.

Author's war story: an earlier CodeRhapsody was built on exactly the
Chapter 1 array — and migrating it to the structures in this chapter
touched everything, one of the most expensive refactors in the project's
history. The student gets for free the restart the author had to pay for.
Delete the Chapter 1 code. Keep the lesson.

## §2.1 History ≠ Context — the two artifacts

- **History**: the event log. Append-only, auditable, self-contained.
  Never engineered, never clever. Answers "what happened?" — but retention
  is *policy*, not principle: keep exact history as long as auditors
  require; after that, delete or compact freely (see §2.2).
- **Context**: vendor-independent state — **the state after playing all
  events in the log**. Rebuilt/updated per round trip; engineered for
  performance. Answers "what does the model see next?"
- The provider is stateless, so you own both artifacts — and they are NOT
  the same thing. Conflating them is the original sin of chat-shaped
  formats.
- Pipeline: `EventLog → (play) → Context → (render) → vendor request`.
- **Two consumers, two projections (ruling)**: the renderer consumes the
  CONTEXT — what the model sees. **The GUI consumes the EVENT LOG — never
  the context.** The human sees what happened (redacted stubs, errors,
  interrupts, tool calls recorded-but-never-executed); the model sees the
  engineered view. This is also why events carry timestamps: ordering is
  for the reducer, timestamps are for the GUI. (GUI itself deferred to the
  second half of the book; the dependency ruling lands here.)
- **Decline the vendors' stateful conversation APIs (ruling).** Anthropic,
  Google, and OpenAI all offer server-side conversation state — but if you
  want to edit history, you can't use it. In an AI coding agent, editing
  history is one of our MAIN tools (redaction of spent tool results,
  compaction, per-round context engineering), so we are forced — happily —
  to send the entire context on every round trip. The full re-send is what
  makes the outgoing copy disposable and rewriting it free; vendor-held
  state would make the context the vendor's, not ours. Full re-send is the
  price of ownership. (Its token cost is largely refunded by prefix
  caching — a later chapter.)

> **Sidebar: "Is this request mid-turn?" — a bug from building this chapter.**
>
> The grader for this chapter needs to know whether an incoming request starts
> a new turn or continues a tool loop. The obvious test: *does it contain
> `tool_result` blocks?*
>
> That is wrong, and wrong in exactly the way this section is about. Once a
> tool loop has happened, **every later request contains those tool results
> forever**, because history is re-sent in full. Four subsequent new-turn
> requests got classified as continuations of a loop that had ended three
> turns earlier.
>
> The fix is to look at what is *new*: match the freshly-typed prompt in the
> final user message. And the interrupt forces the same fix from another
> direction — the synthesized `interrupted by user` tool result (§2.6) makes a
> brand-new turn's request look like an answer to a dangling call.
>
> The general lesson, which will recur every time you touch the wire: **a
> request is not a description of the current moment.** It is the entire
> history, re-sent, with a little new material on the end. Any question you
> ask of it that sounds like "what is happening right now?" has to be asked of
> the *end* of it, or of the event log — never of the whole.

## §2.2 The event log

- Append-only. Fine-grained: one tool call is one event; one tool result is
  one event; one message from one actor is one event. Vendor batching
  (assistant message with N tool_use blocks) is *renderer output, not truth*.
- **Ordering is primary; the order determines the state of the context.**
  Timestamps are metadata for humans, never used for ordering. `Seq` is
  monotonic, never reused.
- Self-contained as a unit: log + blob table together are the whole
  conversation. (Book implementation: all in memory; blobs deduplicated by
  hash — the same DOM snapshot sent as ephemera fifty times is stored once.)
- Events are **facts, past tense** — the log records what happened, not
  commands.
- **The log is disposable; the context is not.** Because the context is the
  state after playing the log, any prefix of the past can be deleted and
  the agent runs on unaffected — going forward requires only the current
  context. Truncation costs audit, debugging, and GUI history depth (the
  GUI plays the log, §2.1) — never correctness. The
  implementation spectrum: full log (maximum auditability) → context
  snapshot + recent tail → no log at all (memory/disk lean, debugging
  hard). Replay guarantees (§2.7) apply to the retained span. (Compact,
  imperfect, searchable renderings of history for debugging and recall —
  CodeRhapsody's `history.md` — are an advanced topic for a later chapter.)

## §2.3 Events

**The self-contained event rule (load-bearing):** given the current context
and just the next event — nothing else — we can correctly compute the new
context state. Consequences:

- Classification lives in the *reducer*, not the event: `MessageReceived`
  while idle is a prompt; while mid-turn, a hint. The context must therefore
  carry turn state (§2.4).
- **The reducer is total**: a defined transition for every (state, event)
  pair; identity is legal. Two interrupts in a row: second is a no-op, not
  an error. Races (interrupt crossing an in-flight response) are normal
  orderings, not corruption.

Write the rule as:

```
newContext = Apply(context, event)
```

That notation describes **information flow**, and it is the claim the chapter
is making: everything needed to advance the context is in the context plus the
one next event. It is not a demand for value semantics. In Go the natural
implementation is a pointer receiver that mutates in place, and that is the one
to write. A `Context` containing slices, copied by value on every event, gives
two contexts sharing one backing array — and the bug that produces looks
exactly like renderer non-determinism (E3), which is a cruel thing to debug in
the same chapter that grades byte-identical replay.

**Actors, defined** (this chapter is where the book defines them): an actor
is a **persistent, stateful** entity with an identity that outlives any
single message — a name, a mailbox of ordered incoming events, state, and
behavior. Humans, agents, tools, and automation are all actors. Actors are
the unit of attribution (every event has one), of addressing (§2.5 rooms),
and of statefulness (§2.8 deployment). Actors are stateful, not stateless —
the deployment consequences land in §2.8.

**Every event has an actor:**

```go
type Actor struct {
    Kind ActorKind  // Human | Agent | System | Tool
    ID   string     // "bill", "fred", "parent", "scheduler", "run_command"
}
```

Human and Agent messages obey the same reducer rules (a parent agent's
mid-turn message is a hint by the same mechanism). System is the automation
actor ("time to do a handoff", watchdog, scheduler) — first-class, never
the engine impersonating a user.

**Content is parts, never a string** (two orthogonal axes: event type =
semantic role; parts = media):

```go
type Part interface{ isPart() }
type TextPart  struct{ Text string }
type ImagePart struct{ MediaType string; Hash BlobRef }  // png, jpeg, gif, webp
type AudioPart struct{ MediaType string; Hash BlobRef }
```

New media = new Part kind. Zero new event types, zero new methods.
(Sidebar: an earlier CodeRhapsody grew `SendUserMessageToAI` →
`...WithImages` → `...WithMedia` — the combinatorial API explosion that
happens when content is string-shaped. The migration to parts deleted two
methods and a class of bugs.)

**Taxonomy v0** (names will evolve with the book, per ruling):

Dialogue events (accumulate into context):
- `MessageReceived{actor, parts}` — any non-model actor speaks
- `AssistantMessage{parts}` — the model's output (models emit images/audio
  now too; don't bake 2024 into the name)
- `AssistantThought{parts, signature}` — extended thinking. Two-consumers
  rule applied: the thinking text/summary lives in the LOG for the GUI —
  it does NOT enter the context, because the model never sees its own past
  thinking back. What the reducer folds into the context is only the
  vendor's opaque REPLAY MATERIAL (Anthropic: signed thinking block,
  replayed within the current tool loop; Gemini: encrypted thought
  signature) — renderer-bound vendor quirks, quarantined per §2.6
- `ToolCalled{call_id, name, args}` — args are structured JSON, not media
- `ToolReturned{call_id, parts, is_error}` — tool output IS media
  (read_file on a PNG, screenshot); `is_error` marks semantic failure the
  model must see (§2.4)

Control events (change what the context is):
- `Interrupted{}` — see state machine
- `RequestSent{code_commit, request_hash}` — the round-trip boundary;
  consumes ephemera; see §2.7
- `ResponseEnded{stop_reason, usage}` — usage is money, recorded in the log
- `ErrorOccurred{source, code, message, related_seq}` — errors carry their
  ORIGIN: `Source ∈ {Provider, Renderer, Tool, Engine}`, a machine-readable
  code ("429", "ETIMEDOUT", "unknown_event_type"), human-readable text, and
  the seq of the event the error pertains to (the RequestSent, the
  ToolCalled). One event type, source enum — the reducer switches on
  source; totality holds per (state, source) pair. Control event: changes
  state, contributes zero content.
- `EphemeraSet{instruction, parts}` — REPLACES prior ephemera (volatility
  is reducer semantics, not a storage hack)
- `Redacted{target_seq}` — the decision is itself an auditable event;
  replay reproduces its effect exactly
- `ModelChanged{model}`, `SystemPromptChanged{...}` — policy is state too
- Session event: binary (re)attached to the log with a new code commit

## §2.4 The context

Vendor-independent. NO vendor rules contaminate it — canonical example in
§2.6 (interrupt + dangling tool calls).

**Turn-state machine:**

```
Idle          — no turn open
InputPending  — input arrived, no request sent (also: retry posture after
                transport error). NOT a commitment to respond.
InFlight      — RequestSent, awaiting response events
ToolsPending  — response ended in tool_use; results accumulating
Interrupted   — turn killed; late events APPEND but execute nothing
```

Key transitions:

| state × event | → | note |
|---|---|---|
| Idle × MessageReceived | InputPending | classified: prompt |
| InputPending × MessageReceived | InputPending | input accumulates (rooms) |
| InFlight × MessageReceived | InFlight | classified: hint |
| ToolsPending × MessageReceived | ToolsPending | classified: hint — a human types while tools are running |
| InputPending × RequestSent | InFlight | engine decided to respond |
| InFlight × ResponseEnded(end_turn) | Idle | |
| InFlight × ResponseEnded(tool_use) | ToolsPending | |
| ToolsPending × ToolReturned | ToolsPending | engine sends next RequestSent when drained |
| InFlight × ErrorOccurred(Provider) | InputPending | retry posture; model never hears of a cured 429 |
| InFlight × Interrupted | Interrupted | canonical case |
| ToolsPending × Interrupted | Interrupted | after_tool: in-flight tools drain here |
| Interrupted × ToolCalled | Interrupted | **recorded, never executed** — truth without action |
| Interrupted × ResponseEnded | Idle | dead turn fully drained |
| Interrupted × Interrupted | Interrupted | no-op, defined |
| Idle × Interrupted | Idle | no-op, defined — nothing to kill |

**Every pair not listed is identity.** That sentence, not the length of the
table, is what makes the reducer total. A table can only ever enumerate the
transitions we thought of; the default is what covers the ones we did not.
Write it as the `default` arm of the switch, not as a `panic`.

Proof `Interrupted` must be a state: without it, tool calls arriving after
an interrupt are indistinguishable from a live turn's, and the
self-contained reducer would start executing them.

**Errors split in two, and the split is the lesson:**
- *Infrastructure* errors (`ErrorOccurred` — provider 429/timeout, renderer
  refusal, engine faults) change **state** (Provider: → InputPending),
  enter the log for audit with source + code + related_seq, contribute
  zero content.
- *Semantic* errors (tool ran and crashed) change **content**: a
  `ToolReturned{IsError: true}` whose parts say "exit 1, stderr: ..." —
  the model must see it. (`IsError` maps to Anthropic's `is_error`; the
  renderer never infers failure from prose.)
- Boundary case proving the split: a tool that couldn't even START (not
  found, sandbox denied) still emits `ToolReturned{IsError: true}` — the
  model has a dangling call awaiting a result — plus optionally
  `ErrorOccurred{Source: Tool}` carrying the infra detail for audit,
  linked by related_seq.

## §2.5 Multi-party: rooms

- `Actor.ID` distinguishes Bill from Fred. Concurrent senders serialize
  into one ordered thread; seq ordering IS the multi-party semantics.
- The reducer is sender-agnostic: Fred mid-turn is a hint; Fred while idle
  accumulates input. No per-actor special cases.
- **A conversation is a room. There is no `To` field.** Everyone in the
  room hears everything; whispering is a *different room* (a different
  conversation), not a flag on an event. Routing belongs to the mailbox
  layer, outside the log.
- Room-awareness: input accumulates in InputPending; the ENGINE decides
  whether/when to respond (emit RequestSent). Use case: Discord voice
  transcription during gameplay — the LLM stays informed of the whole room
  and infers from context whether it's expected to speak ("Fred says ...").

## §2.6 The renderer

- One-way: Context → vendor JSON at request time. Responses parsed into
  events at the boundary; vendor types never leak inward.
- All of Chapter 1's "architecture" reappears here, demoted to one
  backend's quirks: role derivation, alternation rules, system placement.
- **Attribution policy** (per ruling): renderer's decision. Anthropic
  default: solo human renders as plain "user" with no ID; otherwise the
  renderer inserts sender metadata into the user content ("Fred says ...").
- **The canonical vendor-quirk example**: interrupt arrives while a
  response with tool calls is in flight. The log shows Interrupted BEFORE
  ToolCalled. The context truthfully holds an assistant message with
  dangling tool calls, state Idle, new user prompt arriving — which is
  ILLEGAL Anthropic JSON. That's Anthropic's quirk, and the Anthropic
  renderer eats it. Two policies discussed: strip the dangling tool_use
  blocks, or synthesize `tool_result: "interrupted by user"` blocks
  (proposed default — the model behaves better when it knows it was cut
  off). Renderer policy; zero contamination of the context.
- Media asymmetry is a renderer property: an audio part sent to a
  text-only model is a LOUD error, never a silent drop (no-fallbacks
  discipline).

**Show the reader what actually comes out.** The abstract description above
under-sells how strange the result looks, and students will meet this JSON in
their own output:

```json
{
  "role": "user",
  "content": [
    {
      "type": "tool_result",
      "tool_use_id": "toolu_...",
      "content": "interrupted by user"
    },
    {
      "type": "text",
      "text": "stop poking at that and check the config instead"
    }
  ]
}
```

A synthesized answer to a question the human never let finish, and the human's
next instruction, **merged into a single message** — because Anthropic will not
accept two user messages in a row. Nothing in the context looks like this. The
context holds an interrupt, a dangling tool call, and a new prompt, each a
separate honest fact in sequence. The merge exists solely because one vendor's
wire format demands alternation. That is the whole of §2.6 in one object: the
distortion lives in the renderer and dies there.

## §2.6a Real-time hints — steering an agent mid-turn

*(Author note: likely renumbers to §2.7 when prose is written. Placed here
deliberately — it is the renderer section's payoff and must follow it.)*

The agent is three tool calls into a refactor and heading somewhere you don't
want it to go. You type: *"stop, the bug is in the parser."* Those words reach
the model **during** the current turn — not after it finishes, not as the next
question. The agent pivots mid-chain, without the tool chain breaking.

This is the capability Chapter 1 §1.1 named first when it argued that
frameworks hard-code delivery. Here the student builds it.

**Why this chapter needs one tool.** A hint is a message that arrives *during*
a turn — so a turn has to be long enough to have a middle. With no tools a
turn is a single round trip: request, response, done. There is no middle, and
mid-turn steering degenerates into "your next question." That is why the
exercise hardcodes exactly one trivial tool (`agent_status`, below). It is
not a tool chapter; it is the smallest possible thing that makes a turn last
long enough for a human to interrupt it.

**The demonstration.** The model calls `agent_status`. You return the status.
It calls it again. And again — a loop that will happily run until something
stops it. Mid-loop, you type:

> **please stop**

The words reach the model *inside* the turn, attached to the very next tool
result. The loop ends. You did not kill the process, you did not wait for
your turn — you steered a running agent, and you built every part of the path
those words travelled.

In the REPL, try **"Please speak like a pirate"** mid-turn instead, and watch
the rest of the turn come back in pirate. Same mechanism, more fun, and it
makes the timing vivid: everything before the hint is normal, everything
after is piratical, in a single unbroken turn.

**Provenance.** As far as we can determine, this book's author and his agent
invented this in **July 2025** — the first known implementation of mid-turn
human steering in an agentic loop. It came from noticing that the Messages API
would accept a user `text` block in the *same message* as `tool_result`
blocks, which meant there was a place to put words that the model would read
before deciding its next move. Frontier products have since shipped
steering in their own UIs, and the API has since grown a first-class mechanism
for it (below). *(Publication note: the priority claim is the author's, dated
and falsifiable. Verify before print; "as far as we can determine" stays
either way.)*

### Why it belongs in this chapter, and not in the chapter about tools

Because the hard part is not the network call — it is that a hint is
**the same event as a prompt, distinguished only by state.**

§2.3's reducer table already says so: `InFlight × MessageReceived → InFlight`,
classified as a hint. There is no `Hint` event type and there must not be one.
The identical bytes typed by the identical human are a *prompt* when the turn
is `Idle` and a *hint* when it is `InFlight`. Only the reducer knows which,
because only the reducer holds the state. Classify at capture time and you
will be wrong every time the human types fast.

The second half is delivery, and delivery is the **renderer's** problem —
which is what makes this the cleanest demonstration of §2.6's rule. The
context records a fact about the conversation: *a hint is pending, not yet
carried to the model.* It does not record how to carry it. That separation is
not an aesthetic preference; it is load-bearing, because there are currently
**two** ways to carry a hint and which one you use depends on the model:

| carriage | how it is carried | availability |
|---|---|---|
| **Mid-turn `system` message** | a `{"role":"system"}` entry inside `messages`, placed after the current tool results | accepted by every current Claude model we have tested, `claude-sonnet-5` and `claude-opus-4-8` included |
| **The appended text block** (the original 2025 hack) | the hint as a `text` block placed **after** the `tool_result` blocks in the same user message | everywhere, by construction — it is an ordinary user message |

Both are legal, both steer the model, and **the renderer chooses**. Notice what
supporting both costs: a renderer that knows which carriage it is using, and a
context that never had to care. Put the carriage decision inside your HTTP
code, where it will feel natural, and you can serve exactly one shape of model.

> **A note on what this table used to say.** An earlier draft claimed the
> mid-turn `system` message worked on `claude-opus-4-8` and newer but *not* on
> `claude-sonnet-5`, and that the hack cost you a round of lag. Measured
> against the live API, that is wrong on both counts: `claude-sonnet-5`
> accepts the mid-turn `system` entry and obeys an instruction that appears
> nowhere else in the request, and both carriages steer the model on the very
> first response after the hint is carried. The lag was real. It was not where
> we thought. Chasing it down is the rest of this section — and the reason the
> table above deliberately does not tell you which models support what. Any
> such table is a fact about one week in the history of an API. The mechanism
> below outlives it.

**The ordering is load-bearing.** In the hack, the hint text goes *after* the
tool results in the message. A hint placed before them reads as a comment on
nothing — the model has not yet seen what it is being steered about. This is a
real rule with a real failure mode, it is graded, and it is the reason
`agent_status` had to exist at all: an ordering rule needs two things to order.

**A pending hint is not in the dialogue yet.** This is the one place where a
student who has understood everything else will still fail, so it is worth
being slow and explicit.

Look at the timing. The hint arrives *during* request N — that is the
definition of a hint. But it must be carried *after the tool results* of
request N+1, and at the moment it arrives, those results **do not exist yet**.
In `Seq` order, the hint sits before the `ToolCalled` and `ToolReturned` events
of the very message it is supposed to follow.

So the obvious, correct-sounding thing — play the log in order, render the
dialogue in order — puts the hint in the wrong place. The student's reducer is
right. Their renderer is right. Their reading of §2.2 is right. The result is
wrong.

The resolution is §2.6's rule paying for itself in cash:

> A pending hint is **not in the dialogue**. The context records only that a
> hint is *pending*. The renderer places it at the end of the final user
> message of whichever request carries it — and `RequestSent` is the event
> that moves it into the dialogue, at that same end position. From that moment
> it is ordinary history and never moves again.

That single decision buys three things at once: the hint lands in the right
position on delivery, the history is stable afterwards (which prefix caching
will demand in a later chapter), and replay puts it in exactly the same place
every time.

**Render, then record.** A direct corollary, and worth stating as a rule
because the failure is so well disguised: the request is rendered **first**,
and `RequestSent` is recorded **after**. Rendering is what *carries* the
pending hint and the pending ephemera; `RequestSent` is what *consumes* them.
Do it in the other order and the hint goes out one round late — which is
indistinguishable, from the outside, from a model that lags. A student who hits
this will blame the vendor. It is in their own engine, four lines apart.

**A hint is history, not ephemera.** This is the distinction students most
reliably get backwards, and the two graded properties need saying precisely,
because read casually they contradict each other:

- **Delivered once** means the hint appears **at most once within any single
  request** — not duplicated inside a request, and not re-attached as a *fresh*
  delivery on each subsequent round.
- **Retained** means that once delivered it stays in the dialogue, and is still
  there in a request sent many turns later.

Both are true simultaneously because the hint becomes an ordinary part of a
user message, and history is re-sent in full every round (§2.1). Its words
therefore *do* appear in every later request, forever — and that is not the
ephemera mistake, that is the definition of retention. Contrast a stale
timestamp, which must vanish after one delivery because on the next round it
is simply a lie. **A hint is delivered once and remembered always.**

**What actually determines responsiveness.** Here is the mechanism the model
table cannot give you:

> A hint is carried by **the next request the engine sends**. Anything that
> delays that request delays the steer.

That reframes the question from "which carriage?" to "what is my engine doing
right now?", and there are only three answers:

1. **A response is already in flight when you type.** Irreducible. The HTTP
   request has left; nothing can overtake it. You wait for that one response.
   No carriage fixes this, and no carriage needs to — it is one response, not
   one round.
2. **No further request is coming.** If the turn is ending, there is no next
   request to carry anything, and the hint becomes what it always was for a
   plain chatbot: your next message.
3. **The engine cannot hear you while it works.** This is the big one, and it
   is the one that masquerades as vendor lag.

That third case is measurable. Patch the chapter's one tool to sleep fifteen
seconds and type a hint five seconds in:

| carriage | hint typed | hint *received* | mailbox blackout |
|---|---|---|---|
| appended text block | t+5.0s | t+16.4s | **+11.4s** |
| mid-turn `system` | t+5.0s | t+16.3s | **+11.3s** |

Identical, because the delay is not in the API at all. It is in the engine: if
tools execute on the same goroutine that drains the mailbox, then for the whole
duration of a tool the actor is **deaf**. Both carriages then deliver on the
first request after the tool returns, promptly and equally.

(These are single runs against one vendor on one day, quoted to show a shape,
not to characterise a model. Re-measure before you trust any digit here.)

Which lands us somewhere better than a compatibility table. An actor whose
mailbox goes deaf whenever it does work does not really have a mailbox — it has
an inbox it checks between chores. Chapter 2's tool is instant, so the chapter
cannot show you this failure. Chapter 3's tools are not, and that is when the
mailbox starts to earn its keep.

## §2.7 Versioning and replay

- **Replay-with-current-code**: replaying the log through today's reducer +
  renderer. We do not version reducers.
- **The semver contract on reducer semantics**: within a major version,
  replay is sacred — additive evolution only (new event types allowed; the
  transition semantics of existing types are FROZEN); every old log plays
  to the identical state. Across major versions, replay may break, loudly
  and deliberately. **Caveat (ruling): 0.x.x promises nothing** — replay
  compatibility begins at 1.0. Students live on 0.x all semester; the
  discipline is the lesson, not the guarantee.
- Evidence over promises (C's history of reinterpreting "the same code"
  justifies paranoia): `RequestSent{code_commit, request_hash}` — git
  commit (SHA-256 object format) + sha256 of the rendered request bytes.
  Drift is detectable AND attributable. A session event records each binary
  (re)attach, so mid-conversation upgrades are visible history.
- **Unknown event type ⇒ refuse to load, loudly** — name the event type
  and the commit that wrote it. Silently skipping an event would produce a
  context that is subtly wrong while claiming replay-exactness (the
  `void* /* generator */` of conversation formats).

## §2.8 Actors are stateful — deployment follows the data structures

- The context is live, expensive-to-rebuild state. Where it LIVES is a
  property of the data structures, not a free deployment choice.
- Two topologies seen in the wild:
  1. **Stateless workers**: every message rehydrates the full conversation
     from durable transactional storage, processes, writes back. (Some
     frameworks require this.)
  2. **Stateful actors** (ruling: preferred): the conversation lives in
     memory on one server; the load balancer routes a conversation's
     messages consistently to the same machine.
- Why stateful wins: no per-message rehydration cost — and, decisively,
  **ancillary local state**. Example: a tool result too large for the
  context is not dropped; it's kept on local disk, a stub goes in the
  context, and the LLM can scan the full output later via tool calls. With
  random routing, the file is on the wrong machine and this pattern —
  one of the most valuable in a real coding agent — becomes a distributed-
  systems problem.
- **Actors persist. They migrate, and migration is hard** — snapshot the
  context + blob table + ancillary files, move, reattach — but that is the
  primitive worth building, not a reason to go stateless.
- Local single-machine deployment (how this course builds) is the
  degenerate case where stickiness is free. The cloud consequences are
  named here so the student knows what the local design is secretly
  deciding.

## Exercise (auto-graded, Go, fake Anthropic server)

Rebuild the Chapter 1 chatbot on the real structures. For everything
Chapter 1 could already do it is **observably identical from the outside** —
same REPL, same stdio contract, same answers — which is the lesson: the
rewrite bought properties, not features.

Then it does two things the Chapter 1 program could not have expressed at
any price: it accepts a **hint** mid-turn, and it survives an **interrupt**.
Both were impossible before, and neither needed a new idea — they fell out of
the structures. That is the argument of this chapter in one sentence.

**Exactly one tool, and it is not a tool system.** Chapter 3 builds the tool
loop properly. This exercise hardcodes a single zero-argument tool,
`agent_status`, which returns a small JSON object describing the
conversation's own state — the highest `Seq` in the log and the current turn
state. It is deliberately self-referential: the agent's one capability is to
look at the structure you just built.

It exists because **without a tool, a turn is one round trip, and there is no
"mid-turn" to steer into.** The fake drives `agent_status` in a loop — calling
it again, and again — which is what gives you a long turn with a human-shaped
window in the middle of it. Its return value must be a deterministic function
of the log, because `replay` demands byte-identical renders. It returns
something as small as this:

```json
{"highest_seq": 41, "turn_state": "ToolsPending"}
```

What you are *not* building: a tool registry, a dispatch table, schemas,
argument validation, or error plumbing. One `if` statement on the tool name is
the correct amount of machinery here. **You do still declare the tool in the
request** — a real Messages API never emits a `tool_use` for a tool the request
did not advertise, and a submission that omits the `tools` array is relying on
the fake's good manners rather than on the protocol. The declaration is three
hardcoded lines. "No registry" means no machinery for *managing* tools, not
the absence of the field that makes the one tool legal.

### The problem this section exists to solve

Chapter 1's review taught a lesson at the author's expense: **an exercise
spec that a grader cannot be built from is not a spec.** Four of that
chapter's seven graded properties were never stated, and competent students
would have failed for reasons the chapter never mentioned.

Chapter 2 is worse in kind, not degree. "The grader scripts an interrupt
between rounds" presumes a wire format for interrupts that does not exist.
"Two invocations fed the same event log" presumes a log serialization and a
way to feed it, neither of which exists. Those are specified below.

### Surfaces the submission must expose

Three, and the last two are new this chapter:

| invocation | behavior |
|---|---|
| `./ch02` | grader mode — stdio JSON-lines (the default, as in Chapter 1) |
| `./ch02 chat` | the REPL |
| `./ch02 render LOG` | play `LOG` → context → render; print the vendor request JSON that *would* be sent to stdout, and **nothing else on stdout**; exit `0`. **Makes no network call.** |

`render` is the centerpiece. It exposes
`EventLog → Context → vendor request` as a pure function on the command
line, which turns three otherwise-awkward properties — replay determinism,
redaction, ephemera — into mechanical byte comparisons that need no server
at all. If your architecture cannot offer this subcommand cheaply, your
context is not actually separate from your transport, and that is the
finding the exercise is designed to surface.

`render` takes **no flags**. The model id and any other request parameters
come from the environment, exactly as they do in grader mode. This is not
fussiness: two `render` invocations are compared byte for byte, so the moment
rendering accepts a `--model` or a `--max-tokens`, byte-identity becomes a
property of *how you invoked the command* rather than a property of the log.
The log is supposed to be the whole input. Keep it that way.

### Control directives (extends the Chapter 1 stdio protocol)

The grader writes one JSON object per line. `{"user": ...}` behaves exactly
as in Chapter 1 and expects exactly one `{"assistant": ...}` reply. Every
other object is a **directive** and expects exactly one `{"ok": true}`
acknowledgement line.

| directive | event injected |
|---|---|
| `{"interrupt": true}` | `Interrupted{}` |
| `{"hint": "..."}` | `MessageReceived` — arrives while a turn is in flight, so the reducer must classify it as a hint |
| `{"redact": <seq>}` | `Redacted{target_seq}` |
| `{"ephemera": {"instruction": "...", "text": "..."}}` | `EphemeraSet` |
| `{"dump": "<path>"}` | none — serialize the event log to `<path>`, flush, then ack |

The `hint` directive is the one that tests whether you built a conversation
or a request builder. The grader sends it **mid-loop** — while the fake is
driving `agent_status` and your program is blocked on an HTTP call it has
already sent. The line arrives on stdin at a moment when a round-synchronous
program is not listening. If you can only read stdin between rounds you
cannot even receive this directive, let alone act on it. §2.5 already told
you the answer: an actor has a mailbox.

The grader's hint is **"please stop"**, and the fake enforces the whole
contract behaviorally: it keeps calling `agent_status` until it sees that
text correctly delivered, and then stops. A submission that drops the hint,
delivers it in the wrong position, or saves it for the next round keeps the
loop running — and the loop running is itself the failure, visible without
reading a single check name.

*(In the REPL, students should try **"Please speak like a pirate"** instead.
Same path, and the turn audibly changes character halfway through.)*

Directives are acknowledged rather than silent for the same reason Chapter 1
demands one answer per round: it keeps the stream synchronous, so a hung
submission produces a located failure instead of a bare timeout.

### Log serialization

JSON-lines, one event per line, ascending `Seq`. Each line carries at
minimum `seq`, `type`, and the event's own fields; dialogue events carry
`actor`. The format must round-trip: `dump` → `render` must work in a fresh
process with no other state. Blobs may be inlined or written beside the log,
your choice, as long as `render` needs nothing but the path it is given.

**The event-type vocabulary is frozen for this exercise.** §2.3 says the
taxonomy is v0 and will evolve with the book — true, and it does not apply
here. Three checks (`interrupt`, `redaction`, `usage`) assert on the
*contents* of your dumped log: that an `Interrupted` event exists, that a
`ToolCalled` follows it and no `ToolReturned` does, that a `Redacted` event
names its target, that `ResponseEnded` events carry token counts. None of that
is possible unless we agree on names. Use the §2.3 set:

`MessageReceived`, `RequestSent`, `ResponseStarted`, `ResponseEnded`,
`ToolCalled`, `ToolReturned`, `Interrupted`, `Redacted`, `ErrorOccurred`.

Spell them however you like. The grader compares type names **lowercased with
punctuation stripped**, so `ToolCalled`, `tool_called` and `TOOL-CALLED` are
the same event, and it does the same for field names, so `target_seq` and
`targetSeq` both resolve. What it cannot do is guess that you called it
`Halted`. The spelling is yours; the vocabulary is not.

### The checks — 100 points, all must pass

| check | pts | property |
|---|---|---|
| `session` | 0 | the stdio protocol was honoured: every directive acknowledged, no unexpected lines on stdout, and a census of every request the fake saw |
| `ch1parity` | 25 | all seven Chapter 1 checks still pass, unchanged |
| `logdump` | 5 | log serializes and round-trips; `Seq` monotonic, never reused |
| `replay` | 15 | two separate `render` invocations on the same log emit **byte-identical** requests |
| `redaction` | 15 | after `Redacted`, the payload is absent from the rendered request and present in the dumped log |
| `ephemera` | 10 | injected data appears in exactly one request, exactly once, and never in the log's dialogue events |
| `hint` | 15 | a message arriving mid-turn is received while blocked, classified as a hint, delivered **once**, **after** the tool results, and retained in history thereafter |
| `interrupt` | 10 | post-interrupt `ToolCalled` events are recorded and **not executed**; the next rendered request is legal Anthropic JSON, and the next message is a prompt, not a hint |
| `usage` | 5 | cumulative totals derived from `ResponseEnded` events |

Notes on the ones that carry the chapter's thesis:

- **`replay`** is the deep one. It catches wandering timestamps and
  map-iteration ordering — the exact bug class that later destroys prefix
  caching, caught here, chapters before caching is mentioned. A log is a
  fixed input; if two runs of a pure function over a fixed input disagree,
  something non-deterministic leaked into your renderer.

  The leak is worth naming in advance, because the failure message tells you
  only *that* two renders differed, and hunting a one-byte difference can eat
  an afternoon. There are essentially four ways it gets in: **the clock**, a
  **randomly generated id**, **Go's deliberately randomized map iteration
  order**, and **iteration over a set**. The last two are the same bug wearing
  different hats, and they are the reason to build your wire types out of
  structs with ordered fields rather than `map[string]any` — the map will
  serialize in a different order on some future run, on some future machine,
  and never on the one where you tested it.
- **`redaction`** is the pair of assertions that proves History ≠ Context in
  one line of grader code: gone from one artifact, still there in the other.
  A design that conflates them cannot pass both halves simultaneously.
- **`interrupt`** is where the reducer's totality gets tested. The grader
  interrupts mid-loop and the model's already-issued `agent_status` call
  still arrives; it must be **recorded and not executed**. That is only
  expressible if `Interrupted` is a genuine state rather than a boolean
  someone remembers to check — and a second interrupt on an
  already-interrupted turn must be a legal no-op, the identity transition
  §2.3's table promises. The renderer then has to make a context containing
  an unanswered tool call into legal Anthropic JSON, which is §2.6's rule
  earning its keep.

  **One rule the chapter owes you, because Chapter 1's contract does not
  cover it.** Chapter 1 required exactly one `{"assistant": ...}` line per
  `{"user": ...}` line. Chapter 2 kills a turn in the middle — so does that
  round still owe its line? The rule: **a turn killed by an interrupt produces
  no `{"assistant"}` line.** A killed turn produced no answer, and saying so
  is more honest than inventing one. The grader tolerates a line if your
  design prefers to emit the partial text it did receive, but decide which you
  are doing on purpose. Get this wrong by accident and your stream
  desynchronizes against the grader's, turning one defect into a cascade of
  timeouts that all look like different bugs.
- **`hint`** is graded on five separable properties, and students typically
  get four: *received while blocked* (acked mid-loop — proof the program has
  a mailbox and not a read loop), *classified* (it is a hint, not a prompt,
  because of **when** it arrived, so it does not start a turn of its own or
  earn its own `{"assistant"}` reply), *positioned* (after the tool results,
  never before), *delivered once* (it appears **at most once within any single
  request** — not duplicated inside a request, and not re-attached as a fresh
  delivery each round), and
  *retained* (still in the dialogue ten rounds later, because unlike ephemera
  it really was said). Either delivery mechanism passes; the grader asserts
  the property, not the vendor path.

  **A known boundary, stated plainly because it undercuts §2.5's own thesis.**
  This check proves your program stays responsive while blocked on *HTTP*,
  which is the easier half. It cannot prove your program stays responsive
  while blocked on *work*. `agent_status` is instant by specification — it has
  to be, since `replay` byte-compares its output — so a submission that runs
  its tool on the same goroutine that drains the mailbox passes every check
  here and still goes deaf for the duration of every real tool it ever runs
  (§2.6a measures exactly that, at eleven seconds). Chapter 3 runs tools off
  the engine goroutine, and that is where the mailbox starts to earn its keep.
  We are not fixing it here, because the fix belongs with the tool loop and
  this is not the tool chapter.
- **`session`** is worth zero points and can still sink a submission. It
  exists because Chapter 2 *extends* the stdio contract — five directives and
  an acknowledgement — and Draft 2 had no check that the stream itself
  behaved. Without it, a submission that simply fails to acknowledge a
  directive fails `logdump`, `redaction`, `ephemera`, `hint` and `interrupt`
  all at once, and the student gets five mysteries instead of one cause. Zero
  points keeps the table at exactly 100 without reopening the weighting.
  *(This adds a check the outline's ruling did not list — flagged for
  ratification, not slipped in. It could equally carry real points taken from
  `ch1parity`, which is the thing it most resembles.)*
- **`ch1parity`** is deliberately worth less than the sum of its parts. The
  rewrite is not supposed to buy features — it is supposed to keep them while
  buying properties. *(Reweighted 30 → 25, with `logdump` 10 → 5 and `usage`
  10 → 5, to fund the `hint` check at 15. The 30/70 ruling's signal is
  preserved at 25/75; flagging the change because it edits a ruling.)*

## Open questions

**Design questions — all resolved by ruling 2026-09-11:**

1. **Room/InputPending**: CONFIRMED — input accumulates; the engine decides
   whether to respond; InputPending is not a commitment. (Homebrew-VTT use
   case: the LLM must NOT respond unless explicitly spoken to.)
2. **Dangling tool calls after interrupt**: a RENDERING decision — the
   context data structures never carry it. Synthesize-"interrupted"
   confirmed as the Anthropic default.
3. **Error shape**: single `ErrorOccurred` with source enum, confirmed.
4. **Thinking**: `AssistantThought{parts, signature}` events — thinking
   text lives in the log (GUI); only opaque vendor replay material
   (signature / signed block) enters the context.
5. Naming evolves as the book and the resulting agent get written
   (taxonomy v0).

**Exercise-protocol questions — opened and RESOLVED 2026-09-11.**

These arose from applying the Chapter 1 review to this chapter. Bill ruled on
all four together: **agreed as proposed.** They are kept below with their
reasoning intact, because each constrains the grader and a future reader
deserves to know the alternative that was considered and declined.

6. **The `render` subcommand.** I made `EventLog → rendered request` a CLI
   surface (`./ch02 render LOG`) so replay, redaction and ephemera become
   byte comparisons needing no server. It is the single biggest addition to
   the exercise, and it constrains student architecture — it forces context
   to be genuinely separable from transport. I think that constraint is the
   chapter's thesis made executable rather than an imposition, but it is a
   real constraint and it is yours to accept or reject.
7. **Directives are acknowledged (`{"ok": true}`), not silent.** Keeps the
   stdio stream synchronous so a hung submission fails with a location
   instead of a timeout. Costs a small departure from Chapter 1's protocol,
   which the student must notice.
8. **Log format: JSON-lines, one event per line, ascending `Seq`.** Chosen
   for greppability and so the grader can assert on the log with the same
   tools it uses for requests. The alternative — a single JSON document —
   round-trips just as well but is worse to debug at 3am.
9. **Point weighting: `ch1parity` 30, the five new checks 70.** The
   deliberate signal is that a rewrite which merely preserves Chapter 1's
   behavior has not earned the chapter. If you'd rather parity dominate
   (a rewrite that breaks the old contract is a failed rewrite, full stop),
   the split should invert.

**Not yet specified, and deliberately so:** the scripted session itself — how
many rounds, where the interrupt lands, which `Seq` gets redacted. That is
grader construction, not chapter content, and it should be written with the
rig in front of us rather than guessed at here. Chapter 1's script was chosen
that way and the memory probe's design (plant in round 1, check in round 4)
came out of building it, not out of the outline.

**Hint questions — opened and RULED 2026-09-11.**

Adding real-time hints pulled in two things the outline did not previously
require. Both were scope questions; both are now ruled.

10. **Does Chapter 2 introduce a minimal tool? — RULED: YES. Exactly one:
    `agent_status`.** Ruled twice, and the reversal is the instructive part.

    The first ruling was *no — tools are Chapter 3*, on the sound instinct
    that Chapter 2 is already a demolition and a rebuild and should not also
    become the tool chapter. I wrote it up, re-grounded the `hint` and
    `interrupt` checks on tool-free forms, and in doing so produced the
    argument against it: **without a tool, a turn is one round trip, so a
    turn has no middle, and "mid-turn steering" has nothing to steer.** The
    hint degenerates into "your next message," which is precisely the thing
    hints are not. The feature cannot be taught in a chapter that cannot
    produce a long turn.

    So: one zero-argument tool, `agent_status`, returning a deterministic
    function of the log. No registry, no schemas, no dispatch — Chapter 3
    builds the tool loop properly. This one exists to make a turn long enough
    to interrupt. The fake drives it in a loop until the hint **"please
    stop"** arrives correctly delivered, which makes the pass condition
    behavioral: get it wrong and the loop simply keeps going.

    Worth recording for the book's own sake: the scope instinct ("that
    belongs in a later chapter") was right in general and wrong here, and the
    way it got caught was writing the consequence down until it contradicted
    itself. That is the same move the Chapter 1 review made — build the
    artifact, let it tell you the spec is wrong.

11. **Does the exercise require concurrent stdin? — RULED: YES.** The program
    must read stdin while blocked on the network, so a round-synchronous
    Chapter 1-shaped loop cannot pass. This is now graded directly: the fake
    holds its reply open, and the *received while blocked* property of the
    `hint` check is satisfied only by a submission that acks the directive
    during that window.

    This is not incidental difficulty. §2.5 defines an actor as identity +
    **mailbox** + state + behavior, and until now nothing forced the mailbox
    to exist. The hint forces it. A student who ends up with a goroutine
    feeding a channel has discovered why the actor model is in this chapter
    by being unable to proceed without it.

    The risk stands and should be stated in the prose: this is the first
    concurrency in the course, and a race is a miserable place to lose a
    student. The mitigation is that the `hint` check reports *which* of its
    four properties failed, so a student gets a named property rather than
    "hint check failed".

