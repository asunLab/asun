package io.ason.examples;

import io.ason.Ason;
import java.util.*;

public class ComplexExample {

    // ========================================================================
    // Type definitions
    // ========================================================================

    public static class Department {
        public String title;
        public Department() {}
        public Department(String title) { this.title = title; }
        @Override public boolean equals(Object o) {
            return o instanceof Department d && Objects.equals(title, d.title);
        }
        @Override public String toString() { return "Department{title=" + title + "}"; }
    }

    public static class Employee {
        public long id;
        public String name;
        public Department dept;
        public List<String> skills;
        public boolean active;
        public Employee() { skills = new ArrayList<>(); }
        @Override public boolean equals(Object o) {
            if (!(o instanceof Employee e)) return false;
            return id == e.id && active == e.active && Objects.equals(name, e.name)
                && Objects.equals(dept, e.dept) && Objects.equals(skills, e.skills);
        }
        @Override public String toString() {
            return "Employee{id=" + id + ", name=" + name + ", dept=" + dept + ", skills=" + skills + ", active=" + active + "}";
        }
    }

    public static class WithMap {
        public String name;
        public Map<String, Long> attrs;
        public WithMap() { attrs = new LinkedHashMap<>(); }
        @Override public boolean equals(Object o) {
            return o instanceof WithMap w && Objects.equals(name, w.name) && Objects.equals(attrs, w.attrs);
        }
        @Override public String toString() { return "WithMap{name=" + name + ", attrs=" + attrs + "}"; }
    }

    public static class Nested {
        public String name;
        public Address addr;
        public Nested() {}
        public Nested(String name, Address addr) { this.name = name; this.addr = addr; }
        @Override public boolean equals(Object o) {
            return o instanceof Nested n && Objects.equals(name, n.name) && Objects.equals(addr, n.addr);
        }
    }

    public static class Address {
        public String city;
        public long zip;
        public Address() {}
        public Address(String city, long zip) { this.city = city; this.zip = zip; }
        @Override public boolean equals(Object o) {
            return o instanceof Address a && zip == a.zip && Objects.equals(city, a.city);
        }
    }

    public static class Note {
        public String text;
        public Note() {}
        public Note(String text) { this.text = text; }
        @Override public boolean equals(Object o) {
            return o instanceof Note n && Objects.equals(text, n.text);
        }
    }

    public static class Measurement {
        public long id;
        public double value;
        public String label;
        public Measurement() {}
        public Measurement(long id, double value, String label) { this.id = id; this.value = value; this.label = label; }
        @Override public boolean equals(Object o) {
            return o instanceof Measurement m && id == m.id && Double.compare(m.value, value) == 0 && Objects.equals(label, m.label);
        }
    }

    public static class Nums {
        public long a;
        public double b;
        public long c;
        public Nums() {}
        public Nums(long a, double b, long c) { this.a = a; this.b = b; this.c = c; }
        @Override public boolean equals(Object o) {
            return o instanceof Nums n && a == n.a && Double.compare(n.b, b) == 0 && c == n.c;
        }
    }

    // 5-level deep: Country > Region > City > District > Street > Building
    public static class Building {
        public String name;
        public long floors;
        public boolean residential;
        public double heightM;
        public Building() {}
        public Building(String name, long floors, boolean res, double h) {
            this.name = name; this.floors = floors; this.residential = res; this.heightM = h;
        }
        @Override public boolean equals(Object o) {
            return o instanceof Building b && floors == b.floors && residential == b.residential
                && Double.compare(b.heightM, heightM) == 0 && Objects.equals(name, b.name);
        }
        @Override public String toString() { return "Building{" + name + "," + floors + "F}"; }
    }

    public static class Street {
        public String name;
        public double lengthKm;
        public List<Building> buildings;
        public Street() { buildings = new ArrayList<>(); }
        public Street(String name, double len, List<Building> b) { this.name = name; this.lengthKm = len; this.buildings = b; }
        @Override public boolean equals(Object o) {
            return o instanceof Street s && Double.compare(s.lengthKm, lengthKm) == 0
                && Objects.equals(name, s.name) && Objects.equals(buildings, s.buildings);
        }
    }

    public static class District {
        public String name;
        public long population;
        public List<Street> streets;
        public District() { streets = new ArrayList<>(); }
        public District(String name, long pop, List<Street> s) { this.name = name; this.population = pop; this.streets = s; }
        @Override public boolean equals(Object o) {
            return o instanceof District d && population == d.population
                && Objects.equals(name, d.name) && Objects.equals(streets, d.streets);
        }
    }

