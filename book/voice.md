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

## Jokes are welcome, and they have to work

Humor is allowed and encouraged. It is not decoration — it is load-bearing like
everything else, and it earns its place by making a mechanism *stick*.

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
