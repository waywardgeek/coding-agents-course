# Chapter 2 wire-format verification — 2026-09-12

Every claim in §2.6 and §2.4a checked against current vendor documentation and,
where possible, live API probes. Three independent verifications; raw
submissions preserved at
`~/projects/coderhapsody/cr/agents/verify-{anthropic,openai,gemini}/submit.md`.

**Verification date: 2026-09-12.** State this in the chapter text. Wire formats
drift and this chapter is nothing but wire formats.

---

## Summary: what the chapter gets wrong

| # | chapter claim | verdict |
|---|---|---|
| 1 | Anthropic **rejects** two consecutive user messages, forcing the Exhibit B merge | **FALSIFIED** — it *combines* them server-side. The merge is still required, for a different and sharper reason. |
| 2 | Replaying another model's thinking: **Google errors, Anthropic silently drops** | **FALSIFIED in both cells.** |
| 3 | OpenAI has no cache-**write** category | **FALSIFIED** — `cache_write_tokens` exists. |
| 4 | Cache reads cost "roughly an order of magnitude less" | **true as a rule**, with named exceptions in both directions |
| 5 | "At least one vendor bills cache storage by duration" | **CONFIRMED — it is Google.** |
| 6 | Gemini `functionCall` → `finishReason` signals tool use | **FALSIFIED** — it is `STOP`. Parse the parts. |

---

## 1. Alternation — Exhibit B's justification is wrong

Anthropic docs, verbatim: *"Consecutive `user` or `assistant` turns in your
request will be **combined into a single turn**."* There is no rejection. A
reader who tests the chapter's claim will find it merges silently.

**The advice survives; only the reason changes.** The real constraints, both
producing HTTP 400:

- *"Tool result blocks must immediately follow their corresponding tool use
  blocks in the message history. You cannot include any messages between the
  assistant's tool use message and the user's tool result message."*
- *"In the user message containing tool results, the `tool_result` blocks must
  come FIRST in the content array. Any text must come AFTER all tool results."*

So a tool result and the human's next instruction really do belong in one user
message, with the result first. Swap the justification, keep the exhibit.

**Neither OpenAI nor Gemini imposes any alternation requirement.** Gemini was
probed live: two *and three* consecutive `user` turns return 200 and are all
genuinely read (an A=1 / B=2 / "what is A+B?" probe answered "3"), consecutive
`model` turns return 200, a conversation opening with a `model` turn returns
200, and omitting `role` entirely returns 200. The contrast the chapter draws
is therefore real — just re-grounded.

## 2. Thinking replay — the table is wrong in both cells

**Anthropic does silently drop, but DIRECTION decides, not difference.**
Verbatim: *"A model reads its own thinking blocks and those of earlier models,
never those of a newer model… The API drops a block the current model can't
read, without an error and without billing it."* Summarized by Anthropic as
*"switching up keeps the conversation's reasoning and switching down drops
it."* So "a different model ⇒ dropped" is half wrong.

Anthropic *is* loud about a **modified** block: 400 `invalid_request_error`.
And a beta header (`thinking-binding-controls-2026-08-01`) now surfaces drops in
an `input_transformations` array, so the silence is becoming optional.

**Google does NOT error on another model's signature.** A 4×4 live matrix —
signatures harvested from four Gemini models and cross-replayed — returned
**16/16 HTTP 200**, including every cross-model pair. The harness is credible
because the *same* harness produced 400s for the two cases that do fail:

- **corrupted** signature → 400 `"Corrupted thought signature."`
- **missing** signature on a replayed `functionCall` → 400 on **Gemini 3.x
  only** (`gemini-2.5-flash` returns 200 for the identical request). There is a
  matching `finishReason` value, `MISSING_THOUGHT_SIGNATURE`.

Caveat for print: this is HTTP-level acceptance. Whether the backend internally
honours a foreign signature is not externally observable. Say *"accepted
without error"*, not *"honoured"*.

**Recommended replacement framing.** The loud/silent contrast is still
available and is now better grounded, because it is about *integrity* rather
than *authorship*: Google validates the signature and refuses loudly when it is
absent or corrupt; Anthropic validates the binding and discards quietly when
the current model cannot read it. Same lesson — the loud failure is the good
one — without a false table.

