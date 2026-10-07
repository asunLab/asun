#!/usr/bin/env python3
"""Check conformance/cases.json against GRAMMAR.abnf with a stock RFC 5234/7405 parser.

Every `kind: ok` case must be derivable from <asun-document>. A `kind: error`
case must either be rejected by the grammar or carry an `errorHint` naming a
semantic rule (S1-S11), since those errors are outside what ABNF can express.

    pip install abnf
    python3 conformance/runners/abnf/check.py            # check cases.json
    python3 conformance/runners/abnf/check.py 'INPUT'..  # classify inputs
"""
import json
import pathlib
import sys

from abnf import ParseError, Rule

ROOT = pathlib.Path(__file__).resolve().parents[2]


class Asun(Rule):
    pass


Asun.from_file(ROOT / "GRAMMAR.abnf")
DOC = Asun("asun-document")

# errorHint prefixes that denote semantic (non-grammar) errors.
SEMANTIC = ("semantic.", "parse.field_count", "type.")


def accepts(text: str) -> bool:
    try:
        DOC.parse_all(text)
        return True
    except ParseError:
        return False


def main() -> int:
    if len(sys.argv) > 1:
        for text in sys.argv[1:]:
            print(f"{'ACCEPT' if accepts(text) else 'REJECT'}  {text!r}")
        return 0
    cases = json.loads((ROOT / "cases.json").read_text(encoding="utf-8"))["cases"]
    bad = []
    for c in cases:
        ok = accepts(c["input"])
        if c["kind"] == "ok" and not ok:
            bad.append(("ok case rejected by grammar", c))
        elif c["kind"] == "error" and ok and not c.get("errorHint", "").startswith(SEMANTIC):
            bad.append(("error case accepted by grammar without a semantic errorHint", c))
    for why, c in bad:
        print(f"FAIL {c['id']}: {why}\n     input={c['input']!r} hint={c.get('errorHint')}")
    print(f"{len(cases) - len(bad)}/{len(cases)} cases consistent with GRAMMAR.abnf")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
