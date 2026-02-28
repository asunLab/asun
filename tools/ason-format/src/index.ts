/**
 * ason-format — syntax highlighter for ASON (Array-Schema Object Notation).
 *
 * Zero dependencies. Works in browsers, Node.js, Deno, Bun.
 * Compatible with any JS framework.
 *
 * API:
 *   highlight(src, options?)  → HTML string with <span class="ason-*"> tags
 *   tokenize(src)             → Token[]
 *
 * Each token gets a CSS class:  ason-field  ason-type  ason-string  ason-number
 *   ason-bool  ason-value  ason-comment  ason-schema-open  ason-tuple-open …
 */

// ─────────────────────────────────────────────────────────────────────────────
// Token types
// ─────────────────────────────────────────────────────────────────────────────

export type TokenKind =
  | 'schema-open'   // {
  | 'schema-close'  // }
  | 'tuple-open'    // (
  | 'tuple-close'   // )
  | 'array-open'    // [
  | 'array-close'   // ]
  | 'colon'         // :
  | 'comma'         // ,
  | 'field'         // identifier used as field name inside schema
  | 'type'          // type annotation: int str float bool …
  | 'map'           // map keyword
  | 'string'        // "quoted string"
  | 'number'        // integer, float, date (2025-06-24)
  | 'bool'          // true / false
  | 'value'         // unquoted data value (plain string in data context)
  | 'comment'       // /* … */
  | 'ws'            // whitespace (space / tab)
  | 'nl'            // newline
  | 'error';        // unexpected character

export interface Token {
  kind: TokenKind;
  text: string;
}

// ─────────────────────────────────────────────────────────────────────────────
// Tokeniser
// ─────────────────────────────────────────────────────────────────────────────

const TYPE_HINTS = new Set([
  'int', 'integer', 'uint',
  'float', 'double',
  'str', 'string',
  'bool', 'boolean',
]);

// Characters that may not appear in an unquoted value token
const IS_DELIM = (c: string) =>
  c === ',' || c === '(' || c === ')' || c === '[' || c === ']' ||
  c === '{' || c === '}' || c === '"' || c === '/' ||
  c === ' ' || c === '\t' || c === '\r' || c === '\n';

/** Whether character can appear in an identifier / field name. */
const IS_IDENT = (c: string) => /[a-zA-Z0-9_+\-]/.test(c);

/** Whether character can start an identifier. */
const IS_IDENT_START = (c: string) => /[a-zA-Z_]/.test(c);

/**
 * Tokenise ASON source into a flat array of typed tokens.
 * The tokeniser is context-aware: it distinguishes field names from data values
 * by tracking schema nesting depth.
 */
