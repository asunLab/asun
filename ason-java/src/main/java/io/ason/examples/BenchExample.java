package io.ason.examples;

import io.ason.Ason;
import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import java.util.*;

public class BenchExample {

    // ========================================================================
    // Struct definitions
    // ========================================================================

    public static class User {
        public long id;
        public String name;
        public String email;
        public long age;
        public double score;
        public boolean active;
        public String role;
        public String city;

        public User() {
        }

        @Override
        public boolean equals(Object o) {
            if (!(o instanceof User u))
                return false;
            return id == u.id && age == u.age && active == u.active
                    && Double.compare(u.score, score) == 0
                    && Objects.equals(name, u.name) && Objects.equals(email, u.email)
                    && Objects.equals(role, u.role) && Objects.equals(city, u.city);
        }
    }

    public static class Task {
        public long id;
        public String title;
        public long priority;
        public boolean done;
        public double hours;

        public Task() {
        }

        @Override
        public boolean equals(Object o) {
            if (!(o instanceof Task t))
                return false;
            return id == t.id && priority == t.priority && done == t.done
                    && Double.compare(t.hours, hours) == 0 && Objects.equals(title, t.title);
        }
    }

    public static class Project {
        public String name;
        public double budget;
        public boolean active;
        public List<Task> tasks;

        public Project() {
            tasks = new ArrayList<>();
        }

        @Override
        public boolean equals(Object o) {
            if (!(o instanceof Project p))
                return false;
            return active == p.active && Double.compare(p.budget, budget) == 0
                    && Objects.equals(name, p.name) && Objects.equals(tasks, p.tasks);
        }
    }

    public static class Team {
        public String name;
        public String lead;
        public long size;
        public List<Project> projects;

        public Team() {
            projects = new ArrayList<>();
        }

        @Override
        public boolean equals(Object o) {
            if (!(o instanceof Team t))
                return false;
            return size == t.size && Objects.equals(name, t.name)
                    && Objects.equals(lead, t.lead) && Objects.equals(projects, t.projects);
        }
    }

    public static class Division {
        public String name;
        public String location;
        public long headcount;
        public List<Team> teams;

        public Division() {
            teams = new ArrayList<>();
        }

        @Override
        public boolean equals(Object o) {
            if (!(o instanceof Division d))
                return false;
            return headcount == d.headcount && Objects.equals(name, d.name)
                    && Objects.equals(location, d.location) && Objects.equals(teams, d.teams);
        }
    }

    public static class Company {
        public String name;
        public long founded;
        public double revenueM;
        public boolean isPublic;
        public List<Division> divisions;
        public List<String> tags;

        public Company() {
            divisions = new ArrayList<>();
            tags = new ArrayList<>();
        }

        @Override
        public boolean equals(Object o) {
            if (!(o instanceof Company c))
                return false;
            return founded == c.founded && isPublic == c.isPublic
                    && Double.compare(c.revenueM, revenueM) == 0
                    && Objects.equals(name, c.name) && Objects.equals(divisions, c.divisions)
                    && Objects.equals(tags, c.tags);
        }
    }

    // ========================================================================
    // Data generators
    // ========================================================================

    static final String[] NAMES = { "Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Hank" };
    static final String[] ROLES = { "engineer", "designer", "manager", "analyst" };
    static final String[] CITIES = { "NYC", "LA", "Chicago", "Houston", "Phoenix" };

    static List<User> generateUsers(int n) {
        List<User> users = new ArrayList<>(n);
        for (int i = 0; i < n; i++) {
            User u = new User();
            u.id = i;
            u.name = NAMES[i % NAMES.length];
            u.email = NAMES[i % NAMES.length].toLowerCase() + "@example.com";
            u.age = 25 + (i % 40);
            u.score = 50.0 + (i % 50) + 0.5;
            u.active = i % 3 != 0;
            u.role = ROLES[i % ROLES.length];
            u.city = CITIES[i % CITIES.length];
            users.add(u);
        }
        return users;
    }

