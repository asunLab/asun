"""ASON Complex Examples — matches Rust complex.rs (skipping enums)."""

import json
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from dataclasses import dataclass, field
from typing import Optional
import ason


# ====================================================================
# Types
# ====================================================================

@dataclass
class Department:
    title: str = ""


@dataclass
class Employee:
    id: int = 0
    name: str = ""
    dept: Department = field(default_factory=Department)
    skills: list[str] = field(default_factory=list)
    active: bool = True


@dataclass
class WithMap:
    name: str = ""
    attrs: dict[str, int] = field(default_factory=dict)


@dataclass
class Address:
    city: str = ""
    zip: int = 0


@dataclass
class Nested:
    name: str = ""
    addr: Address = field(default_factory=Address)


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
    nested_vec: list[list[int]] = field(default_factory=list)


# 5-level deep
@dataclass
class Building:
    name: str = ""
    floors: int = 0
    residential: bool = False
    height_m: float = 0.0


@dataclass
class Street:
    name: str = ""
    length_km: float = 0.0
    buildings: list[Building] = field(default_factory=list)


@dataclass
class District:
    name: str = ""
    population: int = 0
    streets: list[Street] = field(default_factory=list)


@dataclass
class City:
    name: str = ""
    population: int = 0
    area_km2: float = 0.0
    districts: list[District] = field(default_factory=list)


@dataclass
class Region:
    name: str = ""
    cities: list[City] = field(default_factory=list)


@dataclass
class Country:
    name: str = ""
    code: str = ""
    population: int = 0
    gdp_trillion: float = 0.0
    regions: list[Region] = field(default_factory=list)


# 7-level deep
@dataclass
class State:
    name: str = ""
    capital: str = ""
    population: int = 0


@dataclass
class Nation:
    name: str = ""
    states: list[State] = field(default_factory=list)


@dataclass
class Continent:
    name: str = ""
    nations: list[Nation] = field(default_factory=list)


@dataclass
class Planet:
    name: str = ""
    radius_km: float = 0.0
    has_life: bool = False
    continents: list[Continent] = field(default_factory=list)


@dataclass
class SolarSystem:
    name: str = ""
    star_type: str = ""
    planets: list[Planet] = field(default_factory=list)


@dataclass
class Galaxy:
    name: str = ""
    star_count_billions: float = 0.0
    systems: list[SolarSystem] = field(default_factory=list)


@dataclass
class Universe:
    name: str = ""
    age_billion_years: float = 0.0
    galaxies: list[Galaxy] = field(default_factory=list)


# Config types
@dataclass
class DbConfig:
    host: str = ""
    port: int = 0
    max_connections: int = 0
    ssl: bool = False
    timeout_ms: float = 0.0


@dataclass
class CacheConfig:
    enabled: bool = False
    ttl_seconds: int = 0
    max_size_mb: int = 0


@dataclass
class LogConfig:
    level: str = ""
    file: Optional[str] = None
    rotate: bool = False


@dataclass
class ServiceConfig:
    name: str = ""
    version: str = ""
    db: DbConfig = field(default_factory=DbConfig)
    cache: CacheConfig = field(default_factory=CacheConfig)
    log: LogConfig = field(default_factory=LogConfig)
    features: list[str] = field(default_factory=list)
    env: dict[str, str] = field(default_factory=dict)


