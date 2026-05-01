// Conformance runner for asun-cs.
//
// Loads conformance/cases.json (untyped decode) and conformance/encode-cases.json
// (round-trip), driving each case through Asun.decodeValue / Asun.encodeValue.
//
// Output format mirrors the cpp / c / java runners so the cross-language
// dashboard can grep the same lines.

using System.Globalization;
using System.Text.Json;
using Asun;

namespace AsunConformance;

internal static class Program
{
    private static AsunValue JsonToAsun(JsonElement el) => el.ValueKind switch
    {
        JsonValueKind.Null => AsunValue.Null,
        JsonValueKind.True => AsunValue.Of(true),
        JsonValueKind.False => AsunValue.Of(false),
        JsonValueKind.String => AsunValue.Of(el.GetString() ?? string.Empty),
        JsonValueKind.Number =>
            el.TryGetInt64(out var i) ? AsunValue.Of(i) : AsunValue.Of(el.GetDouble()),
        JsonValueKind.Array => AsunValue.Of(ToList(el)),
        JsonValueKind.Object => AsunValue.Null, // not expected
        _ => AsunValue.Null,
    };

    private static IReadOnlyList<AsunValue> ToList(JsonElement arr)
    {
        var items = new List<AsunValue>(arr.GetArrayLength());
        foreach (var e in arr.EnumerateArray()) items.Add(JsonToAsun(e));
        return items;
    }

    public static int Main()
    {
        Console.OutputEncoding = System.Text.Encoding.UTF8;

        // ------- decode (cases.json, untyped) -------------------------------
        string decodeText = File.ReadAllText("../../cases.json");
        using var decodeDoc = JsonDocument.Parse(decodeText);
        var dcases = decodeDoc.RootElement.GetProperty("cases");
        Console.WriteLine($"loaded {dcases.GetArrayLength()} cases from conformance/cases.json");

        int dTotal = 0, dOkPass = 0, dOkFail = 0, dErrPass = 0, dErrFail = 0, dSkipped = 0;
        var dFailures = new List<(string id, string msg)>();

        foreach (var c in dcases.EnumerateArray())
        {
            dTotal++;
            string id = c.GetProperty("id").GetString()!;
            if (c.TryGetProperty("schemaDriven", out var sd) && sd.ValueKind == JsonValueKind.True)
            {
                dSkipped++;
                continue;
            }
            string input = c.GetProperty("input").GetString()!;
            string kind = c.GetProperty("kind").GetString()!;

            AsunValue got;
            try { got = Asun.Asun.decodeValue(input); }
            catch (Exception e)
            {
                if (kind == "error") dErrPass++;
                else
                {
                    dOkFail++;
                    if (dFailures.Count < 25)
                        dFailures.Add((id, $"expected ok, got error: {e.Message}\n    input: {input}"));
                }
                continue;
            }

            if (kind == "ok")
            {
                var expected = JsonToAsun(c.GetProperty("expected"));
                if (got.Equals(expected)) dOkPass++;
                else
                {
                    dOkFail++;
                    if (dFailures.Count < 25)
                        dFailures.Add((id,
                            $"value mismatch\n    input:    {input}\n    expected: {expected.ToDiagnostic()}\n    actual:   {got.ToDiagnostic()}"));
                }
            }
            else
            {
                dErrFail++;
                if (dFailures.Count < 25)
                    dFailures.Add((id, $"expected error, got ok: {got.ToDiagnostic()}\n    input: {input}"));
            }
        }

        int dExecuted = dTotal - dSkipped;
        double dPct = dExecuted > 0 ? 100.0 * (dOkPass + dErrPass) / dExecuted : 0.0;
        Console.WriteLine();
        Console.WriteLine("================ ASUN-CS conformance ================");
        Console.WriteLine($"total                : {dTotal}");
        Console.WriteLine($"untyped ok-cases pass: {dOkPass}");
        Console.WriteLine($"untyped ok-cases fail: {dOkFail}");
        Console.WriteLine($"untyped err-cases pass: {dErrPass}");
        Console.WriteLine($"untyped err-cases fail: {dErrFail}");
        Console.WriteLine($"skipped (needs typed): {dSkipped}");
        Console.WriteLine(string.Format(CultureInfo.InvariantCulture,
            "untyped pass rate    : {0}/{1} ({2:F1}%)", dOkPass + dErrPass, dExecuted, dPct));
        Console.WriteLine("=====================================================");
        foreach (var (id, msg) in dFailures) Console.WriteLine($"\n[{id}]\n    {msg}");

        // ------- encode (round-trip) ---------------------------------------
        string encText = File.ReadAllText("../../encode-cases.json");
        using var encDoc = JsonDocument.Parse(encText);
        var ecases = encDoc.RootElement.GetProperty("cases");
        Console.WriteLine($"loaded {ecases.GetArrayLength()} encode cases from conformance/encode-cases.json");

        int ePass = 0, eFail = 0;
        var eFailures = new List<(string id, string msg)>();

        foreach (var c in ecases.EnumerateArray())
        {
            string id = c.GetProperty("id").GetString()!;
            var value = JsonToAsun(c.GetProperty("value"));

            string encoded;
            try { encoded = Asun.Asun.encodeValue(value); }
            catch (Exception e)
            {
                eFail++;
                if (eFailures.Count < 25)
                    eFailures.Add((id, $"encode error: {e.Message}"));
                continue;
            }

            AsunValue decoded;
            try { decoded = Asun.Asun.decodeValue(encoded); }
            catch (Exception e)
            {
                eFail++;
                if (eFailures.Count < 25)
                    eFailures.Add((id, $"decode error after encode: {e.Message}\n    encoded: {encoded}"));
                continue;
            }

            if (decoded.Equals(value)) ePass++;
            else
            {
                eFail++;
                if (eFailures.Count < 25)
                    eFailures.Add((id,
                        $"round-trip mismatch\n    encoded:  {encoded}\n    expected: {value.ToDiagnostic()}\n    actual:   {decoded.ToDiagnostic()}"));
            }
        }

        int eTotal = ecases.GetArrayLength();
        double ePct = eTotal > 0 ? 100.0 * ePass / eTotal : 0.0;
        Console.WriteLine();
        Console.WriteLine("================ ASUN-CS encode conformance ================");
        Console.WriteLine($"total : {eTotal}");
        Console.WriteLine($"pass  : {ePass}");
        Console.WriteLine($"fail  : {eFail}");
        Console.WriteLine(string.Format(CultureInfo.InvariantCulture,
            "rate  : {0}/{1} ({2:F1}%)", ePass, eTotal, ePct));
        Console.WriteLine("============================================================");
        foreach (var (id, msg) in eFailures) Console.WriteLine($"\n[{id}]\n    {msg}");

        return (dOkFail | dErrFail | eFail) == 0 ? 0 : 1;
    }
}
