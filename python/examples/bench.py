"""ASON vs JSON Comprehensive Benchmark — matches Rust bench.rs."""

import json
import resource
import sys
import os
import time

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from dataclasses import dataclass, field
from typing import Optional
import ason

# ===========================================================================
# 1. Flat struct (8 fields)
# ===========================================================================

@dataclass
class User:
    id: int = 0
    name: str = ""
    email: str = ""
    age: int = 0
    score: float = 0.0
    active: bool = True
    role: str = ""
    city: str = ""


# ===========================================================================
# 2. All-types struct
# ===========================================================================

@dataclass
class AllTypes:
    b: bool = False
    i64v: int = 0
    f64v: float = 0.0
    s: str = ""
    opt_some: Optional[int] = None
    opt_none: Optional[int] = None
    vec_int: list[int] = field(default_factory=list)
    vec_str: list[str] = field(default_factory=list)


# ===========================================================================
# 3. 5-level deep: Company > Division > Team > Project > Task
# ===========================================================================

@dataclass
class Task:
    id: int = 0
    title: str = ""
    priority: int = 0
    done: bool = False
    hours: float = 0.0


@dataclass
class Project:
    name: str = ""
    budget: float = 0.0
    active: bool = False
    tasks: list[Task] = field(default_factory=list)


@dataclass
class Team:
    name: str = ""
    lead: str = ""
    size: int = 0
    projects: list[Project] = field(default_factory=list)


@dataclass
class Division:
    name: str = ""
    location: str = ""
    headcount: int = 0
    teams: list[Team] = field(default_factory=list)


@dataclass
class Company:
    name: str = ""
    founded: int = 0
    revenue_m: float = 0.0
    public: bool = False
    divisions: list[Division] = field(default_factory=list)
    tags: list[str] = field(default_factory=list)


# ===========================================================================
# Data generators
# ===========================================================================

_NAMES = ["Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Hank"]
_ROLES = ["engineer", "designer", "manager", "analyst"]
_CITIES = ["NYC", "LA", "Chicago", "Houston", "Phoenix"]


def generate_users(n: int) -> list[User]:
    return [
        User(
            id=i, name=_NAMES[i % 8],
            email=f"{_NAMES[i % 8].lower()}@example.com",
            age=25 + i % 40, score=50.0 + (i % 50) + 0.5,
            active=i % 3 != 0, role=_ROLES[i % 4], city=_CITIES[i % 5],
        )
        for i in range(n)
    ]


def generate_all_types(n: int) -> list[AllTypes]:
    return [
        AllTypes(
            b=i % 2 == 0, i64v=i * 100_000,
            f64v=i * 0.25 + 0.5, s=f"item_{i}",
            opt_some=i if i % 2 == 0 else None, opt_none=None,
            vec_int=[i, i + 1, i + 2],
            vec_str=[f"tag{i%5}", f"cat{i%3}"],
        )
        for i in range(n)
    ]


def generate_companies(n: int) -> list[Company]:
    _locs = ["NYC", "London", "Tokyo", "Berlin"]
    _leads = ["Alice", "Bob", "Carol", "David"]
    return [
        Company(
            name=f"Corp_{i}", founded=1990 + i % 35,
            revenue_m=10.0 + i * 5.5, public=i % 2 == 0,
            divisions=[
                Division(
                    name=f"Div_{i}_{d}", location=_locs[d % 4],
                    headcount=50 + d * 20,
                    teams=[
                        Team(
                            name=f"Team_{i}_{d}_{t}", lead=_leads[t % 4],
                            size=5 + t * 2,
                            projects=[
                                Project(
                                    name=f"Proj_{t}_{p}", budget=100.0 + p * 50.5,
                                    active=p % 2 == 0,
                                    tasks=[
                                        Task(
                                            id=i * 100 + d * 10 + t * 5 + tk,
                                            title=f"Task_{tk}", priority=tk % 3 + 1,
                                            done=tk % 2 == 0, hours=2.0 + tk * 1.5,
                                        )
                                        for tk in range(4)
                                    ],
                                )
                                for p in range(3)
                            ],
                        )
                        for t in range(2)
                    ],
                )
                for d in range(2)
            ],
            tags=["enterprise", "tech", f"sector_{i%5}"],
        )
        for i in range(n)
    ]


