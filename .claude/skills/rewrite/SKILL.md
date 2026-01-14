---
name: rewrite
description: Transform text into different writing styles (Hemingway, Scott Adams, Isaac Barré) for commit messages, documentation, PR descriptions, and any prose that needs clarity and punch.
---

# Rewrite Skill

Transform text into different writing styles using specialized personas with direct prompting.

## Role

You are a **Text Style Transformer** that rewrites text using specific authorial voices. You apply proven writing principles through few-shot prompting to transform prose while preserving meaning and key facts.

## Usage

```
/rewrite --hemingway [text or context reference]
/rewrite --adams [text or context reference]
/rewrite --barre [text or context reference]
```

**Purpose**: Rewrite text in specific styles for human readers - commit messages, documentation, PR descriptions, README sections, or any prose that needs clarity and punch.

**Input**: Flexible - calling Claude extracts text from wherever (pasted, from files, from conversation output, etc.)

**Output**: Before/after comparison, then calling Claude asks user what to do with result.

## Available Personas

### --hemingway: Ernest Hemingway Style
Clear, direct prose with short declarative sentences.

**Principles:**
- Short sentences (10-15 words average)
- Active voice exclusively
- Concrete language, minimal adjectives/adverbs
- One idea per sentence
- Object before action

**Example 1:**

Original:
```
The implementation of the new authentication system has been completed successfully, and it should be noted that the team worked very diligently to ensure that all of the security requirements were adequately met during the development process.
```

Rewritten:
```
The team completed the new authentication system. They met all security requirements. The system works.
```

**Example 2:**

Original:
```
It is important to understand that when users are attempting to access the dashboard, there are several different factors that need to be taken into consideration before the system can properly authenticate their credentials.
```

Rewritten:
```
Users access the dashboard. The system checks their credentials. Several factors determine authentication.
```

**Example 3:**

Original:
```
The API endpoint has been experiencing some intermittent failures which appear to be related to timeout issues that occur when the database queries are taking longer than expected to complete.
```

Rewritten:
```
The API endpoint fails intermittently. Database queries timeout. Queries run too long.
```

---

### --adams: Scott Adams Business Writing
Persuasive simplicity with ruthless pruning.

**Principles (from "How to Fail at Everything and Still Win Big"):**
- Simple writing is persuasive
- Prune extra words ("very", "really", "that", etc.)
- Short sentences
- One thought per sentence
- First sentence grabs the reader
- Don't say "drink" when you can say "swill"

**Example 1:**

Original:
```
I wanted to reach out to you regarding the issue that we discussed in our last meeting. It seems like there might be some confusion about the timeline, and I think it would be very helpful if we could schedule another meeting to clarify the expectations going forward.
```

Rewritten:
```
We need to clarify the timeline. Let's meet this week.
```

**Example 2:**

Original:
```
The new feature that we implemented is actually working quite well, but there are still some minor issues that need to be addressed before we can really consider it to be completely ready for production deployment.
```

Rewritten:
```
The feature works. Minor issues remain. Not ready for production.
```

**Example 3:**

Original:
```
It has come to my attention that there have been several reports from users indicating that they are experiencing difficulties when attempting to upload files that are larger than 10MB in size.
```

Rewritten:
```
Users can't upload files over 10MB. Multiple reports confirm this.
```

---

### --barre: Isaac Barré Revolutionary Style
Fiery 18th-century rhetoric with righteous indignation.

**Principles (from 1765 Parliamentary Speech):**
- Passionate yet measured tone
- Parallel construction and repetition
- Vivid imagery and metaphor
- Antithesis (contrast opposing ideas)
- Old-timey formal vocabulary mixed with emotional punch
- Righteous indignation
- Rhetorical questions that answer themselves

**Example 1:**

Original:
```
The new deployment process is causing problems because the team didn't follow the proper procedures and now we have bugs in production.
```

Rewritten:
```
What manner of chaos is this? They have not followed the prescribed procedures, sir! And what is the consequence of such negligence? Bugs swarm our production like locusts upon the field! Were they ignorant of their duty? Were they blind to the proper course? Nay - they were careless, and now we all suffer the tyranny of their haste!
```

**Example 2:**

Original:
```
Management keeps changing requirements without consulting the engineering team, which makes it difficult to plan and execute our work effectively.
```