export function tokenize(src: string): Token[] {
  const tokens: Token[] = [];
  let i = 0;
  const n = src.length;

  // ── Contextual state ────────────────────────────────────────────────────────
  let schemaDepth = 0;    // number of open '{' we are inside
  let expectField = false; // the next significant ident is a field name
  let expectType  = false; // the next significant token is a type annotation
  let mapTypeDepth = 0;   // >0 when inside map[K,V] brackets

  while (i < n) {
    const ch = src[i];

    // ── Block comment /* … */ ────────────────────────────────────────────────
    if (ch === '/' && src[i + 1] === '*') {
      const end = src.indexOf('*/', i + 2);
      const text = end < 0 ? src.slice(i) : src.slice(i, end + 2);
      tokens.push({ kind: 'comment', text });
      i += text.length;
      continue;
    }

    // ── Newlines ─────────────────────────────────────────────────────────────
    if (ch === '\n') { tokens.push({ kind: 'nl', text: '\n' }); i++; continue; }
    if (ch === '\r') {
      const text = src[i + 1] === '\n' ? '\r\n' : '\r';
      tokens.push({ kind: 'nl', text }); i += text.length; continue;
    }

    // ── Whitespace (non-newline) ───────────────────────────────────────────────
    if (ch === ' ' || ch === '\t') {
      let j = i;
      while (j < n && (src[j] === ' ' || src[j] === '\t')) j++;
      tokens.push({ kind: 'ws', text: src.slice(i, j) });
      i = j; continue;
    }

    // ── Quoted string "…" ────────────────────────────────────────────────────
    if (ch === '"') {
      let j = i + 1;
      while (j < n && src[j] !== '"') {
        if (src[j] === '\\') j++; // skip escape
        j++;
      }
      tokens.push({ kind: 'string', text: src.slice(i, j + 1) });
      i = j + 1;
      if (expectType) expectType = false;
      continue;
    }

    // ── Structural chars ──────────────────────────────────────────────────────
    if (ch === '{') {
      schemaDepth++;
      expectField = true;
      expectType  = false;
      tokens.push({ kind: 'schema-open', text: '{' }); i++; continue;
    }
    if (ch === '}') {
      schemaDepth = Math.max(0, schemaDepth - 1);
      expectField = false;
      expectType  = false;
      tokens.push({ kind: 'schema-close', text: '}' }); i++; continue;
    }
    if (ch === '(') { tokens.push({ kind: 'tuple-open',  text: '(' }); i++; continue; }
    if (ch === ')') { tokens.push({ kind: 'tuple-close', text: ')' }); i++; continue; }

    if (ch === '[') {
      if (mapTypeDepth > 0) mapTypeDepth++;
      tokens.push({ kind: 'array-open', text: '[' }); i++; continue;
    }
    if (ch === ']') {
      if (mapTypeDepth > 0) { mapTypeDepth--; }
      tokens.push({ kind: 'array-close', text: ']' }); i++; continue;
    }

    if (ch === ':') {
      if (schemaDepth > 0) { expectType = true; expectField = false; }
      tokens.push({ kind: 'colon', text: ':' }); i++; continue;
    }
    if (ch === ',') {
      if (schemaDepth > 0 && mapTypeDepth === 0) {
        expectField = true; expectType = false;
      }
      tokens.push({ kind: 'comma', text: ',' }); i++; continue;
    }

    // ── Identifier / keyword ───────────────────────────────────────────────────
    if (IS_IDENT_START(ch)) {
      let j = i;
      while (j < n && IS_IDENT(src[j])) j++;
      const word = src.slice(i, j);

      let kind: TokenKind;
      if (schemaDepth > 0 && (expectType || mapTypeDepth > 0)) {
        // Type position inside schema
        if (word === 'map') {
          kind = 'map';
          mapTypeDepth = 1; // next '[' will be part of map[K,V]
          expectType = false;
        } else {
          kind = TYPE_HINTS.has(word) ? 'type' : 'type'; // any word in type pos = type
          expectType = false;
        }
      } else if (schemaDepth > 0 && expectField) {
        kind = 'field';
        expectField = false;
      } else if (schemaDepth === 0) {
        // Data context
        if (word === 'true' || word === 'false') kind = 'bool';
        else kind = 'value';
      } else {
        // Inside schema but no specific expectation (shouldn't normally occur)
        kind = 'field';
      }

      tokens.push({ kind, text: word });
      i = j; continue;
    }

    // ── Number / date / negative number ─────────────────────────────────────
    // Match: digits optionally followed by . digits, or date-like 2025-06-24
    // In schema context (type position), a number could be a literal default.
    if ((ch >= '0' && ch <= '9') ||
        (ch === '-' && i + 1 < n && src[i + 1] >= '0' && src[i + 1] <= '9')) {
      let j = i;
      // Consume leading minus
      if (src[j] === '-') j++;
      // Consume digit run
      while (j < n && src[j] >= '0' && src[j] <= '9') j++;
      // Decimal part
      if (j < n && src[j] === '.' && j + 1 < n && src[j + 1] >= '0' && src[j + 1] <= '9') {
        j++;
        while (j < n && src[j] >= '0' && src[j] <= '9') j++;
      }
      // Date-like continuation: 2025-06-24  (digits - digits - digits)
      while (j < n && (src[j] === '-' || (src[j] >= '0' && src[j] <= '9'))) j++;
      tokens.push({ kind: 'number', text: src.slice(i, j) });
      if (expectType) expectType = false;
      i = j; continue;
    }

    // ── Unquoted plain value (rare: special chars in data) ────────────────────
    {
      let j = i;
      while (j < n && !IS_DELIM(src[j])) j++;
      if (j > i) {
        tokens.push({ kind: 'value', text: src.slice(i, j) });
        i = j; continue;
      }
    }

    // ── Unknown / error ───────────────────────────────────────────────────────
    tokens.push({ kind: 'error', text: ch });
    i++;
  }

  return tokens;
}

// ─────────────────────────────────────────────────────────────────────────────
// HTML highlighter
// ─────────────────────────────────────────────────────────────────────────────

export interface HighlightOptions {
  /**
   * HTML tag used as the wrapper element.
   * @default 'code'
   */
  tag?: string;
  /**
   * CSS class(es) on the wrapper element.
   * @default 'ason-highlight'
   */
  class?: string;
}

/** Escape HTML special chars. */
function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

/**
 * Render ASON source as an HTML string with `<span class="ason-{kind}">` tags.
 *
 * @example
 * ```html
 * <pre id="code"></pre>
 * <script type="module">
 *   import { highlight } from 'ason-format';
 *   document.getElementById('code').innerHTML = highlight(src);
 * </script>
 * ```
 */
export function highlight(src: string, opts: HighlightOptions = {}): string {
  const tag = opts.tag   ?? 'code';
  const cls = opts.class ?? 'ason-highlight';
  const tokens = tokenize(src);

  let html = `<${esc(tag)} class="${esc(cls)}">`;
  for (const tok of tokens) {
    const text = esc(tok.text);
    // Whitespace and newlines are emitted raw (no span needed)
    if (tok.kind === 'ws' || tok.kind === 'nl') {
      html += text;
    } else {
      html += `<span class="ason-${tok.kind}">${text}</span>`;
    }
  }
  html += `</${esc(tag)}>`;
  return html;
}

// ─────────────────────────────────────────────────────────────────────────────
// Default export (convenient for script-tag / CDN usage)
// ─────────────────────────────────────────────────────────────────────────────

const AsonFormat = { tokenize, highlight };
export default AsonFormat;