    static List<Company> generateCompanies(int n) {
        String[] locs = { "NYC", "London", "Tokyo", "Berlin" };
        List<Company> companies = new ArrayList<>(n);
        for (int i = 0; i < n; i++) {
            Company co = new Company();
            co.name = "Corp_" + i;
            co.founded = 1990 + (i % 35);
            co.revenueM = 10.0 + i * 5.5;
            co.isPublic = i % 2 == 0;
            co.tags = new ArrayList<>(List.of("enterprise", "tech", "sector_" + (i % 5)));
            for (int d = 0; d < 2; d++) {
                Division div = new Division();
                div.name = "Div_" + i + "_" + d;
                div.location = locs[d % 4];
                div.headcount = 50 + d * 20;
                for (int t = 0; t < 2; t++) {
                    Team team = new Team();
                    team.name = "Team_" + i + "_" + d + "_" + t;
                    team.lead = NAMES[t % 4];
                    team.size = 5 + t * 2;
                    for (int p = 0; p < 3; p++) {
                        Project proj = new Project();
                        proj.name = "Proj_" + t + "_" + p;
                        proj.budget = 100.0 + p * 50.5;
                        proj.active = p % 2 == 0;
                        for (int tk = 0; tk < 4; tk++) {
                            Task task = new Task();
                            task.id = i * 100L + d * 10L + t * 5L + tk;
                            task.title = "Task_" + tk;
                            task.priority = (tk % 3) + 1;
                            task.done = tk % 2 == 0;
                            task.hours = 2.0 + tk * 1.5;
                            proj.tasks.add(task);
                        }
                        team.projects.add(proj);
                    }
                    div.teams.add(team);
                }
                co.divisions.add(div);
            }
            companies.add(co);
        }
        return companies;
    }

    // ========================================================================
    // Benchmark helpers
    // ========================================================================

    static String formatBytes(long b) {
        if (b >= 1048576)
            return String.format("%.1f MB", b / 1048576.0);
        if (b >= 1024)
            return String.format("%.1f KB", b / 1024.0);
        return b + " B";
    }

    static final Gson gson = new Gson();

    // ========================================================================
    // Benchmark runners
    // ========================================================================

    static void benchFlat(int count, int iterations) {
        List<User> users = generateUsers(count);
        java.lang.reflect.Type listType = new TypeToken<List<User>>() {}.getType();

        // Gson serialize
        long start = System.nanoTime();
        String gsonStr = "";
        for (int i = 0; i < iterations; i++)
            gsonStr = gson.toJson(users);
        double gsonSerMs = (System.nanoTime() - start) / 1e6;

        // ASON serialize
        start = System.nanoTime();
        String asonStr = "";
        for (int i = 0; i < iterations; i++)
            asonStr = Ason.encode(users);
        double asonSerMs = (System.nanoTime() - start) / 1e6;

        // Gson deserialize
        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            List<User> r = gson.fromJson(gsonStr, listType);
        }
        double gsonDeMs = (System.nanoTime() - start) / 1e6;