## 3. Usage conventions — the load-bearing table

| vendor | convention | canonical `Input` | canonical `Output` |
|---|---|---|---|
| Anthropic | **DISJOINT** | `input_tokens` (already excludes cache) | `output_tokens` |
| OpenAI | **SUBSET** | `prompt_tokens − cached_tokens − cache_write_tokens` | `completion_tokens` |
| Gemini | **SUBSET** (input) / **DISJOINT** (thoughts) | `promptTokenCount − cachedContentTokenCount` | `candidatesTokenCount + thoughtsTokenCount` |

The chapter's central claim — *vendors disagree about whether their own
categories overlap* — is **confirmed, and is worse than stated**: Gemini
disagrees with *itself*, subset on the input side and disjoint on the output
side, in the same object.

**Anthropic, documented formula, verbatim:**
`total_input_tokens = cache_read_input_tokens + cache_creation_input_tokens + input_tokens`
with a worked example: 200,000 read + 0 created + 50 input = 200,050. Reading
`input_tokens` alone does not double-count — it **undercounts by ~4000×** on a
warm cache. Trap: `usage.cache_creation` is an object whose `ephemeral_5m` and
`ephemeral_1h` members **sum to** `cache_creation_input_tokens`; adding both
levels double-counts.

**OpenAI is a subset, verified live with numbers.** An identical request
repeated reported `prompt_tokens` **5616 on both** the miss and the hit, with
`cache_write_tokens` 5613 on the first and `cached_tokens` 5613 on the second.
Under a disjoint convention the second call would have reported 3.
`prompt_tokens_details` is **optional** — older models omit it rather than
zeroing it, so parsers must default to 0.

**Gemini's output-side trap is the biggest single costing error available.**
`totalTokenCount = promptTokenCount + candidatesTokenCount + thoughtsTokenCount`
— verified by doc text and by exact arithmetic on three samples (76+792+1001 =
1869 ✅; 76+757+795 = 1628 ✅). Treating thoughts as part of candidates
undercounted billed output by **56%** in the 2.5-flash sample. Thinking tokens
bill at the output rate.

Also unmodelled: Gemini's `toolUsePromptTokenCount`, a fourth billable term.

## 4. Cache pricing

- **Anthropic:** 5-minute writes 1.25×, 1-hour writes 2×, reads 0.1× — *except*
  Fable 5.1 and Mythos 5.1, where reads are **0.025×**. "An order of magnitude
  cheaper" is right as a rule and understates the newest models 4×.
  **No duration-based storage charge**: *"Cache breakpoints themselves don't add
  any cost."* The write premium is the storage charge.
- **OpenAI:** automatic caching, no explicit markers, and a cache-write count
  that *does* exist. No duration storage billing.
- **Google:** the vendor the chapter's "known gap" is about. Storage is billed
  **per 1M tokens per hour**, a separate published line per model (e.g.
  `gemini-3.5-flash` "$1.00 / 1,000,000 tokens per hour"). Explicit caching
  only; implicit caching has no storage cost. **There is no cache-write token
  count anywhere in `usageMetadata`** — the only size figure is returned once,
  at `cachedContents.create`. So the storage cost genuinely has no token count
  attached to any response, exactly as the chapter says.
- The cached-token discount is **90% on Gemini 2.5+**, not 75% (75% was Gemini
  2.0). Pricing arithmetic confirms cached = exactly 10% of input.
- Date-stamp any absolute price: Google's tables already carry a scheduled
  2027-01-01 increase.

## 5. Shapes confirmed

**System prompt, three placements — confirmed.** Anthropic top-level `system`
(string *or* array of blocks). OpenAI a message in the array (`system`;
`developer` preferred on newer models, `system` still accepted everywhere).
Gemini top-level `systemInstruction`, which **must be a `Content` object** —
a bare string is rejected — and whose `role` is accepted but ignored, even
though `role: "system"` *inside* `contents` is a 400.

**Exhibit A, tool-result authorship — confirmed on all three.**