def main():
    print("=== ASON Complex Examples ===\n")

    # -------------------------------------------------------------------
    # 1. Nested struct
    # -------------------------------------------------------------------
    print("1. Nested struct:")
    emp = ason.load("{id,name,dept:{title},skills,active}:(1,Alice,(Manager),[rust],true)", Employee)
    print(f"   {emp}\n")

    # -------------------------------------------------------------------
    # 2. Vec with nested structs
    # -------------------------------------------------------------------
    print("2. Vec with nested structs:")
    inp = """[{id:int,name:str,dept:{title:str},skills:[str],active:bool}]:
  (1, Alice, (Manager), [Rust, Go], true),
  (2, Bob, (Engineer), [Python], false),
  (3, "Carol Smith", (Director), [Leadership, Strategy], true)"""
    employees = ason.load_slice(inp, Employee)
    for e in employees:
        print(f"   {e}")

    # -------------------------------------------------------------------
    # 3. Map/Dict field
    # -------------------------------------------------------------------
    print("\n3. Map/Dict field:")
    wm = ason.load("{name,attrs}:(Alice,[(age,30),(score,95)])", WithMap)
    print(f"   {wm}")

    # -------------------------------------------------------------------
    # 4. Nested struct roundtrip
    # -------------------------------------------------------------------
    print("\n4. Nested struct roundtrip:")
    nested = Nested(name="Alice", addr=Address(city="NYC", zip=10001))
    s = ason.dump(nested)
    print(f"   serialized:   {s}")
    nested2 = ason.load(s, Nested)
    assert nested == nested2
    print("   ✓ roundtrip OK")

    # -------------------------------------------------------------------
    # 5. Escaped strings
    # -------------------------------------------------------------------
    print("\n5. Escaped strings:")

    @dataclass
    class Note:
        text: str = ""

    note = Note(text='say "hi", then (wave)\tnewline\nend')
    s = ason.dump(note)
    print(f"   serialized:   {s}")
    note2 = ason.load(s, Note)
    assert note == note2
    print("   ✓ escape roundtrip OK")

    # -------------------------------------------------------------------
    # 6. Float fields
    # -------------------------------------------------------------------
    print("\n6. Float fields:")

    @dataclass
    class Measurement:
        id: int = 0
        value: float = 0.0
        label: str = ""

    m = Measurement(id=2, value=95.0, label="score")
    s = ason.dump(m)
    print(f"   serialized: {s}")
    m2 = ason.load(s, Measurement)
    assert m == m2
    print("   ✓ float roundtrip OK")

    # -------------------------------------------------------------------
    # 7. Negative numbers
    # -------------------------------------------------------------------
    print("\n7. Negative numbers:")

    @dataclass
    class Nums:
        a: int = 0
        b: float = 0.0
        c: int = 0

    n = Nums(a=-42, b=-3.14, c=-9223372036854775807)
    s = ason.dump(n)
    print(f"   serialized:   {s}")
    n2 = ason.load(s, Nums)
    assert n.a == n2.a and abs(n.b - n2.b) < 1e-10 and n.c == n2.c
    print("   ✓ negative roundtrip OK")

    # -------------------------------------------------------------------
    # 8. All types struct
    # -------------------------------------------------------------------
    print("\n8. All types struct:")
    all_t = AllTypes(
        b=True, i64v=-9223372036854775807,
        f64v=2.718281828459045,
        s='hello, world (test) [arr]',
        opt_some=42, opt_none=None,
        vec_int=[1, 2, 3, -4, 0],
        vec_str=["alpha", "beta gamma", "delta"],
        nested_vec=[[1, 2], [3, 4, 5]],
    )
    s = ason.dump(all_t)
    print(f"   serialized ({len(s)} bytes):")
    print(f"   {s}")
    all_t2 = ason.load(s, AllTypes)
    assert all_t == all_t2
    print("   ✓ all-types roundtrip OK")

    # -------------------------------------------------------------------
    # 9. 5-level deep: Country > Region > City > District > Street > Building
    # -------------------------------------------------------------------
    print("\n9. Five-level nesting (Country>Region>City>District>Street>Building):")
    country = Country(
        name="Rustland", code="RL", population=50_000_000, gdp_trillion=1.5,
        regions=[
            Region(name="Northern", cities=[
                City(name="Ferriton", population=2_000_000, area_km2=350.5, districts=[
                    District(name="Downtown", population=500_000, streets=[
                        Street(name="Main St", length_km=2.5, buildings=[
                            Building(name="Tower A", floors=50, residential=False, height_m=200.0),
                            Building(name="Apt Block 1", floors=12, residential=True, height_m=40.5),
                        ]),
                        Street(name="Oak Ave", length_km=1.2, buildings=[
                            Building(name="Library", floors=3, residential=False, height_m=15.0),
                        ]),
                    ]),
                    District(name="Harbor", population=150_000, streets=[
                        Street(name="Dock Rd", length_km=0.8, buildings=[
                            Building(name="Warehouse 7", floors=1, residential=False, height_m=8.0),
                        ]),
                    ]),
                ]),
            ]),
            Region(name="Southern", cities=[
                City(name="Crabville", population=800_000, area_km2=120.0, districts=[
                    District(name="Old Town", population=200_000, streets=[
                        Street(name="Heritage Ln", length_km=0.5, buildings=[
                            Building(name="Museum", floors=2, residential=False, height_m=12.0),
                            Building(name="Town Hall", floors=4, residential=False, height_m=20.0),
                        ]),
                    ]),
                ]),
            ]),
        ],
    )
    s = ason.dump(country)
    print(f"   serialized ({len(s)} bytes)")
    print(f"   first 200 chars: {s[:200]}...")
    country2 = ason.load(s, Country)
    assert country == country2
    print("   ✓ 5-level roundtrip OK")
    j = json.dumps(country.__dict__, default=lambda o: o.__dict__)
    print(f"   ASON: {len(s)} bytes | JSON: {len(j)} bytes | saving {(1 - len(s)/len(j))*100:.0f}%")

    # -------------------------------------------------------------------
    # 10. 7-level deep
    # -------------------------------------------------------------------
    print("\n10. Seven-level nesting (Universe>Galaxy>SolarSystem>Planet>Continent>Nation>State):")
    universe = Universe(
        name="Observable", age_billion_years=13.8,
        galaxies=[Galaxy(
            name="Milky Way", star_count_billions=250.0,
            systems=[SolarSystem(
                name="Sol", star_type="G2V",
                planets=[
                    Planet(name="Earth", radius_km=6371.0, has_life=True, continents=[
                        Continent(name="Asia", nations=[
                            Nation(name="Japan", states=[
                                State(name="Tokyo", capital="Shinjuku", population=14_000_000),
                                State(name="Osaka", capital="Osaka City", population=8_800_000),
                            ]),
                            Nation(name="China", states=[
                                State(name="Beijing", capital="Beijing", population=21_500_000),
                            ]),
                        ]),
                        Continent(name="Europe", nations=[
                            Nation(name="Germany", states=[
                                State(name="Bavaria", capital="Munich", population=13_000_000),
                                State(name="Berlin", capital="Berlin", population=3_600_000),
                            ]),
                        ]),
                    ]),
                    Planet(name="Mars", radius_km=3389.5, has_life=False, continents=[]),
                ],
            )],
        )],
    )
    s = ason.dump(universe)
    print(f"   serialized ({len(s)} bytes)")
    universe2 = ason.load(s, Universe)
    assert universe == universe2
    print("   ✓ 7-level roundtrip OK")
    j = json.dumps(universe.__dict__, default=lambda o: o.__dict__)
    print(f"   ASON: {len(s)} bytes | JSON: {len(j)} bytes | saving {(1 - len(s)/len(j))*100:.0f}%")

    # -------------------------------------------------------------------
    # 11. Service config
    # -------------------------------------------------------------------
    print("\n11. Complex config struct (nested + map + optional):")
    config = ServiceConfig(
        name="my-service", version="2.1.0",
        db=DbConfig(host="db.example.com", port=5432, max_connections=100, ssl=True, timeout_ms=3000.5),
        cache=CacheConfig(enabled=True, ttl_seconds=3600, max_size_mb=512),
        log=LogConfig(level="info", file="/var/log/app.log", rotate=True),
        features=["auth", "rate-limit", "websocket"],
        env={"RUST_LOG": "debug", "DATABASE_URL": "postgres://localhost:5432/mydb", "SECRET_KEY": "abc123!@#"},
    )
    s = ason.dump(config)
    print(f"   serialized ({len(s)} bytes):")
    print(f"   {s}")
    config2 = ason.load(s, ServiceConfig)
    assert config == config2
    print("   ✓ config roundtrip OK")
    j = json.dumps(config.__dict__, default=lambda o: o.__dict__)
    print(f"   ASON: {len(s)} bytes | JSON: {len(j)} bytes | saving {(1 - len(s)/len(j))*100:.0f}%")

    # -------------------------------------------------------------------
    # 12. Large structure — 100 countries
    # -------------------------------------------------------------------
    print("\n12. Large structure (100 countries × nested regions):")
    countries: list[Country] = []
    for i in range(100):
        regions_list: list[Region] = []
        for r in range(3):
            cities_list: list[City] = []
            for c in range(2):
                cities_list.append(City(
                    name=f"City_{i}_{r}_{c}", population=100_000 + c * 50_000,
                    area_km2=50.0 + c * 25.5,
                    districts=[District(
                        name=f"Dist_{c}", population=50_000 + c * 10_000,
                        streets=[Street(
                            name=f"St_{c}", length_km=1.0 + c * 0.5,
                            buildings=[
                                Building(name=f"Bldg_{c}_0", floors=5, residential=True, height_m=15.0),
                                Building(name=f"Bldg_{c}_1", floors=8, residential=False, height_m=25.5),
                            ],
                        )],
                    )],
                ))
            regions_list.append(Region(name=f"Region_{i}_{r}", cities=cities_list))
        countries.append(Country(
            name=f"Country_{i}", code=f"C{i%100:02d}",
            population=1_000_000 + i * 500_000, gdp_trillion=i * 0.5,
            regions=regions_list,
        ))

    total_ason, total_json = 0, 0
    for c in countries:
        s = ason.dump(c)
        j = json.dumps(c.__dict__, default=lambda o: o.__dict__)
        c2 = ason.load(s, Country)
        assert c == c2
        total_ason += len(s)
        total_json += len(j)
    print("   100 countries with 5-level nesting:")
    print(f"   Total ASON: {total_ason} bytes ({total_ason/1024:.1f} KB)")
    print(f"   Total JSON: {total_json} bytes ({total_json/1024:.1f} KB)")
    print(f"   Saving: {(1 - total_ason/total_json)*100:.0f}%")
    print("   ✓ all 100 countries roundtrip OK")

    # -------------------------------------------------------------------
    # 13. Deserialize with nested schema type hints
    # -------------------------------------------------------------------
    print("\n13. Deserialize with nested schema type hints:")
    deep_input = "{name:str,code:str,population:int,gdp_trillion:float,regions:[{name:str,cities:[{name:str,population:int,area_km2:float,districts:[{name:str,population:int,streets:[{name:str,length_km:float,buildings:[{name:str,floors:int,residential:bool,height_m:float}]}]}]}]}]}:(TestLand,TL,1000000,0.5,[(TestRegion,[(TestCity,500000,100.0,[(Central,250000,[(Main St,2.5,[(HQ,10,false,45.0)])])])])])"
    dc = ason.load(deep_input, Country)
    assert dc.name == "TestLand"
    bld = dc.regions[0].cities[0].districts[0].streets[0].buildings[0]
    assert bld.name == "HQ"
    print("   ✓ deep schema type-hint parse OK")
    print(f"   Building at depth 6: {bld}")

    # -------------------------------------------------------------------
    # 14. Typed serialization (dump_typed)
    # -------------------------------------------------------------------
    print("\n14. Typed serialization (dump_typed):")
    emp_typed = ason.dump_typed(Employee(id=1, name="Alice", dept=Department(title="Engineering"),
                                        skills=["Rust", "Go"], active=True))
    print(f"   nested struct: {emp_typed}")
    emp_back = ason.load(emp_typed, Employee)
    assert emp_back.name == "Alice"
    print("   ✓ typed nested struct roundtrip OK")

    all_typed = ason.dump_typed(all_t)
    print(f"   all-types ({len(all_typed)} bytes): {all_typed[:80]}...")
    all_back = ason.load(all_typed, AllTypes)
    assert all_t == all_back
    print("   ✓ typed all-types roundtrip OK")

    config_typed = ason.dump_typed(config)
    print(f"   config ({len(config_typed)} bytes): {config_typed[:100]}...")
    config_back = ason.load(config_typed, ServiceConfig)
    assert config == config_back
    print("   ✓ typed config roundtrip OK")

    untyped = ason.dump(config)
    print(f"   untyped: {len(untyped)} bytes | typed: {len(config_typed)} bytes | overhead: {len(config_typed) - len(untyped)} bytes")

    # -------------------------------------------------------------------
    # 15. Edge cases
    # -------------------------------------------------------------------
    print("\n15. Edge cases:")

    @dataclass
    class WithVec:
        items: list[int] = field(default_factory=list)

    wv = WithVec(items=[])
    s = ason.dump(wv)
    print(f"   empty vec: {s}")
    wv2 = ason.load(s, WithVec)
    assert wv == wv2

    @dataclass
    class Special:
        val: str = ""

    sp = Special(val='tabs\there, newlines\nhere, quotes"and\\backslash')
    s = ason.dump(sp)
    print(f"   special chars: {s}")
    sp2 = ason.load(s, Special)
    assert sp == sp2

    sp3 = Special(val="true")
    s = ason.dump(sp3)
    print(f"   bool-like string: {s}")
    sp4 = ason.load(s, Special)
    assert sp3 == sp4

    sp5 = Special(val="12345")
    s = ason.dump(sp5)
    print(f"   number-like string: {s}")
    sp6 = ason.load(s, Special)
    assert sp5 == sp6
    print("   ✓ all edge cases OK")

    # -------------------------------------------------------------------
    # 16. Triple-nested arrays
    # -------------------------------------------------------------------
    print("\n16. Triple-nested arrays:")

    @dataclass
    class Matrix3D:
        data: list[list[list[int]]] = field(default_factory=list)

    m3 = Matrix3D(data=[[[1, 2], [3, 4]], [[5, 6, 7], [8]]])
    s = ason.dump(m3)
    print(f"   {s}")
    m3b = ason.load(s, Matrix3D)
    assert m3 == m3b
    print("   ✓ triple-nested array roundtrip OK")

    # -------------------------------------------------------------------
    # 17. Comments
    # -------------------------------------------------------------------
    print("\n17. Comments:")
    emp_c = ason.load("{id,name,dept:{title},skills,active}:/* inline */ (1,Alice,(HR),[rust],true)", Employee)
    print(f"   with inline comment: {emp_c}")
    print("   ✓ comment parsing OK")

    # -------------------------------------------------------------------
    # Summary
    # -------------------------------------------------------------------
    print("\n=== All 17 complex examples passed! ===")


if __name__ == "__main__":
    main()
