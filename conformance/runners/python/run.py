"""Conformance runner for asun-py.

Loads ../../cases.json and decodes each case via asun.decode().
asun-py's `decode` returns Python primitives / dict / list directly,
which is exactly the dynamic-decode shape we want.
"""
import json
import os
import sys
from pathlib import Path

# Build artefact lives in asun-py/ as asun.cpython-*.so
ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "asun-py"))
import asun  # type: ignore

CASES_JSON = Path(__file__).resolve().parents[2] / "cases.json"


def deep_equal(a, b) -> bool:
    if isinstance(a, bool) or isinstance(b, bool):
        return type(a) == type(b) and a == b
    if isinstance(a, (int, float)) and isinstance(b, (int, float)):
        # tolerant numeric compare
        if isinstance(a, int) and isinstance(b, int):
            return a == b
        return abs(float(a) - float(b)) <= 1e-12 or float(a) == float(b)
    if isinstance(a, str) and isinstance(b, str):
        return a == b
    if a is None and b is None:
        return True
    if isinstance(a, list) and isinstance(b, list):
        return len(a) == len(b) and all(deep_equal(x, y) for x, y in zip(a, b))
    if isinstance(a, dict) and isinstance(b, dict):
        return a.keys() == b.keys() and all(deep_equal(a[k], b[k]) for k in a)
    return a == b


def main() -> int:
    manifest = json.loads(CASES_JSON.read_text(encoding="utf-8"))
    cases = manifest["cases"]
    print(f"loaded {len(cases)} cases from {CASES_JSON}")

    total = passed = failed = err_passed = err_failed = skipped = 0
    failures: list[tuple[str, str]] = []

    for c in cases:
        total += 1
        if c.get("schemaDriven"):
            skipped += 1
            continue

        try:
            got = asun.decode(c["input"])
            ok_run = True
        except Exception as e:  # noqa: BLE001
            got = None
            err_msg = f"{type(e).__name__}: {e}"
            ok_run = False

        if c["kind"] == "ok":
            if not ok_run:
                failed += 1
                if len(failures) < 25:
                    failures.append((c["id"], f"expected ok, got error: {err_msg}\n    input: {c['input']!r}"))
                continue
            expected = c["expected"]
            if not deep_equal(got, expected):
                failed += 1
                if len(failures) < 25:
                    failures.append((
                        c["id"],
                        f"value mismatch\n    input:    {c['input']!r}\n    expected: {expected!r}\n    actual:   {got!r}",
                    ))
                continue
            passed += 1
        else:
            if not ok_run:
                err_passed += 1
            else:
                err_failed += 1
                if len(failures) < 25:
                    failures.append((c["id"], f"expected error, got ok: {got!r}\n    input: {c['input']!r}"))

    print()
    print("================ ASUN-PY conformance ================")
    print(f"total                : {total}")
    print(f"untyped ok-cases pass: {passed}")
    print(f"untyped ok-cases fail: {failed}")
    print(f"untyped err-cases pass: {err_passed}")
    print(f"untyped err-cases fail: {err_failed}")
    print(f"skipped (needs typed): {skipped}")
    executed = total - skipped
    pct = (passed + err_passed) / executed * 100.0 if executed else 0.0
    print(f"untyped pass rate    : {passed + err_passed}/{executed} ({pct:.1f}%)")
    print("=====================================================")

    for fid, msg in failures:
        print(f"\n[{fid}]\n    {msg}")

    # ---------- Encode (round-trip) ----------
    enc_path = Path(__file__).resolve().parents[2] / "encode-cases.json"
    enc_failed = 0
    if enc_path.exists():
        em = json.loads(enc_path.read_text(encoding="utf-8"))
        enc_cases = em["cases"]
        print(f"\nloaded {len(enc_cases)} encode cases from {enc_path}")
        enc_passed = 0
        enc_failures: list[tuple[str, str]] = []
        for c in enc_cases:
            val = c["value"]
            try:
                text = asun.encode(val)
            except Exception as e:  # noqa: BLE001
                enc_failed += 1
                if len(enc_failures) < 25:
                    enc_failures.append((c["id"], f"encode failed: {type(e).__name__}: {e}\n    value: {val!r}"))
                continue
            try:
                got = asun.decode(text)
            except Exception as e:  # noqa: BLE001
                enc_failed += 1
                if len(enc_failures) < 25:
                    enc_failures.append((c["id"], f"decode-after-encode failed: {type(e).__name__}: {e}\n    value:   {val!r}\n    encoded: {text!r}"))
                continue
            if not deep_equal(val, got):
                enc_failed += 1
                if len(enc_failures) < 25:
                    enc_failures.append((c["id"], f"round-trip mismatch\n    value:   {val!r}\n    encoded: {text!r}\n    decoded: {got!r}"))
                continue
            enc_passed += 1
        enc_total = enc_passed + enc_failed
        enc_pct = enc_passed / enc_total * 100.0 if enc_total else 0.0
        print()
        print("============ ASUN-PY encode round-trip ==============")
        print(f"total : {enc_total}")
        print(f"pass  : {enc_passed}")
        print(f"fail  : {enc_failed}")
        print(f"rate  : {enc_passed}/{enc_total} ({enc_pct:.1f}%)")
        print("=====================================================")
        for fid, msg in enc_failures:
            print(f"\n[{fid}]\n    {msg}")

    return 0 if (failed == 0 and err_failed == 0 and enc_failed == 0) else 1


if __name__ == "__main__":
    raise SystemExit(main())