    public static class City {
        public String name;
        public long population;
        public double areaKm2;
        public List<District> districts;
        public City() { districts = new ArrayList<>(); }
        public City(String n, long pop, double area, List<District> d) {
            this.name = n; this.population = pop; this.areaKm2 = area; this.districts = d;
        }
        @Override public boolean equals(Object o) {
            return o instanceof City c && population == c.population
                && Double.compare(c.areaKm2, areaKm2) == 0
                && Objects.equals(name, c.name) && Objects.equals(districts, c.districts);
        }
    }

    public static class Region {
        public String name;
        public List<City> cities;
        public Region() { cities = new ArrayList<>(); }
        public Region(String name, List<City> cities) { this.name = name; this.cities = cities; }
        @Override public boolean equals(Object o) {
            return o instanceof Region r && Objects.equals(name, r.name) && Objects.equals(cities, r.cities);
        }
    }

    public static class Country {
        public String name;
        public String code;
        public long population;
        public double gdpTrillion;
        public List<Region> regions;
        public Country() { regions = new ArrayList<>(); }
        public Country(String n, String c, long p, double g, List<Region> r) {
            this.name = n; this.code = c; this.population = p; this.gdpTrillion = g; this.regions = r;
        }
        @Override public boolean equals(Object o) {
            return o instanceof Country c && population == c.population
                && Double.compare(c.gdpTrillion, gdpTrillion) == 0
                && Objects.equals(name, c.name) && Objects.equals(code, c.code)
                && Objects.equals(regions, c.regions);
        }
    }

    // Service config
    public static class DbConfig {
        public String host; public long port; public long maxConnections; public boolean ssl; public double timeoutMs;
        public DbConfig() {}
        @Override public boolean equals(Object o) {
            return o instanceof DbConfig d && port == d.port && maxConnections == d.maxConnections
                && ssl == d.ssl && Double.compare(d.timeoutMs, timeoutMs) == 0 && Objects.equals(host, d.host);
        }
    }

    public static class CacheConfig {
        public boolean enabled; public long ttlSeconds; public long maxSizeMb;
        public CacheConfig() {}
        @Override public boolean equals(Object o) {
            return o instanceof CacheConfig c && enabled == c.enabled && ttlSeconds == c.ttlSeconds && maxSizeMb == c.maxSizeMb;
        }
    }

    public static class LogConfig {
        public String level; public Optional<String> file; public boolean rotate;
        public LogConfig() { file = Optional.empty(); }
        @Override public boolean equals(Object o) {
            return o instanceof LogConfig l && rotate == l.rotate && Objects.equals(level, l.level) && Objects.equals(file, l.file);
        }
    }

    public static class ServiceConfig {
        public String name; public String version; public DbConfig db; public CacheConfig cache;
        public LogConfig log; public List<String> features; public Map<String, String> env;
        public ServiceConfig() { features = new ArrayList<>(); env = new LinkedHashMap<>(); }
        @Override public boolean equals(Object o) {
            return o instanceof ServiceConfig s && Objects.equals(name, s.name) && Objects.equals(version, s.version)
                && Objects.equals(db, s.db) && Objects.equals(cache, s.cache) && Objects.equals(log, s.log)
                && Objects.equals(features, s.features) && Objects.equals(env, s.env);
        }
    }

    public static class Special {
        public String val;
        public Special() {}
        public Special(String val) { this.val = val; }
        @Override public boolean equals(Object o) {
            return o instanceof Special s && Objects.equals(val, s.val);
        }
    }

    public static class Matrix3D {
        public List<List<List<Long>>> data;
        public Matrix3D() { data = new ArrayList<>(); }
        @Override public boolean equals(Object o) {
            return o instanceof Matrix3D m && Objects.equals(data, m.data);
        }
    }

    // ========================================================================
    // Main
    // ========================================================================

