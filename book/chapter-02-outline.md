# Chapter 2 — The Real Data Structures (outline)

**Building Advanced AI Coding Agents** — working outline, chapter 2 of N.
Status: outline agreed in discussion 2026-09-11. Prose not yet written.
Revised 2026-09-11 to apply the lessons of `book/review.md`: the exercise
section is now specified to the level a grader can actually be built from.
Open questions 6–9 ruled (agreed as proposed). Real-time hints added as
§2.6a and as a graded requirement, on Bill's directive — raising two new
scope questions, 10–11.

**Unverified claims in this draft, for the next review pass:** the July 2025
priority claim for mid-turn hints, and the model-support split (mid-turn
`system` messages on `claude-opus-4-8` and newer but not `claude-sonnet-5`).
Both are the author's direct experience, taken on his word and written down
rather than re-derived. Model *IDs* were confirmed against `/v1/models`;
capability was not.

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

## §2.6a Real-time hints — steering an agent mid-turn

*(Author note: likely renumbers to §2.7 when prose is written. Placed here
deliberately — it is the renderer section's payoff and must follow it.)*

The agent is three tool calls into a refactor and heading somewhere you don't
want it to go. You type: *"stop, the bug is in the parser."* Those words reach
the model **during** the current turn — not after it finishes, not as the next
question. The agent pivots mid-chain, without the tool chain breaking.

This is the capability Chapter 1 §1.1 named first when it argued that
frameworks hard-code delivery. Here the student builds it.

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

| mechanism | support | quality |
|---|---|---|
| **Mid-turn `system` message** — the API's first-class path | `claude-opus-4-8` and newer (incl. `claude-opus-5`); **not** `claude-sonnet-5` | clean, no lag |
| **The original hack** — hint as a `text` block appended **after** the `tool_result` blocks in the same user message | everywhere, including Sonnet 5 | works; janky |

Both are legal. Both pass the grader. The hack is what the author used from
2025 and it has not stopped working. But note what supporting both requires:
a renderer that asks *"which model am I talking to?"* and a context that
never had to care. Put the hint logic inside your HTTP code — where it feels
natural — and you can serve exactly one model family.

**The ordering is load-bearing.** In the hack, the hint text goes *after* the
tool results, never before. A hint placed ahead of the results it is meant to
redirect reads as a comment on nothing; the model has not yet seen what it is
being steered away from.

**A hint is history, not ephemera.** This is the distinction students most
often get backwards, and §2.4 is what makes it decidable. Ephemera are
delivered once and vanish, because a stale timestamp is a lie. A hint is a
human being saying something to the agent: it is a real `MessageReceived`
event, permanent in the log, permanent in the dialogue, and it stays in every
subsequent request forever. Only its **carriage** is special — delivered once,
in a particular position, by the renderer. Delivered once; remembered always.

**Testing note, learned the hard way.** Test with `claude-opus-4-8` or newer.
On a model without mid-turn system support you will observe a one-round lag —
the steer appears to land late — and you will conclude your implementation is
broken when what you are seeing is the model. Verify the mechanism on a model
that supports it, *then* go make the hack work.

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

Rebuild the Chapter 1 chatbot on the real structures. **Observably
identical from the outside** — same REPL, same stdio grader contract —
which is itself the lesson: the rewrite bought properties, not features.

Everything Chapter 1 established still holds and is not restated in the
chapter prose: grader mode is the default (no arguments), stdout carries the
protocol only, the three `ANTHROPIC_*` variables are read and never
hardcoded, grading is free and offline, and the grader records violations
and judges afterwards rather than rejecting on the first mistake.

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
| `./ch02 render LOG` | play `LOG` → context → render; print the vendor request JSON that *would* be sent, to stdout; exit 0. **Makes no network call.** |

`render` is the centerpiece. It exposes
`EventLog → Context → vendor request` as a pure function on the command
line, which turns three otherwise-awkward properties — replay determinism,
redaction, ephemera — into mechanical byte comparisons that need no server
at all. If your architecture cannot offer this subcommand cheaply, your
context is not actually separate from your transport, and that is the
finding the exercise is designed to surface.

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
or a request builder. The grader sends it while your program is waiting on a
tool result, which means it arrives on stdin *while you are blocked on an
HTTP call you have already sent.* If your program can only read stdin between
rounds, you cannot pass this check — and discovering that is the point.

Directives are acknowledged rather than silent for the same reason Chapter 1
demands one answer per round: it keeps the stream synchronous, so a hung
submission produces a located failure instead of a bare timeout.

### Log serialization

JSON-lines, one event per line, ascending `Seq`. Each line carries at
minimum `seq`, `type`, and the event's own fields; dialogue events carry
`actor`. The format must round-trip: `dump` → `render` must work in a fresh
process with no other state. Blobs may be inlined or written beside the log,
your choice, as long as `render` needs nothing but the path it is given.

### The checks — 100 points, all must pass

| check | pts | property |
|---|---|---|
| `ch1parity` | 25 | all seven Chapter 1 checks still pass, unchanged |
| `logdump` | 5 | log serializes and round-trips; `Seq` monotonic, never reused |
| `replay` | 15 | two separate `render` invocations on the same log emit **byte-identical** requests |
| `redaction` | 15 | after `Redacted`, the payload is absent from the rendered request and present in the dumped log |
| `ephemera` | 10 | injected data appears in exactly one request, exactly once, and never in the log's dialogue events |
| `hint` | 15 | a message arriving mid-turn is classified as a hint, delivered **once**, in the correct position, and retained in history thereafter |
| `interrupt` | 10 | post-interrupt `ToolCalled` events are recorded and **not executed**; the next rendered request is legal Anthropic JSON |
| `usage` | 5 | cumulative totals derived from `ResponseEnded` events |

Notes on the ones that carry the chapter's thesis:

- **`replay`** is the deep one. It catches wandering timestamps and
  map-iteration ordering — the exact bug class that later destroys prefix
  caching, caught here, chapters before caching is mentioned. A log is a
  fixed input; if two runs of a pure function over a fixed input disagree,
  something non-deterministic leaked into your renderer.
- **`redaction`** is the pair of assertions that proves History ≠ Context in
  one line of grader code: gone from one artifact, still there in the other.
  A design that conflates them cannot pass both halves simultaneously.
- **`interrupt`** is where the reducer's totality gets tested. The grader
  interrupts mid-turn and then feeds tool calls; recording them while
  executing nothing is only possible if `Interrupted` is a real state.
- **`hint`** is graded on four separable properties, and students typically
  get three: *classified* (it is a hint, not a prompt, because of when it
  arrived), *delivered once* (not re-sent every round — that is the ephemera
  mistake applied to the wrong thing), *positioned* (after the tool results,
  never before), and *retained* (still in the dialogue ten rounds later,
  because unlike ephemera it really was said). Either delivery mechanism
  passes; the grader asserts the property, not the vendor path.
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

**Hint questions — NEW, opened 2026-09-11 (Bill's directive), awaiting ruling.**

Adding real-time hints to the exercise pulls in two things the outline did
not previously require. Both are arguably improvements; both are scope, and
scope is yours.

10. **Does Chapter 2 introduce a minimal tool?** The hint check needs a
    moment when the agent is *mid-turn with something outstanding* — which in
    practice means a tool call in flight, and the hack's delivery position is
    defined relative to `tool_result` blocks. The `interrupt` check already
    assumes `ToolCalled` events exist, so the event model is committed
    regardless; the question is whether the student's program must actually
    round-trip one trivial tool (the fake requests `echo`, the student
    returns a `tool_result`). **My leaning: yes, exactly one, and no tool
    *framework*.** Chapter 1 declared tools out of scope and that was right;
    a single hardcoded echo is not a tool system, it is a place to stand.
    Without it, hint delivery can only be demonstrated in the degenerate
    no-tool case, which is precisely the case with the one-round lag.

11. **Does the exercise require concurrent stdin?** To receive a hint while
    an HTTP request is outstanding, the program must read stdin *while
    blocked on the network*. A round-synchronous read loop — the natural
    Chapter 1 shape — structurally cannot pass. **My leaning: yes, require
    it, and say so plainly in the exercise.** This is not incidental
    difficulty: §2.5 defines an actor as identity + **mailbox** + state +
    behavior, and the mailbox has been an abstraction with nothing forcing it
    to exist. The hint is what forces it. A student who builds a goroutine
    feeding a channel has discovered why the actor model is in this chapter,
    by being unable to proceed without it.

    The risk is real and should be stated: this is the first concurrency in
    the course, and concurrency bugs are a miserable place to lose a student.
    Mitigation is the grader's existing record-then-judge discipline — the
    `hint` check reports *which* of the four properties failed
    (classified / delivered once / positioned / retained), so a student
    debugging a race gets a named property rather than "hint check failed".
