// Conformance runner for asun-java.
// Loads ../../cases.json and decodes non-schema-driven cases via Asun.decode(input, Object.class).
// Java has no built-in JSON parser without dependencies, so we ship a tiny one inline.

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;

import io.asun.Asun;

public class Run {

    // ---------- minimal JSON parser (just enough for cases.json) ----------
    static class JParser {
        final String s; int p = 0;
        JParser(String s) { this.s = s; }
        Object parse() { skip(); Object v = val(); skip(); return v; }
        Object val() {
            skip();
            char c = s.charAt(p);
            if (c == '{') return obj();
            if (c == '[') return arr();
            if (c == '"') return str();
            if (c == 't' || c == 'f') return bool();
            if (c == 'n') { p += 4; return null; }
            return num();
        }
        Map<String,Object> obj() {
            Map<String,Object> m = new LinkedHashMap<>();
            p++; skip();
            if (s.charAt(p) == '}') { p++; return m; }
            while (true) {
                skip(); String k = str(); skip();
                p++; // ':'
                m.put(k, val());
                skip();
                char c = s.charAt(p++);
                if (c == ',') continue;
                if (c == '}') return m;
            }
        }
        List<Object> arr() {
            List<Object> a = new ArrayList<>();
            p++; skip();
            if (s.charAt(p) == ']') { p++; return a; }
            while (true) {
                a.add(val()); skip();
                char c = s.charAt(p++);
                if (c == ',') continue;
                if (c == ']') return a;
            }
        }
        String str() {
            StringBuilder sb = new StringBuilder();
            p++;
            while (true) {
                char c = s.charAt(p++);
                if (c == '"') return sb.toString();
                if (c == '\\') {
                    char e = s.charAt(p++);
                    switch (e) {
                        case '"': sb.append('"'); break;
                        case '\\': sb.append('\\'); break;
                        case '/': sb.append('/'); break;
                        case 'n': sb.append('\n'); break;
                        case 't': sb.append('\t'); break;
                        case 'r': sb.append('\r'); break;
                        case 'b': sb.append('\b'); break;
                        case 'f': sb.append('\f'); break;
                        case 'u':
                            int cp = Integer.parseInt(s.substring(p, p+4), 16);
                            p += 4;
                            sb.append((char)cp);
                            break;
                        default: sb.append(e);
                    }
                } else sb.append(c);
            }
        }
        Boolean bool() {
            if (s.startsWith("true", p)) { p += 4; return true; }
            p += 5; return false;
        }
        Object num() {
            int start = p;
            if (s.charAt(p) == '-') p++;
            while (p < s.length() && (Character.isDigit(s.charAt(p)) || s.charAt(p)=='.' || s.charAt(p)=='e' || s.charAt(p)=='E' || s.charAt(p)=='-' || s.charAt(p)=='+')) p++;
            String t = s.substring(start, p);
            if (t.contains(".") || t.contains("e") || t.contains("E")) return Double.parseDouble(t);
            try { return Long.parseLong(t); } catch (NumberFormatException ex) { return Double.parseDouble(t); }
        }
        void skip() {
            while (p < s.length() && Character.isWhitespace(s.charAt(p))) p++;
        }
    }

    static boolean deepEqual(Object a, Object b) {
        if (a == null) return b == null;
        if (b == null) return false;
        if (a instanceof Number && b instanceof Number) {
            Number na = (Number)a, nb = (Number)b;
            if (a instanceof Long && b instanceof Long) return na.longValue() == nb.longValue();
            if (a instanceof Integer && b instanceof Integer) return na.intValue() == nb.intValue();
            if ((a instanceof Long || a instanceof Integer) && (b instanceof Long || b instanceof Integer)) return na.longValue() == nb.longValue();
            double da = na.doubleValue(), db = nb.doubleValue();
            if (Math.abs(da - db) <= 1e-12) return true;
            return da == db;
        }
        if (a instanceof Boolean && b instanceof Boolean) return a.equals(b);
        if (a instanceof String && b instanceof String) return a.equals(b);
        if (a instanceof List && b instanceof List) {
            List<?> la = (List<?>)a, lb = (List<?>)b;
            if (la.size() != lb.size()) return false;
            for (int i = 0; i < la.size(); i++) if (!deepEqual(la.get(i), lb.get(i))) return false;
            return true;
        }
        if (a instanceof Map && b instanceof Map) {
            Map<?,?> ma = (Map<?,?>)a, mb = (Map<?,?>)b;
            if (ma.size() != mb.size()) return false;
            for (Map.Entry<?,?> e : ma.entrySet()) {
                if (!mb.containsKey(e.getKey())) return false;
                if (!deepEqual(e.getValue(), mb.get(e.getKey()))) return false;
            }
            return true;
        }
        return a.equals(b);
    }