    public static void main(String[] args) {
        System.out.println("=== ASON Complex Examples (Java) ===\n");

        // 1. Nested struct
        System.out.println("1. Nested struct:");
        Employee emp = Ason.decode("{id,name,dept:{title},skills,active}:(1,Alice,(Manager),[rust],true)", Employee.class);
        System.out.println("   " + emp + "\n");

        // 2. Vec with nested structs
        System.out.println("2. Vec with nested structs:");
        String input2 = """
            [{id:int,name:str,dept:{title:str},skills:[str],active:bool}]:
              (1, Alice, (Manager), [Rust, Go], true),
              (2, Bob, (Engineer), [Python], false),
              (3, "Carol Smith", (Director), [Leadership, Strategy], true)""";
        List<Employee> employees = Ason.decodeList(input2, Employee.class);
        for (Employee e : employees) System.out.println("   " + e);

        // 3. Map/Dict field
        System.out.println("\n3. Map/Dict field:");
        WithMap item = Ason.decode("{name,attrs}:(Alice,[(age,30),(score,95)])", WithMap.class);
        System.out.println("   " + item);

        // 4. Nested struct roundtrip
        System.out.println("\n4. Nested struct roundtrip:");
        Nested nested = new Nested("Alice", new Address("NYC", 10001));
        String s = Ason.encode(nested);
        System.out.println("   serialized:   " + s);
        Nested deserialized = Ason.decode(s, Nested.class);
        assert nested.equals(deserialized);
        System.out.println("   ✓ roundtrip OK");

        // 5. Escaped strings
        System.out.println("\n5. Escaped strings:");
        Note note = new Note("say \"hi\", then (wave)\tnewline\nend");
        s = Ason.encode(note);
        System.out.println("   serialized:   " + s);
        Note note2 = Ason.decode(s, Note.class);
        assert note.equals(note2);
        System.out.println("   ✓ escape roundtrip OK");

        // 6. Float fields
        System.out.println("\n6. Float fields:");
        Measurement m = new Measurement(2, 95.0, "score");
        s = Ason.encode(m);
        System.out.println("   serialized: " + s);
        Measurement m2 = Ason.decode(s, Measurement.class);
        assert m.equals(m2);
        System.out.println("   ✓ float roundtrip OK");

        // 7. Negative numbers
        System.out.println("\n7. Negative numbers:");
        Nums n = new Nums(-42, -3.15, Long.MIN_VALUE + 1);
        s = Ason.encode(n);
        System.out.println("   serialized:   " + s);
        Nums n2 = Ason.decode(s, Nums.class);
        assert n.equals(n2);
        System.out.println("   ✓ negative roundtrip OK");

        // 8. Five-level nesting
        System.out.println("\n8. Five-level nesting (Country>Region>City>District>Street>Building):");
        Country country = buildCountry();
        s = Ason.encode(country);
        System.out.println("   serialized (" + s.length() + " bytes)");
        System.out.println("   first 200 chars: " + s.substring(0, Math.min(200, s.length())) + "...");
        Country country2 = Ason.decode(s, Country.class);
        assert country.equals(country2);
        System.out.println("   ✓ 5-level ASON-text roundtrip OK");
        byte[] bin = Ason.encodeBinary(country);
        Country country3 = Ason.decodeBinary(bin, Country.class);
        assert country.equals(country3);
        System.out.println("   ✓ 5-level ASON-bin roundtrip OK");
        System.out.printf("   ASON text: %d B | ASON bin: %d B%n", s.length(), bin.length);

        // 9. Service config with maps + optional + nested
        System.out.println("\n9. Complex config struct (nested + map + optional):");
        ServiceConfig config = buildConfig();
        s = Ason.encode(config);
        System.out.println("   serialized (" + s.length() + " bytes):");
        System.out.println("   " + s);
        ServiceConfig config2 = Ason.decode(s, ServiceConfig.class);
        assert config.equals(config2);
        System.out.println("   ✓ config roundtrip OK");
        bin = Ason.encodeBinary(config);
        ServiceConfig config3 = Ason.decodeBinary(bin, ServiceConfig.class);
        assert config.equals(config3);
        System.out.println("   ✓ config ASON-bin roundtrip OK");

        // 10. Large structure — 100 countries
        System.out.println("\n10. Large structure (100 countries × nested regions):");
        long totalAson = 0, totalBin = 0;
        for (int i = 0; i < 100; i++) {
            Country c = buildCountryN(i);
            String cs = Ason.encode(c);
            byte[] cb = Ason.encodeBinary(c);
            Country cd = Ason.decode(cs, Country.class);
            assert c.equals(cd);
            Country cbd = Ason.decodeBinary(cb, Country.class);
            assert c.equals(cbd);
            totalAson += cs.length();
            totalBin += cb.length;
        }
        System.out.printf("   Total ASON text: %d bytes (%.1f KB)%n", totalAson, totalAson / 1024.0);
        System.out.printf("   Total ASON bin:  %d bytes (%.1f KB)%n", totalBin, totalBin / 1024.0);
        System.out.println("   ✓ all 100 countries roundtrip OK (text + bin)");

        // 11. Typed serialization
        System.out.println("\n11. Typed serialization (encodeTyped):");
        Employee empTyped = new Employee();
        empTyped.id = 1; empTyped.name = "Alice";
        empTyped.dept = new Department("Engineering");
        empTyped.skills = List.of("Rust", "Go"); empTyped.active = true;
        String userTyped = Ason.encodeTyped(empTyped);
        System.out.println("   nested struct: " + userTyped);
        Employee empBack = Ason.decode(userTyped, Employee.class);
        assert empTyped.equals(empBack);
        System.out.println("   ✓ typed nested struct roundtrip OK");

        // 12. Edge cases
        System.out.println("\n12. Edge cases:");
        Special sp = new Special("tabs\there, newlines\nhere, quotes\"and\\backslash");
        s = Ason.encode(sp);
        System.out.println("   special chars: " + s);
        Special sp2 = Ason.decode(s, Special.class);
        assert sp.equals(sp2);

        Special sp3 = new Special("true");
        s = Ason.encode(sp3);
        System.out.println("   bool-like string: " + s);
        assert sp3.equals(Ason.decode(s, Special.class));

        Special sp5 = new Special("12345");
        s = Ason.encode(sp5);
        System.out.println("   number-like string: " + s);
        assert sp5.equals(Ason.decode(s, Special.class));
        System.out.println("   ✓ all edge cases OK");

        // 13. Triple-nested arrays
        System.out.println("\n13. Triple-nested arrays:");
        Matrix3D m3 = new Matrix3D();
        m3.data = List.of(
            List.of(List.of(1L, 2L), List.of(3L, 4L)),
            List.of(List.of(5L, 6L, 7L), List.of(8L))
        );
        s = Ason.encode(m3);
        System.out.println("   " + s);
        Matrix3D m3b = Ason.decode(s, Matrix3D.class);
        assert m3.equals(m3b);
        System.out.println("   ✓ triple-nested array roundtrip OK");

        // 14. Comments
        System.out.println("\n14. Comments:");
        Employee empC = Ason.decode("{id,name,dept:{title},skills,active}:/* inline */ (1,Alice,(HR),[rust],true)", Employee.class);
        System.out.println("   with inline comment: " + empC);
        System.out.println("   ✓ comment parsing OK");

        System.out.println("\n=== All 14 complex examples passed! ===");
    }