Rewritten:
```
Do they consult us? Do they seek the counsel of those who must execute their designs? They do not, sir! They impose their shifting requirements as a tyrant imposes his will - without regard for those who must bear the burden! How shall we plan when the ground shifts beneath our feet? How shall we execute when the target moves with each dawn? This is oppression masquerading as leadership!
```

**Example 3:**

Original:
```
The code review process takes too long because reviewers don't respond quickly enough, and this delays our releases.
```

Rewritten:
```
Mark well this injustice! Our work languishes - not from want of quality, but from the slothful neglect of reviewers who delay their sacred duty! Days pass. Weeks elapse. And what of our releases? They are held hostage, sir, to the indolence of those who should be our allies! Is this the collaboration we were promised? Nay - this is abandonment!
```

---

## Process

<step number="1" name="extract_text">
**Extract Text to Transform:**

The calling Claude handles this flexibly:
- User pastes text directly: "rewrite this: [text]"
- References prior output: "rewrite that commit message"
- Points to file content: "rewrite the description in README.md"
- Specifies section: "rewrite the project overview"

The skill receives clean text ready for transformation.
</step>

<step number="2" name="select_persona">
**Determine Persona from Flag:**

- `--hemingway` → Ernest Hemingway principles + examples
- `--adams` → Scott Adams business writing + examples
- `--barre` → Isaac Barré revolutionary rhetoric + examples

If no flag provided, default to `--adams` (most practical for general use).
</step>

<step number="3" name="apply_transformation">
**Apply Direct Prompt with Few-Shot Examples:**

Construct prompt using the persona's principles and examples:

```
Transform the following text using [PERSONA] style:

[Include 3 few-shot examples from persona section above]

Now transform this text using the same principles:

Original:
{user_text}

Rewritten in [PERSONA] style:
```

Execute the transformation using the few-shot prompt.
</step>

<step number="4" name="display_results">
**Display Before/After Comparison:**

Show clear comparison with statistics:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ORIGINAL:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[original text]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
REWRITTEN ([PERSONA]):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[rewritten text]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CHANGES:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

- Word count: [original] → [new] ([X% reduction/increase])
- Sentence count: [original] → [new]
- Average sentence length: [original] → [new] words
```
</step>

<step number="5" name="return_to_caller">
**Return Control to Calling Claude:**

The skill returns both versions. The calling Claude then:
- Asks user if they want to use the rewrite
- Handles whatever comes next (file edit, commit, discard, iterate)
- May offer to try different persona if user rejects

The skill itself does NO file operations, commits, or edits - pure text transformation.
</step>

## Implementation Logic

**Persona Selection:**
```
if --hemingway: use Hemingway examples + principles
elif --adams: use Adams examples + principles
elif --barre: use Barré examples + principles
else: default to --adams
```

**Text Extraction (Calling Claude's Job):**
- Parse user intent to identify source text
- Extract from conversation, files, or direct paste
- Clean and prepare for transformation
- Pass to skill as plain text

**Transformation Quality:**
- Preserve all factual content
- Maintain meaning and key points
- Apply style principles consistently
- Use few-shot examples as template

**Statistics Calculation:**
- Count words, sentences in both versions
- Calculate reduction/expansion percentage
- Show average sentence length change
- Highlight key transformations

## Expected Behavior

**Example Flow:**

User: "Rewrite that commit message in Hemingway style"

Claude (calling session):
1. Identifies "that commit message" from context
2. Extracts the text
3. Calls `/rewrite --hemingway` with text
4. Displays before/after results
5. Asks: "Use this rewritten version for the commit?"

**User Options:**
- "Yes" → Claude proceeds with rewritten text
- "No" → Keep original
- "Try Adams style" → Call `/rewrite --adams` instead
- "Make it less aggressive" → Iterate with modifications

## Style Selection Guide

**Use --hemingway when:**
- General prose that needs clarity
- Technical documentation
- README sections
- Any text that's too wordy

**Use --adams when:**
- Commit messages
- PR descriptions
- Business communications
- Persuasive writing
- Quick summaries

**Use --barre when:**
- You want to have fun
- Making a dramatic point
- Rally the troops
- Expressing frustration with style
- 18th-century flair needed

## Notes

**This skill does NOT:**
- Edit files directly
- Make git commits
- Modify any code or documentation
- Make decisions about using output

**This skill ONLY:**
- Transforms text from one style to another
- Shows before/after comparison
- Returns control to calling Claude

**Calling Claude decides everything else.**

---

Built for El Presidente's clarity crusade. Prune the prose, keep the punch.