    public static void main(String[] args) throws IOException {
        Path casesPath = Paths.get(System.getProperty("user.dir"), "..", "..", "cases.json").normalize();
        String raw = new String(Files.readAllBytes(casesPath));
        Map<String,Object> manifest = (Map<String,Object>) new JParser(raw).parse();
        List<Object> cases = (List<Object>) manifest.get("cases");
        System.out.println("loaded " + cases.size() + " cases from " + casesPath);

        int total = 0, passed = 0, failed = 0, errPassed = 0, errFailed = 0, skipped = 0;
        List<String[]> failures = new ArrayList<>();

        for (Object oc : cases) {
            Map<String,Object> c = (Map<String,Object>) oc;
            total++;
            Boolean sd = (Boolean) c.get("schemaDriven");
            if (sd != null && sd) { skipped++; continue; }
            String input = (String) c.get("input");
            String kind = (String) c.get("kind");
            String id = (String) c.get("id");
            Object got = null;
            boolean threw = false;
            String errMsg = "";
            try {
                got = Asun.decode(input, Object.class);
            } catch (Throwable e) {
                threw = true;
                errMsg = e.getClass().getSimpleName() + ": " + e.getMessage();
            }
            if ("ok".equals(kind)) {
                if (threw) {
                    failed++;
                    if (failures.size() < 25)
                        failures.add(new String[]{id, "expected ok, got error: " + errMsg + "\n    input: " + jsonQ(input)});
                    continue;
                }
                if (!deepEqual(got, c.get("expected"))) {
                    failed++;
                    if (failures.size() < 25)
                        failures.add(new String[]{id, "value mismatch\n    input:    " + jsonQ(input) + "\n    expected: " + c.get("expected") + "\n    actual:   " + got});
                    continue;
                }
                passed++;
            } else {
                if (threw) errPassed++;
                else {
                    errFailed++;
                    if (failures.size() < 25)
                        failures.add(new String[]{id, "expected error, got ok: " + got + "\n    input: " + jsonQ(input)});
                }
            }
        }

        System.out.println();
        System.out.println("================ ASUN-JAVA conformance ================");
        System.out.println("total                : " + total);
        System.out.println("untyped ok-cases pass: " + passed);
        System.out.println("untyped ok-cases fail: " + failed);
        System.out.println("untyped err-cases pass: " + errPassed);
        System.out.println("untyped err-cases fail: " + errFailed);
        System.out.println("skipped (needs typed): " + skipped);
        int exec = total - skipped;
        double pct = exec > 0 ? (double)(passed + errPassed) / exec * 100.0 : 0.0;
        System.out.printf("untyped pass rate    : %d/%d (%.1f%%)%n", passed + errPassed, exec, pct);
        System.out.println("=====================================================");

        for (String[] f : failures) System.out.println("\n[" + f[0] + "]\n    " + f[1]);

        // ---------- Encode (round-trip) ----------
        Path encPath = Paths.get(System.getProperty("user.dir"), "..", "..", "encode-cases.json").normalize();
        int encFailed = 0;
        if (Files.exists(encPath)) {
            String encRaw = new String(Files.readAllBytes(encPath));
            Map<String,Object> em = (Map<String,Object>) new JParser(encRaw).parse();
            List<Object> encCases = (List<Object>) em.get("cases");
            System.out.println("\nloaded " + encCases.size() + " encode cases from " + encPath);
            int encPassed = 0;
            List<String[]> encFailures = new ArrayList<>();
            for (Object oc : encCases) {
                Map<String,Object> c = (Map<String,Object>) oc;
                String id = (String) c.get("id");
                Object val = c.get("value");
                String text;
                try {
                    text = Asun.encode(val);
                } catch (Throwable e) {
                    encFailed++;
                    if (encFailures.size() < 25)
                        encFailures.add(new String[]{id, "encode failed: " + e.getClass().getSimpleName() + ": " + e.getMessage() + "\n    value: " + val});
                    continue;
                }
                Object got;
                try {
                    got = Asun.decode(text, Object.class);
                } catch (Throwable e) {
                    encFailed++;
                    if (encFailures.size() < 25)
                        encFailures.add(new String[]{id, "decode-after-encode failed: " + e.getClass().getSimpleName() + ": " + e.getMessage() + "\n    value:   " + val + "\n    encoded: " + text});
                    continue;
                }
                if (!deepEqual(val, got)) {
                    encFailed++;
                    if (encFailures.size() < 25)
                        encFailures.add(new String[]{id, "round-trip mismatch\n    value:   " + val + "\n    encoded: " + text + "\n    decoded: " + got});
                    continue;
                }
                encPassed++;
            }
            int encTotal = encPassed + encFailed;
            double encPct = encTotal > 0 ? (double)encPassed / encTotal * 100.0 : 0.0;
            System.out.println();
            System.out.println("============ ASUN-JAVA encode round-trip ============");
            System.out.println("total : " + encTotal);
            System.out.println("pass  : " + encPassed);
            System.out.println("fail  : " + encFailed);
            System.out.printf("rate  : %d/%d (%.1f%%)%n", encPassed, encTotal, encPct);
            System.out.println("=====================================================");
            for (String[] f : encFailures) System.out.println("\n[" + f[0] + "]\n    " + f[1]);
        }

        if (failed > 0 || errFailed > 0 || encFailed > 0) System.exit(1);
    }

    static String jsonQ(String s) { return "\"" + s.replace("\\","\\\\").replace("\"","\\\"") + "\""; }
}
