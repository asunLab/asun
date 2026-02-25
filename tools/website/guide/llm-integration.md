# LLM Integration

ASON was designed from the ground up for use in LLM (Large Language Model) workflows. This page explains why it works better than JSON for structured model output.

## The Problem with JSON + LLMs

When you ask an LLM to return structured data as JSON, several failure modes are common:

1. **Trailing commas** — invalid JSON but easy to write
2. **Mismatched brackets** — especially in deeply nested structures
3. **Forgotten quotes** — `"name": Alice` instead of `"name": "Alice"`
4. **Repeated field verbosity** — the model spends tokens repeating key names thousands of times

Even with function calling / tool use, JSON-based structured output has high failure rates on complex schemas.

## Why ASON Works Better

### 1. Less to get wrong

ASON requires fewer special characters. Most string values are unquoted:

```ason
{name:str, role:str}:
  (Alice, admin),
  (Bob,   viewer)
```

No quotes around `Alice`, `admin`, `viewer`. No commas inside value strings unless escaped with `\,`. Substantially fewer syntax rules to violate.

### 2. Schema as a prompt header

Include the schema once in the system prompt or as a header:

```
Respond with ASON using this schema:
{id:int, name:str, sentiment:str, score:float}

Example:
{id:int, name:str, sentiment:str, score:float}:
  (1, Product A, positive, 0.92),
  (2, Product B, neutral,  0.51)

Now analyze the following reviews:
...
```

The model only needs to output the data rows — the schema is already established. Hallucination of field names disappears.

### 3. Token budget

With 65% fewer tokens than equivalent JSON, you can fit 3× as many data rows in the same context window. This matters enormously for batch analysis, few-shot examples, and RAG context injection.

### Rough token comparison for 100 records

| Format | Tokens (approx) |
|--------|----------------|
| JSON   | ~3 000          |
| ASON   | ~1 050          |
| ASON-BIN (not LLM-usable) | — |

### 4. Streaming-friendly

ASON's row-oriented format means each tuple `(...)` is a self-contained unit. Clients can begin parsing results as they stream in — you don't need to wait for the closing `}` of a JSON array.

## Recommended Prompt Pattern

```
## Output format
Use ASON with the schema below. Output ONLY the ASON block, no explanation.

Schema: {field1:type, field2:type, ...}

## Data
...your input data...
```

For structured extraction with validation, use a library wrapper that:
1. Sends the schema in the prompt
2. Parses the ASON response
3. Retries once on parse error (rare with ASON)

## Example: Sentiment Analysis Pipeline

```python
import openai
from ason import decode_rows
from dataclasses import dataclass

@dataclass
class Review:
    id: int
    sentiment: str
    score: float

SCHEMA = "{id:int, sentiment:str, score:float}"

def analyze(reviews: list[str]) -> list[Review]:
    numbered = "\n".join(f"{i+1}. {r}" for i, r in enumerate(reviews))
    prompt = f"""Analyze sentiment for each review. 
Respond ONLY with ASON using this schema: {SCHEMA}

Reviews:
{numbered}"""

    resp = openai.chat.completions.create(
        model="gpt-4o",
        messages=[{"role": "user", "content": prompt}]
    )
    return decode_rows(resp.choices[0].message.content, Review)
```