    // ========================================================================
    // Data builders
    // ========================================================================

    static Country buildCountry() {
        return new Country("Rustland", "RL", 50_000_000, 1.5, List.of(
            new Region("Northern", List.of(
                new City("Ferriton", 2_000_000, 350.5, List.of(
                    new District("Downtown", 500_000, List.of(
                        new Street("Main St", 2.5, List.of(
                            new Building("Tower A", 50, false, 200.0),
                            new Building("Apt Block 1", 12, true, 40.5)
                        )),
                        new Street("Oak Ave", 1.2, List.of(
                            new Building("Library", 3, false, 15.0)
                        ))
                    )),
                    new District("Harbor", 150_000, List.of(
                        new Street("Dock Rd", 0.8, List.of(
                            new Building("Warehouse 7", 1, false, 8.0)
                        ))
                    ))
                ))
            )),
            new Region("Southern", List.of(
                new City("Crabville", 800_000, 120.0, List.of(
                    new District("Old Town", 200_000, List.of(
                        new Street("Heritage Ln", 0.5, List.of(
                            new Building("Museum", 2, false, 12.0),
                            new Building("Town Hall", 4, false, 20.0)
                        ))
                    ))
                ))
            ))
        ));
    }

    static Country buildCountryN(int i) {
        return new Country("Country_" + i, String.format("C%02d", i % 100),
            1_000_000L + i * 500_000L, i * 0.5,
            new ArrayList<>(List.of(
                buildRegion(i, 0), buildRegion(i, 1), buildRegion(i, 2)
            )));
    }

    static Region buildRegion(int i, int r) {
        return new Region("Region_" + i + "_" + r,
            new ArrayList<>(List.of(buildCity(i, r, 0), buildCity(i, r, 1))));
    }

    static City buildCity(int i, int r, int c) {
        return new City("City_" + i + "_" + r + "_" + c,
            100_000L + c * 50_000L, 50.0 + c * 25.5,
            new ArrayList<>(List.of(
                new District("Dist_" + c, 50_000L + c * 10_000L,
                    new ArrayList<>(List.of(
                        new Street("St_" + c, 1.0 + c * 0.5,
                            new ArrayList<>(List.of(
                                new Building("Bldg_" + c + "_0", 5, true, 15.0),
                                new Building("Bldg_" + c + "_1", 8, false, 25.5)
                            ))
                        )
                    ))
                )
            )));
    }

    static ServiceConfig buildConfig() {
        ServiceConfig config = new ServiceConfig();
        config.name = "my-service";
        config.version = "2.1.0";
        config.db = new DbConfig();
        config.db.host = "db.example.com"; config.db.port = 5432;
        config.db.maxConnections = 100; config.db.ssl = true; config.db.timeoutMs = 3000.5;
        config.cache = new CacheConfig();
        config.cache.enabled = true; config.cache.ttlSeconds = 3600; config.cache.maxSizeMb = 512;
        config.log = new LogConfig();
        config.log.level = "info"; config.log.file = Optional.of("/var/log/app.log"); config.log.rotate = true;
        config.features = new ArrayList<>(List.of("auth", "rate-limit", "websocket"));
        config.env = new LinkedHashMap<>();
        config.env.put("RUST_LOG", "debug");
        config.env.put("DATABASE_URL", "postgres://localhost:5432/mydb");
        config.env.put("SECRET_KEY", "abc123!@#");
        return config;
    }
}
