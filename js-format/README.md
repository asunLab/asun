# ason-format

Syntax highlighter for **ASON** (Array-Schema Object Notation).  
Zero dependencies · Works in browsers, Node.js, Deno, Bun · Framework-agnostic.

---

## Installation

```bash
npm install ason-format
```

Or grab the pre-built files directly from `dist/`:

| File | Format | Use case |
|---|---|---|
| `ason-format.js` | ESM | Bundlers (Vite, webpack, Rollup…) |
| `ason-format.cjs` | CJS | Node.js `require()` |
| `ason-format.min.js` | IIFE `AsonFormat.*` | `<script>` tag / CDN |
| `ason-format.css` | CSS | Styles for all four built-in themes |

---

## Quick start

### Via `<script>` tag (no build step)

```html
<link rel="stylesheet" href="ason-format.css">

<pre><code id="output"></code></pre>

<script src="ason-format.min.js"></script>
<script>
  const src = `{name:str, age:int}:(Alice, 30)`;
  document.getElementById('output').innerHTML = AsonFormat.highlight(src);
</script>
```

### ESM / bundler

```js
import { highlight } from 'ason-format';
import 'ason-format/css';          // or import the CSS file from dist/

const src = `{name:str, age:int}:(Alice, 30)`;
document.getElementById('output').innerHTML = highlight(src);
```

### Node.js (CJS)

```js
const { highlight } = require('ason-format');
const html = highlight(`{id:int, name:str}:(1, "Alice")`);
console.log(html); // <code class="ason-highlight">…</code>
```

---

## API

### `highlight(src, options?)`

Converts ASON source text to an HTML string with `<span class="ason-*">` tokens.

```ts
function highlight(src: string, options?: HighlightOptions): string;

interface HighlightOptions {
  /** Wrapper HTML tag. Default: 'code' */
  tag?: string;
  /** CSS class(es) on the wrapper. Default: 'ason-highlight' */
  class?: string;
}
```

**Example:**

```js
// Dark theme (default)
highlight(src);

// Light theme
highlight(src, { class: 'ason-highlight ason-light' });

// Wrap in <pre> instead of <code>
highlight(src, { tag: 'pre' });
```

---

### `tokenize(src)`

Returns a flat array of typed tokens — useful for building custom renderers or
editor integrations.

```ts
function tokenize(src: string): Token[];

interface Token {
  kind: TokenKind;
  text: string;
}
```

**Example:**

```js
import { tokenize } from 'ason-format';

const tokens = tokenize(`{name:str}:(Alice)`);
console.log(tokens);
// [
//   { kind: 'schema-open',  text: '{' },
//   { kind: 'field',        text: 'name' },
//   { kind: 'colon',        text: ':' },
//   { kind: 'type',         text: 'str' },
//   { kind: 'schema-close', text: '}' },
//   { kind: 'colon',        text: ':' },
//   { kind: 'tuple-open',   text: '(' },
//   { kind: 'value',        text: 'Alice' },
//   { kind: 'tuple-close',  text: ')' },
// ]
```

---

## Token kinds

Each token receives a CSS class `ason-<kind>`:

| Kind | CSS class | Description |
|---|---|---|
| `field` | `ason-field` | Field name inside a schema |
| `type` | `ason-type` | Type annotation (`int` `str` `float` `bool` …) |
| `map` | `ason-map` | `map` keyword |
| `string` | `ason-string` | Quoted string `"…"` |
| `number` | `ason-number` | Integer, float, or date (`2025-06-24`) |
| `bool` | `ason-bool` | `true` / `false` |
| `value` | `ason-value` | Unquoted plain data value |
| `comment` | `ason-comment` | Block comment `/* … */` |
| `schema-open` | `ason-schema-open` | `{` |
| `schema-close` | `ason-schema-close` | `}` |
| `tuple-open` | `ason-tuple-open` | `(` |
| `tuple-close` | `ason-tuple-close` | `)` |
| `array-open` | `ason-array-open` | `[` |
| `array-close` | `ason-array-close` | `]` |
| `colon` | `ason-colon` | `:` |
| `comma` | `ason-comma` | `,` |

---

## Themes

Include `ason-format.css` once, then add the appropriate class to the wrapper:

| Class | Theme |
|---|---|
| `ason-highlight` | Dark (default — One Dark-inspired) |
| `ason-highlight ason-light` | Light (One Light) |
| `ason-highlight ason-github` | GitHub |
| `ason-highlight ason-tokyo` | Tokyo Night |

```html
<!-- Dark (default) -->
<code class="ason-highlight">…</code>

<!-- Light -->
<code class="ason-highlight ason-light">…</code>

<!-- GitHub -->
<code class="ason-highlight ason-github">…</code>

<!-- Tokyo Night -->
<code class="ason-highlight ason-tokyo">…</code>
```

---

## Custom theme via CSS variables

Override any color without touching the source:

```css
.ason-highlight {
  --ason-field:   hotpink;
  --ason-type:    #00bcd4;
  --ason-string:  #aed581;
  --ason-number:  #ffb74d;
  --ason-bool:    #ce93d8;
  --ason-value:   #e0e0e0;
  --ason-comment: #666;
  --ason-punct:   #90caf9;
  --ason-punct-dim: rgba(144, 202, 249, 0.5);
  background: #1e1e2e;
  color: #cdd6f4;
}
```

---

## Framework examples

### React

```tsx
import { highlight } from 'ason-format';
import 'ason-format/css';

function AsonBlock({ src }: { src: string }) {
  return (
    <pre>
      <code
        className="ason-highlight"
        dangerouslySetInnerHTML={{ __html: highlight(src) }}
      />
    </pre>
  );
}
```

### Vue 3

```vue
<template>
  <pre><code class="ason-highlight" v-html="html" /></pre>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { highlight } from 'ason-format';
import 'ason-format/css';

const props = defineProps<{ src: string }>();
const html  = computed(() => highlight(props.src));
</script>
```

### Svelte

```svelte
<script lang="ts">
  import { highlight } from 'ason-format';
  import 'ason-format/css';

  export let src: string;
  $: html = highlight(src);
</script>

<pre><code class="ason-highlight">{@html html}</code></pre>
```

---

## Building from source

```bash
git clone <repo>
cd tools/ason-format

npm install
npm run build   # produces dist/
```

Open `examples/demo.html` (served via any static file server) to see the
interactive live-preview with all built-in themes and example snippets.

---

## License

MIT