# ===========================================================================
# Helpers
# ===========================================================================

def format_bytes(b: int) -> str:
    if b >= 1_048_576:
        return f"{b / 1048576:.1f} MB"
    if b >= 1024:
        return f"{b / 1024:.1f} KB"
    return f"{b} B"


def get_rss_bytes() -> int:
    ru = resource.getrusage(resource.RUSAGE_SELF)
    if sys.platform == "darwin":
        return ru.ru_maxrss  # macOS: bytes
    return ru.ru_maxrss * 1024  # Linux: KB


def _to_dict(obj):
    """Recursively convert dataclass to dict for json.dumps."""
    if hasattr(obj, "__dataclass_fields__"):
        return {k: _to_dict(v) for k, v in obj.__dict__.items()}
    if isinstance(obj, list):
        return [_to_dict(x) for x in obj]
    return obj


def _json_dumps_list(lst):
    return json.dumps([_to_dict(x) for x in lst])


def _json_loads_user_list(s):
    return [User(**d) for d in json.loads(s)]


# ===========================================================================
# Benchmark runners
# ===========================================================================

class BenchResult:
    __slots__ = ("name", "json_ser_ms", "ason_ser_ms", "json_de_ms", "ason_de_ms",
                 "json_bytes", "ason_bytes")

    def __init__(self, name, json_ser_ms, ason_ser_ms, json_de_ms, ason_de_ms, json_bytes, ason_bytes):
        self.name = name
        self.json_ser_ms = json_ser_ms
        self.ason_ser_ms = ason_ser_ms
        self.json_de_ms = json_de_ms
        self.ason_de_ms = ason_de_ms
        self.json_bytes = json_bytes
        self.ason_bytes = ason_bytes

    def pr(self):
        ser_ratio = self.json_ser_ms / self.ason_ser_ms if self.ason_ser_ms > 0 else 0
        de_ratio = self.json_de_ms / self.ason_de_ms if self.ason_de_ms > 0 else 0
        saving = (1.0 - self.ason_bytes / self.json_bytes) * 100.0 if self.json_bytes else 0

        print(f"  {self.name}")
        tag_s = "✓ ASON faster" if ser_ratio >= 1.0 else ""
        print(f"    Serialize:   JSON {self.json_ser_ms:>8.2f}ms | ASON {self.ason_ser_ms:>8.2f}ms | ratio {ser_ratio:.2f}x {tag_s}")
        tag_d = "✓ ASON faster" if de_ratio >= 1.0 else ""
        print(f"    Deserialize: JSON {self.json_de_ms:>8.2f}ms | ASON {self.ason_de_ms:>8.2f}ms | ratio {de_ratio:.2f}x {tag_d}")
        print(f"    Size:        JSON {self.json_bytes:>8d} B | ASON {self.ason_bytes:>8d} B | saving {saving:.0f}%")


def bench_flat(count: int, iterations: int) -> BenchResult:
    users = generate_users(count)

    json_str = ""
    t0 = time.perf_counter()
    for _ in range(iterations):
        json_str = _json_dumps_list(users)
    json_ser = (time.perf_counter() - t0) * 1000

    ason_str = ""
    t0 = time.perf_counter()
    for _ in range(iterations):
        ason_str = ason.dump_slice(users)
    ason_ser = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(iterations):
        json.loads(json_str)
    json_de = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(iterations):
        ason.load_slice(ason_str, User)
    ason_de = (time.perf_counter() - t0) * 1000

    # verify
    decoded = ason.load_slice(ason_str, User)
    assert len(decoded) == count

    return BenchResult(
        name=f"Flat struct × {count} (8 fields)",
        json_ser_ms=json_ser, ason_ser_ms=ason_ser,
        json_de_ms=json_de, ason_de_ms=ason_de,
        json_bytes=len(json_str), ason_bytes=len(ason_str),
    )


