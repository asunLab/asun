// Conformance runner for asun-swift.
//
// Loads ../../cases.json (untyped decode) and ../../encode-cases.json
// (round-trip), driving each case through AsunSwift.decode / AsunSwift.encode.
//
// Output mirrors the cs / cpp / c / java runners so the cross-language
// dashboard can grep the same lines.

import Foundation
import AsunSwift

// ---------- JSON → AsunValue ----------

func jsonToAsun(_ any: Any) -> AsunValue {
    if any is NSNull { return .null }
    if let n = any as? NSNumber {
        // Disambiguate Bool from Int — NSNumber bridges Swift Bool to NSNumber
        // with objCType "c", whereas integers use "q" / "i" etc.
        let t = String(cString: n.objCType)
        if t == "c" || t == "B" {
            return .bool(n.boolValue)
        }
        // Try int first
        if t == "q" || t == "i" || t == "l" || t == "s" || t == "Q" || t == "I" || t == "L" || t == "S" {
            return .int(n.int64Value)
        }
        // Float / double
        let d = n.doubleValue
        if d.rounded() == d, d >= Double(Int64.min), d <= Double(Int64.max),
           t == "d" || t == "f" {
            // Even floats that are whole numbers should be treated as float when JSON wrote them with a decimal
            return .float(d)
        }
        return .float(d)
    }
    if let s = any as? String { return .string(s) }
    if let arr = any as? [Any] {
        return .array(arr.map { jsonToAsun($0) })
    }
    if let dict = any as? [String: Any] {
        var out: [String: AsunValue] = [:]
        for (k, v) in dict {
            out[k] = jsonToAsun(v)
        }
        return .object(out)
    }
    return .null
}

// JSON parsing helper that detects whether numbers are integers or floats by
// re-parsing with NSNumber. JSONSerialization gives NSNumber but we can detect
// via objCType.

// ---------- AsunValue equality with tolerant numeric compare ----------

func asunEqual(_ a: AsunValue, _ b: AsunValue) -> Bool {
    switch (a, b) {
    case (.null, .null): return true
    case (.bool(let x), .bool(let y)): return x == y
    case (.string(let x), .string(let y)): return x == y
    case (.int(let x), .int(let y)): return x == y
    case (.float(let x), .float(let y)):
        if x == y { return true }
        let mag = max(abs(x), abs(y))
        let tol = max(1e-12, mag * 1e-12)
        return abs(x - y) <= tol
    case (.int(let x), .float(let y)):
        let xf = Double(x)
        if xf == y { return true }
        let mag = max(abs(xf), abs(y))
        return abs(xf - y) <= max(1e-12, mag * 1e-12)
    case (.float(let x), .int(let y)):
        let yf = Double(y)
        if x == yf { return true }
        let mag = max(abs(x), abs(yf))
        return abs(x - yf) <= max(1e-12, mag * 1e-12)
    case (.array(let x), .array(let y)):
        guard x.count == y.count else { return false }
        for i in x.indices {
            if !asunEqual(x[i], y[i]) { return false }
        }
        return true
    case (.object(let x), .object(let y)):
        guard x.count == y.count else { return false }
        for (k, v) in x {
            guard let w = y[k] else { return false }
            if !asunEqual(v, w) { return false }
        }
        return true
    default:
        return false
    }
}

// ---------- AsunValue → diagnostic string ----------

func diag(_ v: AsunValue) -> String {
    switch v {
    case .null: return "null"
    case .bool(let b): return b ? "true" : "false"
    case .int(let i): return String(i)
    case .float(let f): return String(f)
    case .string(let s): return "\"\(s)\""
    case .array(let arr):
        return "[" + arr.map { diag($0) }.joined(separator: ",") + "]"
    case .object(let obj):
        let keys = obj.keys.sorted()
        return "{" + keys.map { "\($0):\(diag(obj[$0]!))" }.joined(separator: ",") + "}"
    }
}

// ---------- Resolve conformance dir ----------

let runnerDir = URL(fileURLWithPath: CommandLine.arguments[0])
    .deletingLastPathComponent()
let conformanceDir: URL = {
    // Allow override via ASUN_CONFORMANCE_DIR for `swift run` from any cwd
    if let env = ProcessInfo.processInfo.environment["ASUN_CONFORMANCE_DIR"] {
        return URL(fileURLWithPath: env)
    }
    // Default: walk up from runner dir / cwd to find conformance/cases.json
    let candidates: [URL] = [
        URL(fileURLWithPath: FileManager.default.currentDirectoryPath),
        runnerDir,
    ]
    for base in candidates {
        var cur = base
        for _ in 0..<8 {
            let cases = cur.appendingPathComponent("cases.json")
            if FileManager.default.fileExists(atPath: cases.path) {
                return cur
            }
            let nested = cur.appendingPathComponent("conformance/cases.json")
            if FileManager.default.fileExists(atPath: nested.path) {
                return cur.appendingPathComponent("conformance")
            }
            cur = cur.deletingLastPathComponent()
        }
    }
    // Fallback to CWD/../..
    return URL(fileURLWithPath: FileManager.default.currentDirectoryPath)
        .deletingLastPathComponent().deletingLastPathComponent()
}()

