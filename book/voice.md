# Voice

## The aspiration

Writing a compiler is a rite of passage. Thousands of engineers have written one
they never shipped, because doing it once changes how you read every program
afterwards. **Writing an AI coding agent should join that list.**

That sets the register. The tonal models are *Crafting Interpreters*, Crenshaw's
*Let's Build a Compiler*, Nand2Tetris: funny, warm, concrete, rigorous, and
never one word longer than the idea requires. Not a textbook. Not a whitepaper.
Not a blog post with a table of contents.

## Who is reading, and why it constrains us

A professional software engineer, at home, on their own time. No degree, no
certificate, no grade, no employer making them. They came because coding agents
are suddenly worth billions of dollars and they want to know what's actually in
one.

**They can leave at any moment and will never tell us why.** That is the whole
argument for brevity, and it is worth remembering every time a paragraph starts
to sprawl. There is no captive audience to absorb padding. Every paragraph is
buying the next one. A reader who smells filler closes the tab.

## Authority without credentials

We claim no academic standing and should never reach for the posture of it. The
authority here is different and stronger for this audience: **we built the
thing, and we publish our receipts.** We verified the wire formats against live
APIs. We state predictions before running experiments. When our own chapter
claims turned out false — three of them in one week — we said so in print.

This is what buys the right to be informal. You can make jokes if your numbers
are right. Nobody minds a relaxed voice from someone who checked.

Corollary: never inflate. No productivity multipliers, no "40 years of
experience," no vendor claims we haven't verified. One unearned number costs
more trust than ten jokes.

## Criticizing vendors: receipts, never venting

We will say hard things about vendor APIs, and we should — they are true, they
are useful, and a reader who has fought the same silence will feel seen. The
form is what decides whether it reads as evidence or as an axe to grind.

**Every criticism carries a receipt.** A reproducible behavior, a status code, a
measured number, a request you can paste. *"Gemini 3.x returns 400 on a missing
thought signature but 200 on another model's"* is unarguable and taught us
something. *"Gemini's API is terrible"* is a tweet, and it spends exactly the
credibility the receipts bought two pages earlier.

**Frustration is not an argument, because the reader cannot verify it.** Say
what the API did. The reader supplies the feeling, and it lands harder coming
from them.

**Date the claim.** Vendors ship fixes. "As of September 2026" costs four words
and prevents a corrected bug from making the whole book look careless. A stale
complaint is worse than no complaint: it tells the reader our measurements have
an expiry date we did not track.

**Name the API and the model family, never the company.** Write *"the Gemini
API deprecated `generateContent`"*, not *"Google deprecated `generateContent`"*.
Two reasons, both independent of who happens to be writing:

1. A criticism aimed at a technical artifact is something the reader can
   reproduce. A criticism aimed at a company is a claim about corporate intent,
   and no receipt can settle it — so it reads as a grievance no matter how
   carefully it is phrased.
2. API surfaces outlive the org charts that shipped them. A sentence about
   `generateContent` is still true after a reorg; a sentence about a company's
   priorities is stale the moment someone changes jobs.

One exception, and it is a hard one: **verbatim wire identifiers are protocol
facts.** `GoogleSearch` and `GoogleMaps` are literal tool names in the
discovery document. Renaming them to fit this rule would make the record false,
and a false record is a worse failure than an inelegant one. Quote identifiers
exactly as they appear on the wire, always.

**Criticize the API, not the company.** The former is a technical claim we
verified. The latter is a fight we cannot win in a book that takes a year to
reach print.

## Jokes are welcome, and they have to work

Humor is allowed and encouraged. It is not decoration — it is load-bearing like
everything else, and it earns its place by making a mechanism *stick*.

**Err toward comedy, not dryness.** (Ruled by Bill, 2026-09-12: *"I'd rather
err on the side of comedy than dryness."*) The operational form of this rule is
**not** "add jokes" — that produces forced whimsy, banned four bullets down. It
is **stop sanding the absurdity out.** The default drafting instinct
neutralizes a funny truth into a neutral one:

> *Neutralized:* Anthropic represents tool results as `user` messages.
>
> *Not neutralized:* the wire format makes you file the tool's testimony under
> the user's name, because the schema will not let anyone else speak.

Same fact. The second is funnier **and more accurate**, because it tells you
the misattribution is forced rather than chosen. That is the test: when the
funnier phrasing is also the more precise one, it is the right phrasing. When
the joke needs the fact bent even slightly, cut the joke — never the fact.

This rule fails safe. Its worst outcome is a flat sentence. The worst outcome
of "be funnier" is a book trying too hard in front of readers who came for
receipts.

- **Best source: the material's own absurdity.** Gemini disagrees with itself
  about whether its own token counts are nested. Anthropic's API makes you file
  a tool result under the *user's* name because the schema demands alternation.
  You don't need to invent a gag; you need to point.
- **A joke that makes a mechanism memorable is worth its words.** A joke that
  decorates a paragraph is padding with a laugh track. Cut it.
- **Never joke inside a correction, a security warning, or a cost figure.**
  Those paragraphs are where a reader is deciding whether to trust us.
- No forced whimsy, no memes with a shelf life, no winking at the reader about
  how quirky this all is.
- **No profanity.** (Ruled 2026-09-12.) Bill swears freely when describing a
  vendor in conversation, and those conversations are where the best material
  comes from — but the book stays clean. The dry version is funnier anyway: the
  polling-loop line in §2.6 began life as "no fucking way" and lands harder as
  "the only exercise in this book with no defensible answer." Translate the
  heat into precision; do not transcribe it.

## Whose "I"

The book is **Bill's, first person singular.** The war stories are his and are
told as his — the `AIClientInterface` that got copy-pasted into 30,000 lines,
StackAgent thrown away, the migration that is still unfinished. When a draft
needs a story the author does not have, the correct move is to leave a marked
gap and ask, never to invent a plausible one.

"We" is the course talking to the reader about what the course does. "You" is
the reader. Both are fine. Present tense, contractions, short sentences.

## Banned moves

- Throat-clearing: "In this chapter we will…", "Before we begin, let's…"
- Recaps of what the reader just read, and previews of what's coming next.
- Section headings that promise and then deliver a paragraph of setup.
- LLM prose crutches: *delve*, *tapestry*, *testament to*, *it's not just X,
  it's Y*, three-item lists where two items are real and the third is filler,
  and the reflexive em-dash aside that adds nothing. Run the
  `manuscript-manager` prose linter before calling a chapter done.
- Restating a rule in three ways because one way felt too blunt. Blunt is good.
- Apologizing for difficulty, or praising the reader for getting this far.

## The test

Delete any sentence. If no claim, number, instruction, or laugh dies with it,
it was never load-bearing — leave it deleted. **Cutting is the default edit.**

The second test, for a chapter as a whole: would an engineer who has written a
toy compiler recognize this as the same kind of fun?