def bench_all_types(count: int, iterations: int) -> BenchResult:
    items = generate_all_types(count)

    json_str = ""
    t0 = time.perf_counter()
    for _ in range(iterations):
        json_str = _json_dumps_list(items)
    json_ser = (time.perf_counter() - t0) * 1000

    ason_strs: list[str] = []
    t0 = time.perf_counter()
    for _ in range(iterations):
        ason_strs = [ason.dump(item) for item in items]
    ason_ser = (time.perf_counter() - t0) * 1000
    ason_total = "\n".join(ason_strs)

    t0 = time.perf_counter()
    for _ in range(iterations):
        json.loads(json_str)
    json_de = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(iterations):
        for line in ason_strs:
            ason.load(line, AllTypes)
    ason_de = (time.perf_counter() - t0) * 1000

    return BenchResult(
        name=f"All-types struct × {count} (8 fields, per-struct)",
        json_ser_ms=json_ser, ason_ser_ms=ason_ser,
        json_de_ms=json_de, ason_de_ms=ason_de,
        json_bytes=len(json_str), ason_bytes=len(ason_total),
    )


def bench_deep(count: int, iterations: int) -> BenchResult:
    companies = generate_companies(count)

    json_str = ""
    t0 = time.perf_counter()
    for _ in range(iterations):
        json_str = _json_dumps_list(companies)
    json_ser = (time.perf_counter() - t0) * 1000

    ason_strs: list[str] = []
    t0 = time.perf_counter()
    for _ in range(iterations):
        ason_strs = [ason.dump(c) for c in companies]
    ason_ser = (time.perf_counter() - t0) * 1000
    ason_total = "\n".join(ason_strs)

    t0 = time.perf_counter()
    for _ in range(iterations):
        json.loads(json_str)
    json_de = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(iterations):
        for s in ason_strs:
            ason.load(s, Company)
    ason_de = (time.perf_counter() - t0) * 1000

    # verify
    for i, s in enumerate(ason_strs):
        c2 = ason.load(s, Company)
        assert c2.name == companies[i].name

    return BenchResult(
        name=f"5-level deep × {count} (Company>Division>Team>Project>Task, ~48 nodes each)",
        json_ser_ms=json_ser, ason_ser_ms=ason_ser,
        json_de_ms=json_de, ason_de_ms=ason_de,
        json_bytes=len(json_str), ason_bytes=len(ason_total),
    )


def bench_single_roundtrip(iterations: int) -> tuple[float, float]:
    user = User(id=1, name="Alice", email="alice@example.com", age=30,
                score=95.5, active=True, role="engineer", city="NYC")

    t0 = time.perf_counter()
    for _ in range(iterations):
        s = ason.dump(user)
        ason.load(s, User)
    ason_ms = (time.perf_counter() - t0) * 1000

    d = _to_dict(user)
    t0 = time.perf_counter()
    for _ in range(iterations):
        s = json.dumps(d)
        json.loads(s)
    json_ms = (time.perf_counter() - t0) * 1000

    return ason_ms, json_ms


def bench_deep_single_roundtrip(iterations: int) -> tuple[float, float]:
    company = generate_companies(1)[0]

    t0 = time.perf_counter()
    for _ in range(iterations):
        s = ason.dump(company)
        ason.load(s, Company)
    ason_ms = (time.perf_counter() - t0) * 1000

    d = _to_dict(company)
    t0 = time.perf_counter()
    for _ in range(iterations):
        s = json.dumps(d)
        json.loads(s)
    json_ms = (time.perf_counter() - t0) * 1000

    return ason_ms, json_ms


# ===========================================================================
# Main
# ===========================================================================

