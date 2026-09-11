# Chapter 1 — A Conversation, the Obvious Way (Outline, Draft 1)

**Book:** Building Advanced AI Coding Agents (working title)
**Status:** outline for discussion. 2026-09-11.
**Chapter thesis:** Hold a real multi-round conversation with Claude using the
least machinery possible — the role-tagged message array that every tutorial,
every framework, and the provider's own documentation teach. It works, and the
chapter ends clean and confident. (Ruling: **ambush** — Chapter 1 gives no
hint that Chapter 2 throws this code away.)

---

## 1.0 Opening — the gold rush, and the bet (THE BOOK'S OPENING)

Before any code: what an AI coding agent is *worth*, and the story of how
this book came to exist. The valuations are insane, and the insanity is the
hook.

**The gold rush:**

- **SpaceX acquires Cursor (Anysphere), ~$60B all-stock, June 2026** — an AI
  code editor valued above most aerospace companies.
- **The Windsurf drama** — one product, three suitors, three deal shapes:
  OpenAI's ~$3B buyout collapses → Google takes leadership + IP in a ~$2.4B
  licensing/acqui-hire → Cognition buys the remaining product, brand, and
  enterprise base.
- **OpenAI buys Ona** (secure cloud execution & orchestration, for Codex);
  **Anthropic buys Bun** (high-performance JS runtime, for Claude Code).

