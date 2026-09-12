# Voice audit — `chapter-02-outline.md` (Draft 4, 2026-09-12)

Audited against `book/voice.md`. Target is 1,252 lines. Findings are split
between **draft prose** (judged strictly) and **authorial guidance** (exempt);
every finding below is marked where the call was not obvious.

**Summary — 59 findings.** A: 8 throat-clearing · B: 9 non-load-bearing ·
C: 7 crutches · D: 8 restatements · E: 11 receipt/date failures ·
F: 3 decorative jokes + 2 jokes in banned positions + 5 missed absurdities ·
G: 6 voice drifts.
**Measured:** 142 em-dashes / 1,252 lines (one per 8.8 lines); 53 instances of
the `X, not Y` construction; **zero** classic LLM crutch words (no *delve*,
*tapestry*, *testament to*, *seamless*, *robust*, *leverage*) — clean.
**Highest-value recommendation:** the write-once/additive promise is stated in
**eight** places (lines 31, 37, 113–135, 249, 281, 833, 1208, 1213), one of
which is a section titled *"The contract, stated once."* Collapse to two —
§2.0's contract and §2.3's one-line `Additive, always` — and delete §2.9 item 5,
which is a near-verbatim copy of §2.0. That recovers ~55 lines and removes the
chapter's most visible self-contradiction.
**Highest-risk single passage:** line 1000 — the only paragraph in the chapter
that vents about a vendor without a receipt, and it spends the credibility the
16-of-16 measurement at line 493 earned 500 lines earlier.

---

## A. Throat-clearing — 8 findings

**A1 · line 845** — *"the single sharpest demonstration in the chapter. Hold that thought."*
An explicit preview of what's coming, which `voice.md` bans outright; "hold that
thought" is the purest form of announcing instead of making.
**Fix:** cut "Hold that thought." Cut the forward reference too — let §2.6 land it.

**A2 · line 851** — *"**The centerpiece.** Everything before this exists to make this section possible."*
A heading that promises and then delivers setup — the named banned move. The
section is the centerpiece or it isn't; saying so doesn't make it one.
**Fix:** cut both sentences; open on the blockquote at line 853.

**A3 · line 335** — *"Why this matters enough to state before we need it:"*
Announces that a justification is coming instead of giving it; the sentence
after it is strong and self-justifying.
**Fix:** cut the clause; start at "The system prompt is the most abused surface…"

**A4 · line 1016** — *"**And here is how the bet actually came out, which this chapter prints rather than hides.**"*
Announces the result and congratulates the book for printing it, in one
sentence, immediately before a table that speaks for itself.
**Fix:** cut to "Built in the fixed order, verified against live APIs on 2026-09-12:"

**A5 · line 88** — *"The lesson has two halves and readers usually get taught only the first:"*
Announces the structure of a lesson that is two lines long and formatted as a
blockquote; the reader can see both halves.
**Fix:** cut. The blockquote at 90–93 needs no usher.

**A6 · line 81** — *"Then the part most books leave out."*
Announces novelty rather than demonstrating it, and claims a survey of "most
books" the chapter cannot support.
**Fix:** cut the sentence; "The remedy was worse than the disease." is the line.

**A7 · line 171** — *"The distinction the whole book rests on."*
Tells the reader the stakes before giving them anything to weigh. The three
bullets beneath it establish the stakes in twelve lines.
**Fix:** cut, or demote to the section's last line once the distinction exists.

**A8 · line 479** — *"**`Provenance` is the subtle one, and it is where a seam that looks finished turns out not to be.**"*
Pre-labels a point as subtle. Compare line 481, which just says the naive
version is `Vendor string` and is wrong — that one works.
**Fix:** cut; begin at "The naive version of this field is `Vendor string`."

*(Borderline, not counted: line 865 "The demonstration the chapter is built
around" and line 966 "The prediction the chapter makes out loud" are section
scaffolding, closer to guidance than prose. If they survive into the draft they
become A-class problems; flagging now so the drafter knows.)*