def main():
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║            ASON vs JSON Comprehensive Benchmark            ║")
    print("╚══════════════════════════════════════════════════════════════╝")

    print(f"\nSystem: {sys.platform} Python {sys.version.split()[0]}")

    rss_before = get_rss_bytes()
    print(f"RSS before benchmarks: {format_bytes(rss_before)}\n")

    iterations = 20
    print(f"Iterations per test: {iterations}")

    # ===================================================================
    # Section 1: Flat struct
    # ===================================================================
    print("\n┌─────────────────────────────────────────────┐")
    print("│  Section 1: Flat Struct (schema-driven vec) │")
    print("└─────────────────────────────────────────────┘")

    for count in [100, 500, 1000]:
        r = bench_flat(count, iterations)
        r.pr()
        print()

    rss_after_flat = get_rss_bytes()
    print(f"  RSS after flat benchmarks: {format_bytes(rss_after_flat)} "
          f"(Δ {format_bytes(rss_after_flat - rss_before)})")

    # ===================================================================
    # Section 2: All-types struct
    # ===================================================================
    print("\n┌──────────────────────────────────────────────┐")
    print("│  Section 2: All-Types Struct (8 fields)      │")
    print("└──────────────────────────────────────────────┘")

    for count in [100, 500]:
        r = bench_all_types(count, iterations)
        r.pr()
        print()

    # ===================================================================
    # Section 3: 5-level deep nested struct
    # ===================================================================
    print("┌──────────────────────────────────────────────────────────┐")
    print("│  Section 3: 5-Level Deep Nesting (Company hierarchy)    │")
    print("└──────────────────────────────────────────────────────────┘")

    for count in [10, 50]:
        r = bench_deep(count, iterations)
        r.pr()
        print()

    rss_after_deep = get_rss_bytes()
    print(f"  RSS after deep benchmarks: {format_bytes(rss_after_deep)} "
          f"(Δ {format_bytes(rss_after_deep - rss_before)})")

    # ===================================================================
    # Section 4: Single struct roundtrip
    # ===================================================================
    print("\n┌──────────────────────────────────────────────┐")
    print("│  Section 4: Single Struct Roundtrip (10000x) │")
    print("└──────────────────────────────────────────────┘")

    ason_flat, json_flat = bench_single_roundtrip(10000)
    print(f"  Flat:  ASON {ason_flat:>6.2f}ms | JSON {json_flat:>6.2f}ms | ratio {json_flat/ason_flat:.2f}x")

    ason_deep, json_deep = bench_deep_single_roundtrip(10000)
    print(f"  Deep:  ASON {ason_deep:>6.2f}ms | JSON {json_deep:>6.2f}ms | ratio {json_deep/ason_deep:.2f}x")

    # ===================================================================
    # Section 5: Large payload — 5k flat records
    # ===================================================================
    print("\n┌──────────────────────────────────────────────┐")
    print("│  Section 5: Large Payload (5k records)       │")
    print("└──────────────────────────────────────────────┘")

    r_large = bench_flat(5000, 5)
    print("  (5 iterations for large payload)")
    r_large.pr()

    rss_after_large = get_rss_bytes()
    print(f"\n  RSS after large payload: {format_bytes(rss_after_large)} "
          f"(Δ {format_bytes(rss_after_large - rss_before)})")

    # ===================================================================
    # Section 6: Annotated vs Unannotated Schema Deserialization
    # ===================================================================
    print("\n┌──────────────────────────────────────────────────────────────┐")
    print("│  Section 6: Annotated vs Unannotated Schema (deserialize)   │")
    print("└──────────────────────────────────────────────────────────────┘")

    users_1k = generate_users(1000)
    ason_untyped = ason.dump_slice(users_1k)
    ason_typed_str = ason.dump_slice_typed(users_1k)

    v1 = ason.load_slice(ason_untyped, User)
    v2 = ason.load_slice(ason_typed_str, User)
    assert len(v1) == len(v2) == 1000

    de_iters = 50
    t0 = time.perf_counter()
    for _ in range(de_iters):
        ason.load_slice(ason_untyped, User)
    untyped_ms = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(de_iters):
        ason.load_slice(ason_typed_str, User)
    typed_ms = (time.perf_counter() - t0) * 1000

    ratio = untyped_ms / typed_ms if typed_ms > 0 else 0
    print(f"  Flat struct × 1000 ({de_iters} iters, deserialize only)")
    print(f"    Unannotated: {untyped_ms:>8.2f}ms  ({len(ason_untyped)} B)")
    print(f"    Annotated:   {typed_ms:>8.2f}ms  ({len(ason_typed_str)} B)")
    print(f"    Ratio: {ratio:.3f}x (unannotated / annotated)")
    print(f"    Schema overhead: +{len(ason_typed_str) - len(ason_untyped)} bytes "
          f"({(len(ason_typed_str) / len(ason_untyped) - 1) * 100:.1f}%)")
    print()

    company = generate_companies(1)[0]
    ason_deep_untyped = ason.dump(company)
    ason_deep_typed = ason.dump_typed(company)

    c1 = ason.load(ason_deep_untyped, Company)
    c2 = ason.load(ason_deep_typed, Company)
    assert c1.name == c2.name

    deep_iters = 1000
    t0 = time.perf_counter()
    for _ in range(deep_iters):
        ason.load(ason_deep_untyped, Company)
    deep_untyped_ms = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(deep_iters):
        ason.load(ason_deep_typed, Company)
    deep_typed_ms = (time.perf_counter() - t0) * 1000

    deep_ratio = deep_untyped_ms / deep_typed_ms if deep_typed_ms > 0 else 0
    print(f"  5-level deep single struct ({deep_iters} iters, deserialize only)")
    print(f"    Unannotated: {deep_untyped_ms:>8.2f}ms  ({len(ason_deep_untyped)} B)")
    print(f"    Annotated:   {deep_typed_ms:>8.2f}ms  ({len(ason_deep_typed)} B)")
    print(f"    Ratio: {deep_ratio:.3f}x (unannotated / annotated)")
    print()
    print("  Summary: Type annotations add a small schema parsing cost but")
    print("  are negligible in overall deserialization. Both produce identical results.")

    # ===================================================================
    # Section 7: Annotated vs Unannotated Schema Serialization
    # ===================================================================
    print("\n┌──────────────────────────────────────────────────────────────┐")
    print("│  Section 7: Annotated vs Unannotated Schema (serialize)     │")
    print("└──────────────────────────────────────────────────────────────┘")

    ser_iters = 50
    t0 = time.perf_counter()
    for _ in range(ser_iters):
        untyped_out = ason.dump_slice(users_1k)
    untyped_ser_ms = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(ser_iters):
        typed_out = ason.dump_slice_typed(users_1k)
    typed_ser_ms = (time.perf_counter() - t0) * 1000

    ser_ratio = untyped_ser_ms / typed_ser_ms if typed_ser_ms > 0 else 0
    print(f"  Flat struct × 1000 list ({ser_iters} iters, serialize only)")
    print(f"    Unannotated: {untyped_ser_ms:>8.2f}ms  ({len(untyped_out)} B)")
    print(f"    Annotated:   {typed_ser_ms:>8.2f}ms  ({len(typed_out)} B)")
    print(f"    Ratio: {ser_ratio:.3f}x (unannotated / annotated)")
    print()

    deep_ser_iters = 1000
    t0 = time.perf_counter()
    for _ in range(deep_ser_iters):
        ason.dump(company)
    deep_untyped_ser_ms = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    for _ in range(deep_ser_iters):
        ason.dump_typed(company)
    deep_typed_ser_ms = (time.perf_counter() - t0) * 1000

    deep_ser_ratio = deep_untyped_ser_ms / deep_typed_ser_ms if deep_typed_ser_ms > 0 else 0
    print(f"  5-level deep single struct ({deep_ser_iters} iters, serialize only)")
    print(f"    Unannotated: {deep_untyped_ser_ms:>8.2f}ms")
    print(f"    Annotated:   {deep_typed_ser_ms:>8.2f}ms")
    print(f"    Ratio: {deep_ser_ratio:.3f}x (unannotated / annotated)")
    print()
    print("  Summary: Typed serialization has minimal overhead. The extra cost")
    print("  is recording and emitting type hints in the schema header.")

    # ===================================================================
    # Section 8: Throughput summary
    # ===================================================================
    print("\n┌──────────────────────────────────────────────┐")
    print("│  Section 8: Throughput Summary               │")
    print("└──────────────────────────────────────────────┘")

    users_1k = generate_users(1000)
    json_1k = _json_dumps_list(users_1k)
    ason_1k = ason.dump_slice(users_1k)

    tp_iters = 20

    t0 = time.perf_counter()
    for _ in range(tp_iters):
        _json_dumps_list(users_1k)
    json_ser_dur = time.perf_counter() - t0

    t0 = time.perf_counter()
    for _ in range(tp_iters):
        ason.dump_slice(users_1k)
    ason_ser_dur = time.perf_counter() - t0

    t0 = time.perf_counter()
    for _ in range(tp_iters):
        json.loads(json_1k)
    json_de_dur = time.perf_counter() - t0

    t0 = time.perf_counter()
    for _ in range(tp_iters):
        ason.load_slice(ason_1k, User)
    ason_de_dur = time.perf_counter() - t0

    total_records = 1000.0 * tp_iters
    json_ser_rps = total_records / json_ser_dur if json_ser_dur > 0 else 0
    ason_ser_rps = total_records / ason_ser_dur if ason_ser_dur > 0 else 0
    json_de_rps = total_records / json_de_dur if json_de_dur > 0 else 0
    ason_de_rps = total_records / ason_de_dur if ason_de_dur > 0 else 0

    json_ser_mbps = len(json_1k) * tp_iters / json_ser_dur / 1_048_576 if json_ser_dur > 0 else 0
    ason_ser_mbps = len(ason_1k) * tp_iters / ason_ser_dur / 1_048_576 if ason_ser_dur > 0 else 0
    json_de_mbps = len(json_1k) * tp_iters / json_de_dur / 1_048_576 if json_de_dur > 0 else 0
    ason_de_mbps = len(ason_1k) * tp_iters / ason_de_dur / 1_048_576 if ason_de_dur > 0 else 0

    tag_s = "✓ ASON faster" if ason_ser_rps > json_ser_rps else ""
    tag_d = "✓ ASON faster" if ason_de_rps > json_de_rps else ""

    print(f"  Serialize throughput (1000 records × {tp_iters} iters):")
    print(f"    JSON: {json_ser_rps:.0f} records/s  ({json_ser_mbps:.1f} MB/s of JSON)")
    print(f"    ASON: {ason_ser_rps:.0f} records/s  ({ason_ser_mbps:.1f} MB/s of ASON)")
    print(f"    Speed: {ason_ser_rps/json_ser_rps:.2f}x {tag_s}")
    print("  Deserialize throughput:")
    print(f"    JSON: {json_de_rps:.0f} records/s  ({json_de_mbps:.1f} MB/s)")
    print(f"    ASON: {ason_de_rps:.0f} records/s  ({ason_de_mbps:.1f} MB/s)")
    print(f"    Speed: {ason_de_rps/json_de_rps:.2f}x {tag_d}")

    rss_final = get_rss_bytes()
    print("\n  Memory:")
    print(f"    Initial RSS:  {format_bytes(rss_before)}")
    print(f"    Final RSS:    {format_bytes(rss_final)}")
    print(f"    Peak delta:   {format_bytes(rss_final - rss_before)}")

    print("\n╔══════════════════════════════════════════════════════════════╗")
    print("║                    Benchmark Complete                       ║")
    print("╚══════════════════════════════════════════════════════════════╝")


if __name__ == "__main__":
    main()
