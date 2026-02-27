// ASON syntax highlighting for VS Code Markdown Preview
(function () {
  "use strict";

  function highlightAsonBlocks() {
    // VS Code markdown preview may use various class patterns:
    // - code.language-ason  (standard markdown-it)
    // - code with parent pre having data-lang="ason" (shiki-based)
    var selectors = [
      'code.language-ason',
      'code[class*="language-ason"]',
      'pre[data-lang="ason"] > code',
      'pre[data-language="ason"] > code'
    ];
    var blocks = document.querySelectorAll(selectors.join(', '));

    // Also scan all <code> elements inside <pre> for unrecognized ason blocks
    if (blocks.length === 0) {
      document.querySelectorAll('pre > code').forEach(function (code) {
        var pre = code.parentElement;
        // Check if the info string or class hints at ason
        var cls = (code.className || '') + ' ' + (pre.className || '');
        if (/\bason\b/i.test(cls)) {
          // Already matched above
        }
      });
    }

    blocks.forEach(function (block) {
      // Skip if already highlighted by us
      if (block.dataset.asonHighlighted) return;
      block.dataset.asonHighlighted = "true";

      // Get plain text (strip any existing shiki spans)
      var text = block.textContent || "";
      if (!text.trim()) return;
      block.innerHTML = tokenize(text);
    });
  }

  function tokenize(src) {
    var result = "";
    var i = 0;
    while (i < src.length) {
      // Comments /* ... */
      if (src[i] === "/" && src[i + 1] === "*") {
        var end = src.indexOf("*/", i + 2);
        var comment = end === -1 ? src.slice(i) : src.slice(i, end + 2);
        result += sp("cmt", esc(comment));
        i += comment.length;
        continue;
      }
      // Strings "..."
      if (src[i] === '"') {
        var j = i + 1;
        while (j < src.length && src[j] !== '"') {
          if (src[j] === "\\") j++;
          j++;
        }
        var str = src.slice(i, j + 1);
        result += sp("str", esc(str));
        i = j + 1;
        continue;
      }
      // Type hints :int :str :bool :float etc
      if (src[i] === ":") {
        var m = src
          .slice(i)
          .match(/^:(int|integer|float|double|str|string|bool|boolean)\b/);
        if (m) {
          result += sp("op", ":") + sp("typ", m[1]);
          i += m[0].length;
          continue;
        }
        result += sp("op", ":");
        i++;
        continue;
      }
      // Numbers (must check before identifiers for negative)
      if (
        /[0-9]/.test(src[i]) ||
        (src[i] === "-" && i + 1 < src.length && /[0-9]/.test(src[i + 1]))
      ) {
        var nm = src.slice(i).match(/^-?[0-9]+(\.[0-9]+)?/);
        if (nm) {
          result += sp("num", nm[0]);
          i += nm[0].length;
          continue;
        }
      }
      // Identifiers / keywords
      if (/[a-zA-Z_]/.test(src[i])) {
        var wm = src.slice(i).match(/^[a-zA-Z_][a-zA-Z0-9_]*/);
        if (wm) {
          var word = wm[0];
          if (word === "true" || word === "false") {
            result += sp("bool", word);
          } else if (word === "map") {
            result += sp("kw", word);
          } else {
            result += sp("var", esc(word));
          }
          i += word.length;
          continue;
        }
      }
      // Brackets
      if ("{[(".indexOf(src[i]) !== -1 || "}])".indexOf(src[i]) !== -1) {
        result += sp("bkt", esc(src[i]));
        i++;
        continue;
      }
      // Comma
      if (src[i] === ",") {
        result += sp("op", ",");
        i++;
        continue;
      }
      // Whitespace and other
      result += esc(src[i]);
      i++;
    }
    return result;
  }

  function sp(cls, content) {
    return '<span class="ason-' + cls + '">' + content + "</span>";
  }

  function esc(s) {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }

  // Inject CSS
  var style = document.createElement("style");
  style.textContent = [
    /* ── Dark theme (default) ── */
    ".ason-cmt  { color: #6a9955; font-style: italic; }",
    ".ason-str  { color: #ce9178; }",
    ".ason-typ  { color: #4ec9b0; }",
    ".ason-num  { color: #b5cea8; }",
    ".ason-bool { color: #569cd6; }",
    ".ason-kw   { color: #c586c0; }",
    ".ason-bkt  { color: #ffd700; }",
    ".ason-op   { color: #d4d4d4; }",
    ".ason-var  { color: #9cdcfe; }",
    /* ── Light theme ── */
    "body.vscode-light .ason-cmt  { color: #008000; }",
    "body.vscode-light .ason-str  { color: #a31515; }",
    "body.vscode-light .ason-typ  { color: #267f99; }",
    "body.vscode-light .ason-num  { color: #098658; }",
    "body.vscode-light .ason-bool { color: #0000ff; }",
    "body.vscode-light .ason-kw   { color: #af00db; }",
    "body.vscode-light .ason-bkt  { color: #0431fa; }",
    "body.vscode-light .ason-op   { color: #000000; }",
    "body.vscode-light .ason-var  { color: #001080; }",
    /* ── High-contrast theme ── */
    "body.vscode-high-contrast .ason-cmt  { color: #7ca668; }",
    "body.vscode-high-contrast .ason-str  { color: #ce9178; }",
    "body.vscode-high-contrast .ason-typ  { color: #4ec9b0; }",
    "body.vscode-high-contrast .ason-num  { color: #b5cea8; }",
    "body.vscode-high-contrast .ason-bool { color: #569cd6; }",
    "body.vscode-high-contrast .ason-kw   { color: #c586c0; }",
    "body.vscode-high-contrast .ason-bkt  { color: #ffd700; }",
    "body.vscode-high-contrast .ason-op   { color: #d4d4d4; }",
    "body.vscode-high-contrast .ason-var  { color: #9cdcfe; }",
  ].join("\n");
  document.head.appendChild(style);

  // Run on load
  highlightAsonBlocks();

  // Watch for dynamic preview updates (markdown preview re-renders on edit)
  var observer = new MutationObserver(function () {
    highlightAsonBlocks();
  });
  observer.observe(document.body, { childList: true, subtree: true });
})();