let casesPath = conformanceDir.appendingPathComponent("cases.json")
let encPath = conformanceDir.appendingPathComponent("encode-cases.json")

// ---------- Decode cases ----------

guard let casesData = try? Data(contentsOf: casesPath) else {
    FileHandle.standardError.write("cannot read \(casesPath.path)\n".data(using: .utf8)!)
    exit(2)
}

guard let casesJson = try? JSONSerialization.jsonObject(with: casesData) as? [String: Any],
      let casesArr = casesJson["cases"] as? [[String: Any]] else {
    FileHandle.standardError.write("cannot parse cases.json\n".data(using: .utf8)!)
    exit(2)
}

print("loaded \(casesArr.count) cases from \(casesPath.path)")

var dTotal = 0, dOkPass = 0, dOkFail = 0, dErrPass = 0, dErrFail = 0, dSkipped = 0
var dFailures: [(String, String)] = []

for c in casesArr {
    dTotal += 1
    let id = c["id"] as? String ?? "?"
    let schemaDriven = c["schemaDriven"] as? Bool ?? false
    if schemaDriven {
        dSkipped += 1
        continue
    }
    let input = c["input"] as? String ?? ""
    let kind = c["kind"] as? String ?? "ok"

    do {
        let got = try decode(input)
        if kind == "ok" {
            let expected: AsunValue
            if let exp = c["expected"] {
                expected = jsonToAsun(exp)
            } else {
                expected = .null
            }
            if asunEqual(got, expected) {
                dOkPass += 1
            } else {
                dOkFail += 1
                if dFailures.count < 25 {
                    dFailures.append((id,
                        "value mismatch\n    input:    \"\(input)\"\n    expected: \(diag(expected))\n    actual:   \(diag(got))"))
                }
            }
        } else {
            dErrFail += 1
            if dFailures.count < 25 {
                dFailures.append((id,
                    "expected error, got ok: \(diag(got))\n    input: \"\(input)\""))
            }
        }
    } catch {
        if kind == "error" {
            dErrPass += 1
        } else {
            dOkFail += 1
            if dFailures.count < 25 {
                dFailures.append((id,
                    "expected ok, got error: \(error)\n    input: \"\(input)\""))
            }
        }
    }
}

let dExecuted = dTotal - dSkipped
let dPct: Double = dExecuted > 0
    ? 100.0 * Double(dOkPass + dErrPass) / Double(dExecuted)
    : 0.0

print("")
print("================ ASUN-SWIFT conformance ================")
print("total                : \(dTotal)")
print("untyped ok-cases pass: \(dOkPass)")
print("untyped ok-cases fail: \(dOkFail)")
print("untyped err-cases pass: \(dErrPass)")
print("untyped err-cases fail: \(dErrFail)")
print("skipped (needs typed): \(dSkipped)")
print(String(format: "untyped pass rate    : %d/%d (%.1f%%)",
             dOkPass + dErrPass, dExecuted, dPct))
print("========================================================")
for (id, msg) in dFailures {
    print("")
    print("[\(id)]")
    print("    \(msg)")
}

// ---------- Encode (round-trip) cases ----------

var ePass = 0, eFail = 0, eTotal = 0
var eFailures: [(String, String)] = []

if let encData = try? Data(contentsOf: encPath),
   let encJson = try? JSONSerialization.jsonObject(with: encData) as? [String: Any],
   let encArr = encJson["cases"] as? [[String: Any]] {
    eTotal = encArr.count
    print("loaded \(encArr.count) encode cases from \(encPath.path)")

    for c in encArr {
        let id = c["id"] as? String ?? "?"
        let value = jsonToAsun(c["value"] ?? NSNull())

        let encoded: String
        do {
            encoded = try encode(value)
        } catch {
            eFail += 1
            if eFailures.count < 25 {
                eFailures.append((id, "encode error: \(error)\n    value: \(diag(value))"))
            }
            continue
        }

        let decoded: AsunValue
        do {
            decoded = try decode(encoded)
        } catch {
            eFail += 1
            if eFailures.count < 25 {
                eFailures.append((id,
                    "decode error after encode: \(error)\n    encoded: \"\(encoded)\""))
            }
            continue
        }

        if asunEqual(decoded, value) {
            ePass += 1
        } else {
            eFail += 1
            if eFailures.count < 25 {
                eFailures.append((id,
                    "round-trip mismatch\n    encoded:  \"\(encoded)\"\n    expected: \(diag(value))\n    actual:   \(diag(decoded))"))
            }
        }
    }

    let ePct: Double = eTotal > 0
        ? 100.0 * Double(ePass) / Double(eTotal)
        : 0.0
    print("")
    print("============ ASUN-SWIFT encode round-trip ==============")
    print("total : \(eTotal)")
    print("pass  : \(ePass)")
    print("fail  : \(eFail)")
    print(String(format: "rate  : %d/%d (%.1f%%)", ePass, eTotal, ePct))
    print("========================================================")
    for (id, msg) in eFailures {
        print("")
        print("[\(id)]")
        print("    \(msg)")
    }
}

if dOkFail > 0 || dErrFail > 0 || eFail > 0 {
    exit(1)
}