---

## B. Not load-bearing — 9 findings

Applying `voice.md`'s test: delete it, and see what dies.

**B1 · line 1036** — *"A book that prints the result either way is the only kind whose predictions were worth reading in the first place."*
Nothing dies. No claim, no number, no instruction, no laugh — it is the book
praising itself for a virtue the table three lines above already demonstrated.
**Fix:** cut.

**B2 · lines 1033–1035** — *"A design that survives its own falsification test with one field's worth of damage is a design worth copying."*
Nothing dies except self-congratulation. The reader decides whether it is worth
copying; telling them the verdict weakens it.
**Fix:** cut. End the paragraph on "…instead of letting you meet it as a 400."

**B3 · line 283** — *"…the write-once discipline in miniature, and it is worth saying so."*
"and it is worth saying so" is five words asserting that the preceding clause
deserved to exist. Nothing dies.
**Fix:** cut the clause.

**B4 · line 132** — *"So: trust it."*
An instruction the reader cannot act on and would not obey if they could; the
argument for trusting the promise is the two sentences before it.
**Fix:** cut.

**B5 · line 1013** — *"That habit outlives every vendor named in this chapter."*
Aphoristic closer. The advice — build against the honest API first so you have a
control — is complete at line 1012.
**Fix:** cut.

**B6 · line 638** — *"Deleting it removes a field and a failure mode at the same time, which is usually the sign of a correct simplification."*
The first clause is load-bearing; the trailing generalization ("which is usually
the sign of…") adds a rule of thumb the chapter never uses again.
**Fix:** cut from "which is usually".

**B7 · lines 616–617** — *"Watch for that pattern: the rule you are about to read will try to reassert itself in disguise, and it will look like a reasonable local fix every time."*
A preview of a rule that appears two lines later, at line 619. The reader does
not need to be warned about a sentence they are about to read.
**Fix:** cut; move any surviving warning to *after* the rule at 619.

**B8 · line 833** — *"…and it is the chapter's answer to 'am I over-engineering?' — no, and here is the receipt."*
The table above it is the receipt. Narrating that a receipt has been produced is
not a receipt.
**Fix:** cut from "and it is the chapter's answer"; keep "This table is the
write-once discipline made visible."

**B9 — unearned superlatives, 4 instances.** `voice.md`: *"never inflate… One
unearned number costs more trust than ten jokes."* Each of these is a
market-wide ranking claim with no measurement:
- line 335 — *"the most abused surface in agent engineering"*
- line 659 — *"the single most valuable compaction there is"*
- line 724 — *"the most costly one-line mistake in agent engineering"*
- line 735 — *"the purest seam bug in the chapter"*
**Fix:** keep at most one (line 724 is the best earned, since the cache-read
price table at 714–719 supports it). Downgrade the other three to the specific
claim — e.g. 659 to "the compaction this chapter's students will reach for
first." Line 735's "purest… in the chapter" is a narrower claim and survives if
the other two go.

---

## C. LLM prose crutches — 7 findings

**Clean bill first:** grep finds **zero** instances of *delve*, *tapestry*,
*testament to*, *seamless(ly)*, *robust*, *leverage*, *crucial*, *pivotal*,
*landscape*, *realm*, *nuanced*. That is unusual and worth keeping.

**C1 · 53 instances of `X, not Y` / "is not X. It is Y" — the chapter's tic.**
`voice.md` bans *"it's not just X, it's Y"*; the outline uses the whole family as
its default sentence shape. Representative cluster:
- line 138 — *"This seam is not mainly for people running three models. It is for people running one."*
- line 146 — *"That is not a hypothetical migration used to motivate a design. It is a live one"*
- line 41 — *"not a grudging concession — it is what makes write-once safe to promise"*
- line 650 — *"Redaction is a family, not a flag"* (heading)
- line 705 — *"Usage is four numbers, not two"* (heading)
- line 640 — *"A redaction is not metadata about content — **it is content**"*
- line 283 — *"That is not an accident of sequencing; it is the write-once discipline in miniature"*
The construction is genuinely good two or three times; at 53 it becomes audible
and the reader starts hearing the rhythm instead of the claim.
**Fix:** target a 60% reduction. Convert the weakest to plain assertion —
line 138 becomes "This seam is for people running one vendor"; line 146 becomes
"This migration is live." Keep the two headings (650, 705) and the strongest
three in §2.4a.

**C2 · 142 em-dashes in 1,252 lines — one every 8.8 lines.**
`voice.md` bans "the reflexive em-dash aside that adds nothing." Many here are
load-bearing appositives, but the density has passed the point where they read
as a mannerism. Worst offenders are stacked pairs, e.g. line 625 and line 1033.
**Fix:** audit for asides that could be commas or full stops; aim for under 80.
Do not touch em-dashes that introduce a measurement or a definition.

**C3 · line 180** — *"Three consumers, three needs: the renderer reads context; the GUI reads the log; the auditor reads the log."*
A three-item list where **items two and three are identical**. It is literally
three consumers and *two* needs — the sentence contradicts its own opening
count. This is the banned "third item is filler" pattern in its clearest form.
**Fix:** either merge ("the renderer reads context; the GUI and the auditor read
the log") or differentiate the auditor's need, which is presumably immutability
rather than access.

**C4 · lines 953–954** — *"redaction, replay, context surgery"*
Three-item list where the third is vague filler — "context surgery" is not
defined anywhere in the chapter and names no mechanism the other two don't.
**Fix:** cut the third item.

**C5 · line 625** — *"a slow leak with a long fuse, and the fuse burns in production"*
Mixed metaphor: leaks do not have fuses. Two images competing in nine words.
**Fix:** pick one. The fuse is the better image because it carries the delay.

**C6 · clothing-metaphor tic, 3 instances** — line 320 *"the Chapter 1 mistake
wearing a struct"*, line 561 *"a default wearing a disguise"*, line 1066 *"the
same bug in different hats."*
Individually fine; three in one chapter makes it a verbal habit rather than an
image.
**Fix:** keep line 320 (it is the sharpest and lands on the chapter's thesis);
rewrite the other two without the costume.

**C7 · lines 1064–1066** — *"the clock, a randomly generated id, Go's deliberately randomized map iteration order, and iteration over a set. The last two are the same bug in different hats"*
A four-item list that the very next sentence admits is a three-item list.
Announcing "four ways" and then collapsing two of them undercuts the count.
**Fix:** say three, and note that the third has two disguises — or keep four and
cut the retraction.

---

## D. Restatement — 8 findings

**D1 · The write-once/additive promise, stated 8 times.** Lines **31–35**
(Draft-4 note, guidance), **37–47** (escape hatch), **113–135** (§2.0 "The
contract, **stated once**"), **249** ("Additive, always"), **281–285** (§2.3
"shape before the capability"), **833–834** (the Ch3/Ch7 table), **1208–1215**
(§2.9 item 5), **1213** (item 5's duplicate sentence).
The section titled "The contract, stated once" is one of eight.
**Fix — keep two:** §2.0 lines 113–135, and line 249's `Additive, always`. Cut
§2.9 item 5 entirely (it is a resolved open question, not reader-facing).
Compress 281–285 to its first sentence.

**D2 · lines 130–132 vs 1213 — near-verbatim duplicate.**
- 130: *"A reader who does not trust this promise will over-engineer defensively, and defensive over-engineering is the failure mode this book argues against everywhere else."*
- 1213: *"A reader who distrusts the promise over-engineers defensively, which is the failure mode the book argues against everywhere else."*
Same sentence, 1,083 lines apart, differing only in tense and contraction.
**Fix:** keep 130 (it is in reader-facing prose); delete 1213 with the rest of
§2.9 item 5.

**D3 · line 705 — the heading contradicts its own section.**
Heading: *"Usage is four numbers, not two — and they are **not disjoint** the same way."*
Line 738 blockquote: *"**Canonical form: the four fields are disjoint and sum to the billable total.**"*
Plus line 593 code comment (*"All four fields are DISJOINT"*), line 1131 checks
table (*"one **disjoint** set"*), and line 1230 (§2.9 item 8). Five statements of
one rule, and the heading states the opposite of the other four.
**Fix:** rewrite the heading to drop the negation — it is trying to say "vendors
disagree about overlap," which is line 726's job. Keep the blockquote at 738 as
the canonical statement; the code comment at 593 is a listing and exempt, but
the §2.9 copy should go.

**D4 · "the order is fixed," stated 4 times.** Lines **974** ("the **order is
fixed**, and it is fixed to make the test honest"), **984** ("**The hardest
vendor goes last on purpose.**"), **1007** ("**This is the second reason the
order is fixed**"), **1095** (§2.8 "Build them in this order… The reason is in
§2.6"), and §2.9 item 1 ("The order is load-bearing and is fixed in §2.6").
Two of these give genuinely different reasons (falsifiability at 984, control
group at 1007) and both are worth keeping. The other three are pointers.
**Fix:** keep 984 and 1007. Reduce 1095 to the bare instruction with a section
reference. Delete the §2.9 restatement.

**D5 · "Provenance is captured at write time, never inferred," stated 4 times.**
Lines **396–397** (code comment — exempt as a listing), **530–532** (the
blockquote rule), **534–537** ("Inference is impossible in principle here"),
and **958–962** (§2.6 bullet, "hand them back to the **exact model**").
**Fix:** keep the blockquote at 530. Lines 534–537 add the *reason* inference
fails and should be compressed to one sentence rather than cut. The §2.6 bullet
is a different rule (carry, don't interpret) and should stop restating capture
timing.

**D6 · "the seam is the point / one implementation is not a seam," stated 3 times.**
- line 14 (Draft-4 note, guidance) — *"a seam with one implementation is not a seam, it is a naming convention"*
- lines 814–816 — *"An interface extracted from one implementation records that implementation's accidents as though they were requirements."*
- lines 1038–1041 — *"Two points always fit a line… The third implementation is what separates an abstraction from a bridge between two specific things."*
**Fix:** keep 1038 — "Two points always fit a line" is the best compression of
the idea in the chapter and it closes §2.6. Keep 814–816 (it is doing different
work: reading the tell off a signature). The Draft-4 note is guidance and exempt.

**D7 · `iota + 1` / invalid zero value, stated twice at length.**
Lines **407–408** (code comment: *"the zero value is INVALID, so a Provenance
that was never populated is detectable instead of silently meaning 'Anthropic'"*)
and **556–561** (prose: *"Otherwise a `Provenance` that nobody populated is
indistinguishable from a genuine Anthropic/messages one"*).
The listing is exempt from voice rules, but the reader still reads the same
explanation twice, 175 lines apart.
**Fix:** shorten the code comment to *"zero value is invalid on purpose — see
below"* and let the prose carry it.

**D8 · The ephemera rule, stated twice.** Line **1130** (checks table:
*"delivered in exactly one request, and absent from every later one"*) and line
**1156–1160** (notes bullet: *"delivered in exactly one request, absent from every
later one"*) — verbatim apart from one "and."
**Fix:** the table row is the contract; cut the repeat from the bullet, which
should keep only its genuinely new content (the reducer decides, not the capture
site).

---

## E. Vendor criticism without a receipt or a date — 11 findings

**E1 · lines 999–1002** — *"Google's API fails in ways that do not announce they are Google's: a request that returns nothing at all, an error that describes a problem you do not have, a silence that is indistinguishable from a bug in your own assembly code."*
**The worst finding in the audit.** No status code, no reproducible request, no
date, no measurement — three impressionistic complaints in a row. This is the
"Gemini's API is terrible" shape `voice.md` names explicitly, and it sits 500
lines after the chapter's best receipt (16 of 16, line 491), spending exactly
the credibility that measurement bought. It is also a three-item list whose
third item is a simile.
**Fix:** replace with one dated, reproducible failure — a status code and the
request that produces it — or cut the paragraph and keep only the operational
advice at 1004–1012, which stands without the venting.

**E2 · lines 495–498** — *"Anthropic's drop is real but **directional**: it reads its own thinking and that of *earlier* models, and silently discards a *newer* model's — while returning 400 for one that has been modified."*
The Google claim beside it is dated ("Measured 2026-09-12") and quantified
("16 of 16"). This one is undated and unquantified, and it is a *stronger* claim:
"silently discards" asserts an unobservable internal behavior. How was the
silence detected? No method given.
**Fix:** add the date and the measurement that established directionality — how
many model pairings, which direction, what was observed — or demote to "we have
seen" and mark it as needing verification alongside §2.6's other pre-print
checks.

**E3 · line 500 table row — *"| Anthropic | model **binding** | quiet — drops what this model cannot read |"***
The table launders E2's undated assertion into a tidy fact sitting next to a
measured one. A reader scanning only the table cannot tell which row was
measured and which was asserted.
**Fix:** date each row, or add a "measured" column. Given they were established
by different methods, a shared date line above the table is not enough.

**E4 · lines 519–523** — *"Gemini **Interactions** attaches signatures to thought steps and built-in tool steps, but never to standard function calls."*
A precise, falsifiable claim about a surface with no date, no status code, and no
statement of how it was checked — and it concerns a surface the chapter says at
line 143 is **not yet available on Vertex AI**, which raises the question of
where it was observed at all.
**Fix:** date it and say whether it was measured or read from documentation. If
documentation, say so — that is still a receipt, just a weaker one.

**E5 · lines 660–665** — *"The common framework approach — Google's ADK does this — is to replace the oldest *portion* of history with an LLM-written summary."*
A named vendor product criticized with no version, no date, no citation, and no
reproducible behavior. It is also the one criticism aimed at something the
author's employer ships (per the authorial note at line 161), which makes the
missing receipt more costly, not less.
**Fix:** add version and date, and link the documentation or the code path.
Alternatively describe the pattern without naming the product — the argument
about position vs category does not need ADK to work.

**E6 · lines 668–671** — *"in real coding sessions, tool results and tool-call arguments together are the clear majority of a conversation's tokens, while carrying almost none of its continuity."*
"Clear majority" is a measurement claim stated without the measurement. This is
the chapter's central argument for category-over-position compaction and it
rests on a number that is not shown.
**Fix:** state the actual percentages and the sample — this appears to be
measurable from the author's own logs, which would make it one of the chapter's
strongest receipts. Currently it is the weakest link in its best argument.

**E7 · lines 729–733** — *"Some report cached tokens as a **subset** of the prompt total. Others report them as **disjoint** additions alongside it."*
Names no vendor, gives no date, cites no field. The chapter acknowledges this at
lines 742–745 ("must be verified against current documentation before print"),
so it is a known gap — but as written the reader is told a trap exists and given
no way to check whether they are in it.
**Fix:** already flagged by the author; this audit confirms it as blocking. Name
each vendor, the field, the convention, and the date.

**E8 · line 906** — *"Anthropic will not accept two consecutive user messages."*
The load-bearing premise of Exhibit B, stated as a bare assertion with no status
code and no date. It is almost certainly true, which is why it will be the easiest
one to ship unverified.
**Fix:** add the status code and error string, and a date.

**E9 · lines 471–476** — *"A context with nowhere to put per-call replay material cannot produce a valid Gemini request after a tool call at all — the API answers 400."*
Has a status code (good) but no date and no model version, in a chapter that
elsewhere distinguishes Gemini 3.x from 2.5 on exactly this behavior (line 495).
**Fix:** add version and date; it is likely covered by the 2026-09-12
measurement, but the reader cannot tell.

**E10 · lines 754–758** — *"At least one vendor bills cache *storage* by duration"*
An unnamed vendor and an undated pricing claim. Naming it costs one word, and
pricing claims decay fastest of all.
**Fix:** name the vendor, date the claim, or drop the specificity and say the
`Usage` struct cannot express duration-based cost — which is the structural point
and needs no vendor at all.

**E11 · lines 141–147** — *"**As of September 2026, Google has deprecated the API this project uses, and its replacement — the Interactions API — is not yet available on Vertex AI…**"*
Correctly dated and the strongest-formed criticism in the chapter — flagged only
because "not yet available on Vertex AI" is a negative existence claim with no
stated verification path, and negatives are the hardest thing for a reader to
check. The authorial note at 158–163 already handles the sourcing sensitivity
well.
**Fix:** minor — cite the deprecation notice and the Vertex availability page so
the reader can re-check after the date expires.

---

## F. Humor audit

### (i) Jokes that are merely decorative — 3

**F1 · line 157 vs line 1033 — the same joke twice.**
*"One company, one model, two incompatible surfaces is an ordinary Tuesday"* (157)
and *"letting you meet it as a 400 on a Tuesday"* (1033). The first is excellent
and load-bearing — it makes the vendor/surface distinction stick. The second
reuses the Tuesday 876 lines later for no mechanical gain and retroactively
makes the first look like a verbal habit.
**Fix:** keep 157; cut "on a Tuesday" from 1033.

**F2 · line 1066** — *"the same bug in different hats"*
Decorates a list; the mechanism (map iteration order and set iteration are both
unordered traversal) is what makes it memorable, and the joke does not carry it.
See also C6 — it is the third clothing metaphor.
**Fix:** cut the gag, state the mechanism.

**F3 · line 561** — *"a default wearing a disguise"*
A closing flourish on a paragraph whose actual point — an unpopulated
`Provenance` must be detectable — was already made twice in the preceding four
lines. Padding with a laugh track.
**Fix:** cut.

### (ii) Jokes inside a correction, security warning, or cost figure — 2 (banned)

**F4 · line 732** — *"producing a cost figure that is **confidently wrong** in opposite directions depending on which model you are talking to."*
"Confidently wrong" is a wink, and it sits inside the chapter's cost-accounting
passage — precisely the paragraph where `voice.md` says the reader is deciding
whether to trust us. The finding is serious (you will bill against a wrong
number for months); the phrasing invites a smile.
**Fix:** cut "confidently." "Wrong in opposite directions depending on the
vendor" is more alarming without it.

**F5 · lines 1010–1012** — *"…will spend the afternoon apologizing to a machine that was wrong."*
A genuinely good joke, but it is inside a *correction* — the passage exists to
stop the student misattributing a vendor bug to their own code. `voice.md`'s
prohibition is categorical, and this is the paragraph where a student who has
lost an afternoon is deciding whether the book understands them.
**Fix:** borderline — the joke is at the vendor's expense and at the reader's
side, not at the reader's. I would keep it and move it *out* of the corrective
sentence into the paragraph's last line, so the correction lands clean first.
Flagging it so the decision is deliberate rather than accidental.

### (iii) Absurdity played straight — 5 opportunities

These are the highest-value items in the audit. In each case the material is
already funny and the prose is analysing it instead of pointing at it.
**No jokes written here, per brief — opportunities only.**

**F6 · lines 726–733 — vendors that disagree with themselves about their own arithmetic.**
`voice.md` names this exact absurdity as a model joke source ("Gemini disagrees
with itself about whether its own token counts are nested"), and the outline
reaches it independently — then writes it up as "The trap," followed by
subtraction rules. **The absurdity:** a company cannot say whether the number it
charges you for includes the other number it charges you for. **What the reader
is missing:** the moment of realising the vendor is not being difficult on
purpose — nobody over there agrees either. Pre-approved by the voice guide and
currently unclaimed. Best single opportunity in the chapter.

**F7 · lines 878–896 — Exhibit A, filing the machine's words under the human's name.**
`voice.md` cites this one too. Line 890 gets close — *"which is false, and is the
tidiest available lie under a format that demands strict user/assistant
alternation"* — but "tidiest available lie" is an analytic verdict, and then the
prose moves on to OpenAI. **The absurdity:** the schema requires you to sign the
tool's output with the user's name; the format demands a small, mandatory forgery
on every single tool call, forever. **What the reader is missing:** the
recognition that they are about to write the function that commits it, and that
this is the normal, documented, recommended way.

**F8 · lines 484–532 — the sealed envelope you carry between models that cannot read each other.**
Thinking signatures are encrypted reasoning you are forbidden to interpret, must
hand back to the exact model that made it, and which one vendor **silently throws
away** while returning 200. The chapter renders this as a two-row table. **The
absurdity:** you are an unpaid courier for a letter you may never open, between
two correspondents, one of whom bins it without telling you and smiles. **What
the reader is missing:** that "carried, never interpreted" — which sounds like
sober engineering discipline — describes a genuinely strange job, and that the
quiet vendor is the one you should fear, which the chapter argues at line 506 but
never *lets* the reader feel.

**F9 · lines 277–285, 1080, 1116 — a whole chapter of elaborate correspondence with nobody.**
The student writes a reducer for tool events that "it cannot yet produce" (279),
runs a `render` command that "**Makes no network call**" (1080), and is graded
against "**fake endpoints for all three vendors**" (1116) for vendors they may
have no account with. **The absurdity:** an hour spent writing three scrupulously
correct letters to three different people, none of which will be posted, marked
by someone impersonating all three recipients. **What the reader is missing:**
the chapter is slightly embarrassed about this and defends it earnestly at
281–285. Owning the strangeness would defuse the objection faster than the
defence does — and it is exactly the kind of thing Crenshaw would have said out
loud.

**F10 · lines 64–96 — the war story, told solemnly.**
The material: an interface named `AIClientInterface` — which says "interface"
twice and was neither — whose methods take `ClaudeMessage`; copy-pasted twice;
30,000 lines; a rewrite that is still unfinished a year later; and a product that
is *"semi-broken in ways its author has not finished cataloguing, and some bugs
have not yet been reported to anyone, including himself."* **The absurdity:** the
last clause is the funniest line in the chapter and it is placed at the end of a
subordinate list, unremarked. The name `AIClientInterface` goes entirely
uncommented on. **What the reader is missing:** the cold open is the place they
decide whether the next hour is going to be fun. It currently reads as a
confession delivered in a flat voice. The self-deprecation is *already there* —
it just needs to land on a beat rather than trail off. (See also G1: the third
person is what is muffling it.)

*Runners-up, not counted:* line 1064's Go map randomisation (a language feature
that deliberately breaks your program at random so you won't depend on it); lines
158–159's `SurfaceInteractions` "in the enum because this was foreseeable, not
because it was foreseen" — **already a good joke, load-bearing, keep it.**

---

## G. Voice inconsistency — 6 findings

**G1 · lines 85–86 — third person for Bill's own war story.**
*"It is also semi-broken in ways **its author** has not finished cataloguing, and
some bugs have not yet been reported to anyone, including **himself**."*
The section heading at line 62 says *"the seam **I** got wrong."* Twenty-three
lines later the same person is "its author" and "himself." `voice.md`: *"The book
is Bill's, first person singular. The war stories are his and are told as his."*
This is the clearest drift in the document, and it is in the cold open — the
first prose the reader meets.
**Fix:** first person throughout: "ways I have not finished cataloguing… not yet
been reported to anyone, including me."

**G2 · line 1145 — the same drift, in the grading rationale.**
*"Parsing is where vendor shape hides, and where **the author's** own seam failed."*
**Fix:** "…and where my own seam failed."

**G3 · The chapter as grammatical subject where Bill should be — 7 instances.**
Lines **94** (*"the chapter charges an hour"*), **477** (*"the chapter bet against
needing it and lost"*), **966** (*"the prediction the chapter makes out loud"*),
**972** (*"the chapter's falsifiable claim about its own design"*), **1016**
(*"this chapter prints rather than hides"*), **993–995** (*"the chapter has done
its job by losing its own bet"*), **702–703** (*"Chapter 2 owes it only a shape"*).
A book that personifies itself is a book avoiding the word "I." Line 477 is the
sharpest case: *someone* bet against needing that field and lost, and it was not
the chapter.
**Fix:** convert the three where a human actually acted — 477, 995, 1016 — to
first person. The other four are idiomatic and can stay; the problem is density,
not any single use.

**G4 · line 64 — guidance, but it is the source of G1.**
*"Open with **the author's own** failure, told plainly…"*
This is **authorial guidance and exempt from voice.md**, flagged only
diagnostically: the outline is written *to* Bill about Bill, in the third person,
and that framing is leaking into the draft prose at 85 and 1145. Whoever drafts
from this outline will inherit the leak unless warned.
**Fix:** no change to the guidance; add a one-line note to the drafter that every
"the author" in this file becomes "I" in the prose.

**G5 · lines 161–166 — the authorial note, and an unresolved "I".**
*"…the public record carries it, and **the book's author works there**."*
**Guidance, exempt** — but it establishes a fact (Bill works at Google) that
bears directly on findings E5 (criticising Google's ADK without a receipt) and
E4/E11 (Google surface claims). If the prose never discloses this, a reader who
learns it later will re-read every Google criticism in a different light.
**Fix:** not a voice bug — a disclosure decision. Recommend deciding explicitly
whether the prose says it in first person. The chapter's own standard ("we
publish our receipts") argues for yes.

**G6 · lines 1097–1098** — *"That number is the chapter's actual lesson, and it is **yours rather than ours**."*
Correct `we`/`you` usage — flagged only because it is immediately adjacent to
line 1095's instruction voice and the shift from imperative ("Build them in this
order") to possessive contrast within three lines is bumpy. Minor.
**Fix:** optional; cut "and it is yours rather than ours" (it is also a B-class
candidate — nothing dies).

*No instances found of "we" standing in for Bill personally, or of "I" being used
for the course. The `we`/`you` discipline is otherwise clean across 1,252 lines.*

---

## Notes on what was deliberately **not** flagged

To avoid wasting the author's time, the following were read as **authorial
guidance** and judged exempt: lines 9–57 (Draft-4 changelog and claim status),
line 64 and similar "Open with…/State this…/Tell the reader to…" instructions,
the authorial note at 161–166, all Go listings and their comments (§2.4a,
§2.6), all tables (turn transitions, vendor comparison, checks and points,
"what exists now for a later chapter"), the parenthetical "(For the coder: …)"
at 742–745, §2.9 in its entirety as a decision record, and the Appendix.

**Two judgement calls I am flagging as uncertain rather than assuming:**
1. **§2.9 item 3, lines 1188–1200** reads as guidance *and* contains a
   reader-facing instruction ("Say that last part in the prose"). Its final
   paragraph — *"A fake is a model of a vendor, and a model is wrong in exactly
   the places you did not think to model"* — is draft prose in a guidance
   wrapper and is good. Not flagged, but the drafter should know it is prose.
2. **Lines 277–285 (§2.3, "Tool events without a tool loop")** shift mid-passage
   from description to reader-facing argument. B3 flags only the trailing clause
   at 283; if the whole passage is guidance, ignore B3.
