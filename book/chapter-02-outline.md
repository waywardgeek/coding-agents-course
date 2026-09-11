# Chapter 2 — The Real Data Structures (outline)

**Building Advanced AI Coding Agents** — working outline, chapter 2 of N.
Status: outline agreed in discussion 2026-09-11. Prose not yet written.

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

Grader checks (all Chapter 1 checks, plus):

1. **Replay determinism**: two separate process invocations fed the same
   event log must emit byte-identical requests. Catches wandering
   timestamps and map-iteration serialization — the class of bugs that
   later kills prefix caching, caught before caching is ever mentioned.
2. **Redaction semantics**: after a `Redacted` event, the payload is absent
   from the next rendered request but present in the serialized history.
3. **Ephemera semantics**: injected data appears in exactly one request,
   exactly once, never in the log's dialogue events.
4. **Interrupt ordering**: grader scripts an interrupt directive between
   rounds; subsequent tool-call events must be recorded in the log and
   NOT executed, and the next request must be legal Anthropic JSON (the
   renderer handled the dangling calls).
5. **Usage accounting** from `ResponseEnded` events, cumulative.

## Open questions

None as of 2026-09-11 — all resolved by ruling:

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