**Two caveats worth a footnote.** Anthropic's `MessageParam.role` enum now
includes `"system"`, for appending mid-conversation instructions without
invalidating a cached prefix — supported on Fable/Mythos/Opus 4.8/Opus 5,
explicitly *not* on Sonnet 5. So "there is no system role in the Messages API"
is no longer unconditionally true, and the docs are internally inconsistent
about it. *(This bears on Chapter 4's hint carriage; see the parking file.)*
Also: `docs.anthropic.com` now redirects to **platform.claude.com** — cite the
new host in print.

**Gemini specifics that bite a parser:**

- `finishReason` is **`STOP`** when the model returns a `functionCall`. There is
  no tool-call value in the 22-value enum. **Detect tool calls by inspecting
  parts.**
- `thoughtSignature` is a **sibling key on the Part**, beside `functionCall` —
  not inside it. In the Interactions API the same concept is `signature`
  *inside* the step. Two names, one idea.
- `functionCall.id` is **optional and model-dependent**: present on Gemini 3.x,
  absent on 2.5.
- `functionResponse.response` must be a JSON **object**; a bare string, number
  or array is a 400.
- `functionResponse.name` is required — and the context has no field for it
  (see finding G1 in the review).

**The Interactions API is real, GA, and genuinely different** — `input` (a step
list), not `contents`; snake_case, not camelCase; `contents` and `messages` are
both rejected. It is absent from the v1beta discovery document yet answers on
the wire. Its usage vocabulary is separate again (`total_input_tokens`,
`total_thought_tokens`, …) and **disjoint**: 5+1+31 = 37 = `total_tokens`. So
`Surface` genuinely needs two usage mappers per vendor, not one — which is a
strong independent justification for the `Surface` field.

Signature-by-step-type, from the machine-readable OpenAPI spec (decisive):
signatures attach to `ThoughtStep` and every built-in tool step
(GoogleSearch, CodeExecution, UrlContext, GoogleMaps, FileSearch, Processing)
and **never** to `FunctionCallStep`, `FunctionResultStep`, `ModelOutputStep`,
`UserInputStep`, or MCP tool steps. The chapter's claim is confirmed exactly.
`FunctionCallStep` properties are exactly `['arguments','id','name','type']`,
with `id` **required** there — unlike generateContent.

## 6. Model ids — never cite from memory

Enumerated live, not recalled.

- **Anthropic** (`GET /v1/models`, 11 entries): `claude-fable-5-1`,
  `claude-opus-5`, `claude-sonnet-5`, `claude-fable-5`, `claude-opus-4-8`,
  `claude-opus-4-7`, `claude-sonnet-4-6`, `claude-opus-4-6`,
  `claude-opus-4-5-20251101`, `claude-haiku-4-5-20251001`,
  `claude-sonnet-4-5-20250929`. Newer ids are **undated**; only 2025-era ids
  keep the `-YYYYMMDD` suffix.
- **Gemini** (`GET /v1beta/models`, 55 entries) includes `gemini-3.5-flash`,
  `gemini-3.6-flash`, `gemini-3.7-flash`, `gemini-3.8-flash`,
  `gemini-3.1-pro-preview`. Avoid the floating aliases `gemini-flash-latest`,
  `gemini-pro-latest` in print.

Training data was badly stale in every case — it contained none of the
`gemini-3.5/3.6/3.7/3.8` family. This vindicates the standing rule.

## 7. Method notes for whoever re-runs this

- **`crawl_web` truncated Anthropic's prompt-caching page before the "Tracking
  cache performance" section** — the single most important section in the task.
  Trusting the crawler would have produced "usage convention UNVERIFIED" on the
  highest-value claim. Fetching the raw `.md` with `curl` and grepping locally
  is what produced the formula.
- **Machine-readable sources beat prose.** Google's discovery document
  (revision 20260910) and the Interactions OpenAPI spec settled the enum list,
  usage semantics and signature-by-step-type exactly, where HTML guides were
  vague or silent. But the Interactions API is *missing* from discovery and only
  answers on the wire — so no single source was sufficient.
- Verify a live response's envelope shape before recording it: a stale `/tmp`
  filename nearly caused a Gemini model list to be reported as Anthropic's.