**The bet (the author's story, told straight):** The Windsurf acquisition is
what broke me. Billions for *that*? I told my team I could write a better AI
coding agent PoC than Windsurf in two weeks. My manager told me to prove it.
I did. The result was StackAgent — a vibe-coded pile of shit I had to throw
away. The two weeks were not wasted: the value was never the code, it was
what I learned building it. I then spent a month writing the initial version
of CodeRhapsody properly — production-worthy — and that is the direct
ancestor of the agent that is helping me write this book right now.

**Why the story opens the book:**

1. It sets the stakes in dollars and the craft in reach: one experienced
   engineer, two weeks, a working PoC in a space trading for billions.
2. It states the book's core thesis in miniature: **the code was disposable;
   the learning was the asset.** (The reader will live this arc themselves —
   more than once.)
3. The pattern that matters: acquirers are buying the **execution layer** —
   runtimes, orchestration, secure execution — not chat wrappers. The value
   in an AI coding agent is the systems layer underneath the conversation,
   and that layer is what this book teaches you to build.
4. Feeding §1.1: a space consolidating at this speed is definitionally
   bleeding-edge — and the bleeding edge is where frameworks cannot take you.

*(Publication note: deal figures are from press summaries; verify against
primary sources before print. The section will date — embrace it: "as of
this writing" is honest, and the numbers will only have grown.)*

## 1.1 Frameworks vs. bare metal — the chapter's real lesson

Not a disclaimer; a full section. The student must leave Chapter 1 able to
*defend* the bare-metal ruling, because everything else in the book rests on it.

**The two-sided decision rule, stated honestly:**

- If you just want to build an agent, **use a framework**. That is the right
  call for most agents, and the author has shipped several that way. A simple
  agent → use a framework; you get storage, retries, tool plumbing, and model
  discovery for free.
- An advanced AI **coding** agent is on the bleeding edge, or it is not
  advanced. A framework encodes what its authors anticipated you would need;
  the bleeding edge is precisely what nobody anticipated yet.

**The mechanism, not just the slogan.** Frameworks are generous with *storage*
and *discovery*, but they **hard-code delivery** — what goes into the request
payload, in what order, at what position. And delivery is where the leverage
lives. Three concrete capabilities, all real at the raw API surface today,
that a framework either cannot express or buries:

1. **Mid-turn steering (hints).** The Messages API accepts a `tool_result`
   and a user `text` block *in the same message* — which is how a human
   redirects an agent between tool calls without breaking the tool chain.
   Frontier products validate the idea (real-time steering in their UIs) but
   expose it as a product toggle, not an API their frameworks let you reach.
2. **Ephemeral context placement.** Volatile data (time, screen state, live
   status) must go *last* in the payload, exactly one copy, never in history —
   because delivery *position* is the feature: one wandering timestamp in the
   wrong place destroys prefix caching (a real production case went 0%→98%
   cache hit rate by moving it). Frameworks decide placement for you.
3. **Cache breakpoint control.** You know which suffix of your context is
   volatile; the provider does not, and neither does a framework that
   assembles requests on your behalf. Owning the request bytes is owning
   your cache economics.

**Conclusion:** this class starts with the Anthropic API and an HTTP client.
No SDK, no framework, ever. Every request byte in this book is one you put
there.

(Chapter 1 only *argues* this; the capabilities themselves come much later.
The naive code below doesn't need them — which is exactly why frameworks feel
fine on day one.)

## 1.2 Anatomy of a Messages API request

- Endpoint, `x-api-key`, `anthropic-version`.
- `model`, `max_tokens` (required — and why the API refuses to guess).
- `system` — a string outside the messages array (for now, one fixed line).
- `messages` — the array of `{role, content}`; roles alternate
  `user`/`assistant`.
- Response: `content`, `stop_reason`, `usage`.

## 1.3 The obvious data structure

```go
type Message struct {
    Role    string // "user" or "assistant"
    Content string
}
type Conversation []Message
```

Straight from the docs. Every framework on earth is a wrapper around this
shape. It fits in your head, it serializes trivially, and it will carry this
chapter comfortably.

## 1.4 The loop

Append the user's message. POST the **entire** conversation. Parse the reply.
Append it. Repeat.

The one deep fact of the chapter: **the API is stateless.** The provider
retains nothing between calls; every request replays the whole history. The
conversation lives in *your* process or it lives nowhere.

## 1.5 Usage is money

Read `usage` on every response. Keep cumulative input/output token counts
from the very first request. Watch input tokens grow every round as the
resent history lengthens. (No caching, no remedies — just the habit and the
observed cost curve.)

## 1.6 Chat with it

The payoff, and the point of the chapter: an interactive mode — plain
terminal REPL, live via the course proxy (§1.7) — where the student *talks
to a chatbot they built from raw HTTP*. Encourage playing: give it a personality via the
system line, ask it about the code that created it, show a friend. This is
deliberate: confidence and a little pride first. (Ungraded. The bond formed
here is what Chapter 2's opening is aimed at.)

## 1.7 API access — the toll booth (and the course proxy)

The awkward truth, taught straight because it is part of the landscape:

- **You cannot cheaply get a key that goes fast enough.** Writing code with
  an agent needs sustained token throughput, and the big providers gate that
  behind expensive tiers — Anthropic (last checked) wants ~$400 and a couple
  of weeks of account aging before you can burn tokens at coding speed.
- **Why: the economics of the tool layer.** Every major model advance is
  followed within months by cheap distilled competitors, so raw model access
  is a melting asset. The providers' response is to own the high-value tools
  on top — Claude Code, Claude Cowork, Codex — and to *discount tokens
  consumed through their products* relative to the same tokens via API key.
  The message is explicit: use our agent; don't build your own. This book
  exists to ignore that message. (Same strategy as §1.0's acquisitions —
  buying the execution layer — pointed downmarket at individual developers.)
- **The course's answer: a proxy — optional.** Students fund a modest amount
  (Stripe) on the course site, and their programs hit the course URL, which
  proxies to Anthropic / Google / OpenAI with metered, per-student budgets.
  No provider account, no tier wall, no waiting. **If you already have an
  API key, you don't need us**: the entire course runs identically pointed
  straight at the provider. The proxy exists to remove the toll booth, not
  to become one.
- **The honest cost warning (state it early and plainly):** the course
  author takes no profit on proxied tokens — but tokens cost what they cost,
  direct or proxied. Expect the course's toy agent to burn **$1,000, maybe
  more**, by the end. Building a *real* AI coding agent — the thing this
  book prepares you to do — will run you on the order of **$10K in tokens**.
  That is the real tuition, it goes to the providers, and no route around it
  exists. Budget accordingly before starting.
- **The happy accident:** the Chapter 1 program already targets
  `ANTHROPIC_BASE_URL` because the auto-grader's fake server needs it. The
  proxy is the same seam: fake server for grading, course proxy for live
  chat. One env var; the student's code never knows the difference.

**Infrastructure build item (not chapter content):** rebuild the AI safety
proxy (previous implementation deleted) + Stripe metering + per-student keys
and budgets. Scope separately.

## Exercise (auto-graded)

**Contract:** the student ships a Go program speaking JSON-lines on stdio:
grader writes `{"user": "..."}`, program replies `{"assistant": "..."}`, N
rounds, then EOF → program prints `{"usage": {"input": i, "output": o}}` and
exits 0. **Go required** — the book's code is Go, and later chapters build on
this program.

**Grading rig — a fake Anthropic server.** The grader sets
`ANTHROPIC_BASE_URL` to a local fake that validates every request and returns
scripted responses. No API key, no cost, fully deterministic. What the fake
checks:

1. **Wire correctness** — headers present, `max_tokens` set, roles strictly
   alternating, valid JSON.
2. **Cross-round memory (structural)** — the fake plants a fact in its
   round-1 response and verifies the round-4 *request* still contains the
   full history including that fact. A program that makes a fresh
   single-message call per question fails automatically. This is the proof
   that a conversation exists.
3. **Growth** — request N+1 strictly contains request N's messages as a
   prefix (append-only behavior).
4. **Usage accounting** — final report present, nonzero, consistent with the
   fake's returned usage fields.

**Two modes, one binary:** grader mode (the stdio contract above) and
interactive mode (§1.6's REPL, live via proxy or the student's own API
key). Same conversation code underneath.

**Optional live smoke test:** same program, live via the course proxy, three
rounds. Not
graded; exists so students see a real model answer.

**Out of scope, by design:** tools, hints, thinking, streaming, images,
system-prompt assembly, multiple providers.

---

## Open questions (for Bill)

1. ~~Telegraph or ambush?~~ **RESOLVED: ambush.** Chapter 1 ends clean;
   Chapter 2 opens cold with the demolition.
2. ~~Student language~~ **RESOLVED: Go, required.**
3. ~~Retry in Chapter 1?~~ **RESOLVED: cut, minimalism.** The naive program
   may die on a 429; failure handling belongs to the real build.
