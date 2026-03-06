using System.Runtime.CompilerServices;

namespace Ason;

/// <summary>
/// High-performance ASON text encoder. Uses WriteValues for zero-boxing encode.
/// </summary>
public static class Encoder
{
    public static string Encode(IAsonSchema value)
    {
        var w = new AsonWriter(256);
        try { EncodeStruct(ref w, value, false); return w.ToString(); }
        finally { w.Dispose(); }
    }

    public static string Encode<T>(IReadOnlyList<T> values) where T : IAsonSchema
    {
        int cap = values.Count * 64 + 128;
        var w = new AsonWriter(cap);
        try { EncodeTopList(ref w, values, false); return w.ToString(); }
        finally { w.Dispose(); }
    }

    public static string EncodeTyped(IAsonSchema value)
    {
        var w = new AsonWriter(256);
        try { EncodeStruct(ref w, value, true); return w.ToString(); }
        finally { w.Dispose(); }
    }

    public static string EncodeTyped<T>(IReadOnlyList<T> values) where T : IAsonSchema
    {
        int cap = values.Count * 64 + 128;
        var w = new AsonWriter(cap);
        try { EncodeTopList(ref w, values, true); return w.ToString(); }
        finally { w.Dispose(); }
    }

    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    internal static void EncodeStruct(ref AsonWriter w, IAsonSchema obj, bool typed)
    {
        WriteSchemaHeader(ref w, obj, typed);
        w.WriteSpan("}:");
        w.WriteChar('(');
        obj.WriteValues(ref w);
        w.WriteChar(')');
    }

    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    private static void WriteSchemaHeader(ref AsonWriter w, IAsonSchema obj, bool typed)
    {
        var names = obj.FieldNames;
        var types = typed ? obj.FieldTypes : ReadOnlySpan<string?>.Empty;
        var values = obj.FieldValues; // only for schema detection (nested types)

        w.WriteChar('{');
        for (int i = 0; i < names.Length; i++)
        {
            if (i > 0) w.WriteChar(',');
            w.WriteSpan(names[i]);

            // Check for nested schema types
            var v = i < values.Length ? values[i] : null;
            if (v is IAsonSchema nested)
            {
                w.WriteChar(':');
                WriteNestedSchemaHeader(ref w, nested, typed);
            }
            else if (v is System.Collections.IList list && list.Count > 0 && list[0] is IAsonSchema firstSchema)
            {
                w.WriteSpan(":[");
                WriteNestedSchemaHeader(ref w, firstSchema, typed);
                w.WriteChar(']');
            }
            else if (typed && i < types.Length && types[i] != null)
            {
                w.WriteChar(':');
                w.WriteSpan(types[i]!);
            }
        }
    }

    private static void WriteNestedSchemaHeader(ref AsonWriter w, IAsonSchema obj, bool typed)
    {
        var names = obj.FieldNames;
        var types = typed ? obj.FieldTypes : ReadOnlySpan<string?>.Empty;
        var values = obj.FieldValues;

        w.WriteChar('{');
        for (int i = 0; i < names.Length; i++)
        {
            if (i > 0) w.WriteChar(',');
            w.WriteSpan(names[i]);

            var v = i < values.Length ? values[i] : null;
            if (v is IAsonSchema nested)
            {
                w.WriteChar(':');
                WriteNestedSchemaHeader(ref w, nested, typed);
            }
            else if (typed && i < types.Length && types[i] != null)
            {
                w.WriteChar(':');
                w.WriteSpan(types[i]!);
            }
        }
        w.WriteChar('}');
    }

    internal static void EncodeTopList<T>(ref AsonWriter w, IReadOnlyList<T> list, bool typed) where T : IAsonSchema
    {
        if (list.Count == 0) { w.WriteSpan("[]"); return; }
        var first = list[0];
        w.WriteSpan("[");
        WriteNestedSchemaHeader(ref w, first, typed);
        w.WriteSpan("]:");
        for (int r = 0; r < list.Count; r++)
        {
            if (r > 0) w.WriteChar(',');
            w.WriteChar('(');
            list[r].WriteValues(ref w);
            w.WriteChar(')');
        }
    }
}