        // ASON deserialize (byte[] API avoids String→byte[] per iteration)
        byte[] asonBytes = asonStr.getBytes(java.nio.charset.StandardCharsets.UTF_8);
        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            List<User> r = Ason.decodeList(asonBytes, User.class);
        }
        double asonDeMs = (System.nanoTime() - start) / 1e6;

        // Binary
        start = System.nanoTime();
        byte[] binBuf = new byte[0];
        for (int i = 0; i < iterations; i++)
            binBuf = Ason.encodeBinary(users);
        double binSerMs = (System.nanoTime() - start) / 1e6;

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            List<User> r = Ason.decodeBinaryList(binBuf, User.class);
        }
        double binDeMs = (System.nanoTime() - start) / 1e6;

        double serRatio = gsonSerMs / asonSerMs;
        double deRatio = gsonDeMs / asonDeMs;
        double saving = (1.0 - (double) asonStr.length() / gsonStr.length()) * 100;
        double binSaving = (1.0 - (double) binBuf.length / gsonStr.length()) * 100;

        System.out.printf("  Flat struct × %d (%d fields)%n", count, 8);
        System.out.printf("    Serialize:   Gson %8.2fms | ASON %8.2fms | ratio %.2fx %s%n",
                gsonSerMs, asonSerMs, serRatio, serRatio >= 1 ? "✓ ASON faster" : "");
        System.out.printf("    Deserialize: Gson %8.2fms | ASON %8.2fms | ratio %.2fx %s%n",
                gsonDeMs, asonDeMs, deRatio, deRatio >= 1 ? "✓ ASON faster" : "");
        System.out.printf("    BIN ser: %8.2fms | BIN de: %8.2fms%n", binSerMs, binDeMs);
        System.out.printf("    Size: Gson %8d B | ASON %8d B (%.0f%% smaller) | BIN %8d B (%.0f%% smaller)%n",
                gsonStr.length(), asonStr.length(), saving, binBuf.length, binSaving);
        System.out.println();
    }

    static void benchDeep(int count, int iterations) {
        List<Company> companies = generateCompanies(count);
        java.lang.reflect.Type listType = new TypeToken<List<Company>>() {}.getType();

        long start = System.nanoTime();
        String gsonStr = "";
        for (int i = 0; i < iterations; i++)
            gsonStr = gson.toJson(companies);
        double gsonSerMs = (System.nanoTime() - start) / 1e6;

        start = System.nanoTime();
        String asonStr = "";
        for (int i = 0; i < iterations; i++)
            asonStr = Ason.encode(companies);
        double asonSerMs = (System.nanoTime() - start) / 1e6;

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            List<Company> r = gson.fromJson(gsonStr, listType);
        }
        double gsonDeMs = (System.nanoTime() - start) / 1e6;

        byte[] asonBytes = asonStr.getBytes(java.nio.charset.StandardCharsets.UTF_8);
        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            List<Company> r = Ason.decodeList(asonBytes, Company.class);
        }
        double asonDeMs = (System.nanoTime() - start) / 1e6;

        start = System.nanoTime();
        byte[] binBuf = new byte[0];
        for (int i = 0; i < iterations; i++)
            binBuf = Ason.encodeBinary(companies);
        double binSerMs = (System.nanoTime() - start) / 1e6;

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            List<Company> r = Ason.decodeBinaryList(binBuf, Company.class);
        }
        double binDeMs = (System.nanoTime() - start) / 1e6;

        double serRatio = gsonSerMs / asonSerMs;
        double deRatio = gsonDeMs / asonDeMs;
        double saving = (1.0 - (double) asonStr.length() / gsonStr.length()) * 100;
        double binSaving = (1.0 - (double) binBuf.length / gsonStr.length()) * 100;

        System.out.printf("  Deep struct × %d (5-level nested, ~48 nodes each)%n", count);
        System.out.printf("    Serialize:   Gson %8.2fms | ASON %8.2fms | ratio %.2fx %s%n",
                gsonSerMs, asonSerMs, serRatio, serRatio >= 1 ? "✓ ASON faster" : "");
        System.out.printf("    Deserialize: Gson %8.2fms | ASON %8.2fms | ratio %.2fx %s%n",
                gsonDeMs, asonDeMs, deRatio, deRatio >= 1 ? "✓ ASON faster" : "");
        System.out.printf("    BIN ser: %8.2fms | BIN de: %8.2fms%n", binSerMs, binDeMs);
        System.out.printf("    Size: Gson %8d B | ASON %8d B (%.0f%% smaller) | BIN %8d B (%.0f%% smaller)%n",
                gsonStr.length(), asonStr.length(), saving, binBuf.length, binSaving);
        System.out.println();
    }

    static void benchSingleRoundtrip(int iterations) {
        User user = new User();
        user.id = 1;
        user.name = "Alice";
        user.email = "alice@example.com";
        user.age = 30;
        user.score = 95.5;
        user.active = true;
        user.role = "engineer";
        user.city = "NYC";

        long start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            String s = Ason.encode(user);
            Ason.decode(s.getBytes(java.nio.charset.StandardCharsets.UTF_8), User.class);
        }
        double asonMs = (System.nanoTime() - start) / 1e6;

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            String s = gson.toJson(user);
            gson.fromJson(s, User.class);
        }
        double gsonMs = (System.nanoTime() - start) / 1e6;

        System.out.printf("  Flat:  ASON %6.2fms | Gson %6.2fms | ratio %.2fx%n",
                asonMs, gsonMs, gsonMs / asonMs);

        // Deep single
        Company company = generateCompanies(1).getFirst();
        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            String s = Ason.encode(company);
            Ason.decode(s.getBytes(java.nio.charset.StandardCharsets.UTF_8), Company.class);
        }
        asonMs = (System.nanoTime() - start) / 1e6;

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            String s = gson.toJson(company);
            gson.fromJson(s, Company.class);
        }
        gsonMs = (System.nanoTime() - start) / 1e6;

        System.out.printf("  Deep:  ASON %6.2fms | Gson %6.2fms | ratio %.2fx%n",
                asonMs, gsonMs, gsonMs / asonMs);
    }

    // ========================================================================
    // Main
    // ========================================================================

    public static void main(String[] args) {
        System.out.println("╔══════════════════════════════════════════════════════════════╗");
        System.out.println("║            ASON vs JSON (Gson) Benchmark — Java              ║");
        System.out.println("╚══════════════════════════════════════════════════════════════╝");

        System.out.printf("%nSystem: %s %s | JDK %s%n",
                System.getProperty("os.name"), System.getProperty("os.arch"),
                System.getProperty("java.version"));

        Runtime rt = Runtime.getRuntime();
        long memBefore = rt.totalMemory() - rt.freeMemory();
        System.out.printf("Heap before: %s%n%n", formatBytes(memBefore));

        int iterations = 100;
        System.out.println("Iterations per test: " + iterations);

        // Warmup: trigger JIT compilation for both ASON and Gson
        System.out.println("Warming up JIT...");
        java.lang.reflect.Type warmUserListType = new TypeToken<List<User>>() {}.getType();
        java.lang.reflect.Type warmCompanyListType = new TypeToken<List<Company>>() {}.getType();
        List<User> warmUsers = generateUsers(100);
        byte[] warmAsonFlatBytes = Ason.encode(warmUsers).getBytes(java.nio.charset.StandardCharsets.UTF_8);
        String warmGsonFlat = gson.toJson(warmUsers);
        for (int i = 0; i < 1000; i++) {
            gson.toJson(warmUsers);
            Ason.encode(warmUsers);
            gson.fromJson(warmGsonFlat, warmUserListType);
            Ason.decodeList(warmAsonFlatBytes, User.class);
        }
        // Warmup deep struct encode+decode paths
        List<Company> warmCos = generateCompanies(10);
        String warmAsonDeep = Ason.encode(warmCos);
        byte[] warmAsonDeepBytes = warmAsonDeep.getBytes(java.nio.charset.StandardCharsets.UTF_8);
        String warmGsonDeep = gson.toJson(warmCos);
        for (int i = 0; i < 500; i++) {
            gson.toJson(warmCos);
            Ason.encode(warmCos);
            gson.fromJson(warmGsonDeep, warmCompanyListType);
            Ason.decodeList(warmAsonDeepBytes, Company.class);
        }
        System.out.println("Warmup complete.\n");

        // Section 1: Flat struct
        System.out.println("\n┌─────────────────────────────────────────────┐");
        System.out.println("│  Section 1: Flat Struct (schema-driven vec) │");
        System.out.println("└─────────────────────────────────────────────┘");
        for (int count : new int[] { 100, 500, 1000, 5000 }) {
            benchFlat(count, iterations);
        }

        // Section 2: 5-level deep nested struct
        System.out.println("┌──────────────────────────────────────────────────────────┐");
        System.out.println("│  Section 2: 5-Level Deep Nesting (Company hierarchy)     │");
        System.out.println("└──────────────────────────────────────────────────────────┘");
        for (int count : new int[] { 10, 50, 100 }) {
            benchDeep(count, iterations);
        }

        // Section 3: Single struct roundtrip
        System.out.println("┌──────────────────────────────────────────────┐");
        System.out.println("│  Section 3: Single Struct Roundtrip (10000x) │");
        System.out.println("└──────────────────────────────────────────────┘");
        benchSingleRoundtrip(10000);

        // Section 4: Throughput summary
        System.out.println("\n┌──────────────────────────────────────────────┐");
        System.out.println("│  Section 4: Throughput Summary               │");
        System.out.println("└──────────────────────────────────────────────┘");
        List<User> users1k = generateUsers(1000);
        String gsonStr1k = gson.toJson(users1k);
        String asonStr1k = Ason.encode(users1k);
        int iters = 100;
        java.lang.reflect.Type userListType = new TypeToken<List<User>>() {}.getType();

        long start = System.nanoTime();
        for (int i = 0; i < iters; i++)
            gson.toJson(users1k);
        double gsonSerDur = (System.nanoTime() - start) / 1e9;

        start = System.nanoTime();
        for (int i = 0; i < iters; i++)
            Ason.encode(users1k);
        double asonSerDur = (System.nanoTime() - start) / 1e9;

        start = System.nanoTime();
        for (int i = 0; i < iters; i++) {
            List<User> r = gson.fromJson(gsonStr1k, userListType);
        }
        double gsonDeDur = (System.nanoTime() - start) / 1e9;

        byte[] asonBytes1k = asonStr1k.getBytes(java.nio.charset.StandardCharsets.UTF_8);
        start = System.nanoTime();
        for (int i = 0; i < iters; i++)
            Ason.decodeList(asonBytes1k, User.class);
        double asonDeDur = (System.nanoTime() - start) / 1e9;

        double totalRecords = 1000.0 * iters;
        System.out.printf("  Serialize throughput (1000 records × %d iters):%n", iters);
        System.out.printf("    Gson: %.0f records/s  (%.1f MB/s)%n",
                totalRecords / gsonSerDur, gsonStr1k.length() * (double) iters / gsonSerDur / 1048576);
        System.out.printf("    ASON: %.0f records/s  (%.1f MB/s)%n",
                totalRecords / asonSerDur, asonStr1k.length() * (double) iters / asonSerDur / 1048576);
        System.out.printf("    Speed: %.2fx%n", (totalRecords / asonSerDur) / (totalRecords / gsonSerDur));
        System.out.printf("  Deserialize throughput:%n");
        System.out.printf("    Gson: %.0f records/s  (%.1f MB/s)%n",
                totalRecords / gsonDeDur, gsonStr1k.length() * (double) iters / gsonDeDur / 1048576);
        System.out.printf("    ASON: %.0f records/s  (%.1f MB/s)%n",
                totalRecords / asonDeDur, asonStr1k.length() * (double) iters / asonDeDur / 1048576);
        System.out.printf("    Speed: %.2fx%n", (totalRecords / asonDeDur) / (totalRecords / gsonDeDur));

        // Memory
        long memAfter = rt.totalMemory() - rt.freeMemory();
        System.out.printf("%n  Memory:%n");
        System.out.printf("    Initial heap: %s%n", formatBytes(memBefore));
        System.out.printf("    Final heap:   %s%n", formatBytes(memAfter));
        System.out.printf("    Delta:        %s%n", formatBytes(memAfter - memBefore));

        System.out.println("\n╔══════════════════════════════════════════════════════════════╗");
        System.out.println("║                    Benchmark Complete                        ║");
        System.out.println("╚══════════════════════════════════════════════════════════════╝");
    }
}
