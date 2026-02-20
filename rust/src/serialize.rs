use crate::error::{Error, Result};
use serde::ser::{self, Serialize};

// ---------------------------------------------------------------------------
// Lookup tables
// ---------------------------------------------------------------------------

/// Bytes that force a string to be wrapped in `"..."`.
static NEEDS_QUOTE: [bool; 256] = {
    let mut t = [false; 256];
    let mut i = 0usize;
    while i < 32 {
        t[i] = true;
        i += 1;
    }
    t[b',' as usize] = true;
    t[b'(' as usize] = true;
    t[b')' as usize] = true;
    t[b'[' as usize] = true;
    t[b']' as usize] = true;
    t[b'"' as usize] = true;
    t[b'\\' as usize] = true;
    t
};

/// Escape replacement: `ESCAPE[b]` is the char after `\` for byte `b` inside
/// a quoted string.  0 means the byte needs no escaping.
static ESCAPE: [u8; 256] = {
    let mut t = [0u8; 256];
    t[b'"' as usize] = b'"';
    t[b'\\' as usize] = b'\\';
    t[b'\n' as usize] = b'n';
    t[b'\t' as usize] = b't';
    t
};

/// Two-digit lookup table for fast integer formatting (itoa-style).
static DEC_DIGITS: &[u8; 200] = b"0001020304050607080910111213141516171819\
2021222324252627282930313233343536373839\
4041424344454647484950515253545556575859\
6061626364656667686970717273747576777879\
8081828384858687888990919293949596979899";

// ---------------------------------------------------------------------------
// Stack-based number formatting (no heap allocation)
// ---------------------------------------------------------------------------

#[inline(always)]
fn write_u64(buf: &mut Vec<u8>, mut v: u64) {
    if v < 10 {
        buf.push(b'0' + v as u8);
        return;
    }
    if v < 100 {
        let idx = v as usize * 2;
        buf.push(DEC_DIGITS[idx]);
        buf.push(DEC_DIGITS[idx + 1]);
        return;
    }
    let mut tmp = [0u8; 20];
    let mut i = 20usize;
    while v >= 100 {
        let rem = (v % 100) as usize;
        v /= 100;
        i -= 2;
        tmp[i] = DEC_DIGITS[rem * 2];
        tmp[i + 1] = DEC_DIGITS[rem * 2 + 1];
    }
    if v >= 10 {
        let idx = v as usize * 2;
        i -= 2;
        tmp[i] = DEC_DIGITS[idx];
        tmp[i + 1] = DEC_DIGITS[idx + 1];
    } else {
        i -= 1;
        tmp[i] = b'0' + v as u8;
    }
    buf.extend_from_slice(&tmp[i..]);
}

#[inline(always)]
fn write_i64(buf: &mut Vec<u8>, v: i64) {
    if v < 0 {
        buf.push(b'-');
        write_u64(buf, (-(v as i128)) as u64);
    } else {
        write_u64(buf, v as u64);
    }
}

/// Write f64 to buffer.
/// - Integer-valued floats: fast path via write_i64 + ".0"
/// - One-decimal floats (e.g. 50.5): fast path via integer arithmetic
/// - General: core::fmt::Display directly into buffer
#[inline]
fn write_f64(buf: &mut Vec<u8>, v: f64) {
    if v.is_finite() && v.fract() == 0.0 {
        if v >= i64::MIN as f64 && v <= i64::MAX as f64 {
            write_i64(buf, v as i64);
            buf.extend_from_slice(b".0");
        } else {
            fmt_f64(buf, v);
        }
        return;
    }
    if v.is_finite() {
        // Fast path: one decimal place (covers xx.5, xx.1, etc.)
        let v10 = v * 10.0;
        if v10.fract() == 0.0 && v10.abs() < 1e18 {
            let vi = v10 as i64;
            let (int_part, frac) = if vi < 0 {
                buf.push(b'-');
                let pos = (-vi) as u64;
                ((pos / 10), (pos % 10) as u8)
            } else {
                let pos = vi as u64;
                ((pos / 10), (pos % 10) as u8)
            };
            write_u64(buf, int_part);
            buf.push(b'.');
            buf.push(b'0' + frac);
            return;
        }
        // Fast path: two decimal places (covers xx.25, xx.75, etc.)
        let v100 = v * 100.0;
        if v100.fract() == 0.0 && v100.abs() < 1e18 {
            let vi = v100 as i64;
            let (int_part, frac) = if vi < 0 {
                buf.push(b'-');
                let pos = (-vi) as u64;
                ((pos / 100), (pos % 100) as usize)
            } else {
                let pos = vi as u64;
                ((pos / 100), (pos % 100) as usize)
            };
            write_u64(buf, int_part);
            buf.push(b'.');
            buf.push(DEC_DIGITS[frac * 2]);
            let d2 = DEC_DIGITS[frac * 2 + 1];
            if d2 != b'0' {
                buf.push(d2);
            }
            return;
        }
    }
    fmt_f64(buf, v);
}

/// Fallback: format f64 using core::fmt::Display directly into the Vec.
#[inline(never)]
fn fmt_f64(buf: &mut Vec<u8>, v: f64) {
    use core::fmt::Write;
    struct W<'a>(&'a mut Vec<u8>);
    impl core::fmt::Write for W<'_> {
        #[inline]
        fn write_str(&mut self, s: &str) -> core::fmt::Result {
            self.0.extend_from_slice(s.as_bytes());
            Ok(())
        }
    }
    let _ = write!(W(buf), "{}", v);
}

// ---------------------------------------------------------------------------
// String quoting / escaping
// ---------------------------------------------------------------------------

/// Single-pass check: does `s` need to be wrapped in quotes?
#[inline]
fn needs_quoting(s: &str) -> bool {
    let bytes = s.as_bytes();
    if bytes.is_empty() {
        return true;
    }
    if bytes[0] == b' ' || bytes[bytes.len() - 1] == b' ' {
        return true;
    }
    if (bytes.len() == 4 && bytes == b"true") || (bytes.len() == 5 && bytes == b"false") {
        return true;
    }
    // Single pass: detect special chars AND number-like patterns
    let mut could_be_number = true;
    let num_start = if bytes[0] == b'-' { 1 } else { 0 };
    if num_start >= bytes.len() {
        could_be_number = false;
    }
    for i in 0..bytes.len() {
        let b = bytes[i];
        if NEEDS_QUOTE[b as usize] {
            return true;
        }
        if could_be_number && i >= num_start && !b.is_ascii_digit() && b != b'.' {
            could_be_number = false;
        }
    }
    could_be_number && bytes.len() > num_start
}

/// Write `s` wrapped in quotes with escaping. Bulk writes non-escaped runs.
#[inline]
fn write_escaped(buf: &mut Vec<u8>, s: &str) {
    buf.push(b'"');
    let bytes = s.as_bytes();
    let mut start = 0;
    for i in 0..bytes.len() {
        let esc = ESCAPE[bytes[i] as usize];
        if esc != 0 {
            if start < i {
                buf.extend_from_slice(&bytes[start..i]);
            }
            buf.push(b'\\');
            buf.push(esc);
            start = i + 1;
        }
    }
    if start < bytes.len() {
        buf.extend_from_slice(&bytes[start..]);
    }
    buf.push(b'"');
}

// ---------------------------------------------------------------------------
// Serializer
// ---------------------------------------------------------------------------

pub struct Serializer {
    pub(crate) buf: Vec<u8>,
    in_tuple: bool,
    first: bool,
    /// When true, record type hints for top-level struct fields.
    typed: bool,
    /// Accumulates type hint for the current field being serialized.
    current_type_hint: Option<&'static str>,
}

pub fn to_string<T: Serialize>(value: &T) -> Result<String> {
    let mut serializer = Serializer {
        buf: Vec::with_capacity(256),
        in_tuple: false,
        first: true,
        typed: false,
        current_type_hint: None,
    };
    value.serialize(&mut serializer)?;
    Ok(unsafe { String::from_utf8_unchecked(serializer.buf) })
}

/// Serialize a single struct to ASON string with type-annotated schema.
///
/// Output example: `{id:int,name:str,active:bool}:(1,Alice,true)`
pub fn to_string_typed<T: Serialize>(value: &T) -> Result<String> {
    let mut serializer = Serializer {
        buf: Vec::with_capacity(256),
        in_tuple: false,
        first: true,
        typed: true,
        current_type_hint: None,
    };
    value.serialize(&mut serializer)?;
    Ok(unsafe { String::from_utf8_unchecked(serializer.buf) })
}

impl Serializer {
    #[inline(always)]
    fn push_separator(&mut self) {
        if !self.first {
            self.buf.push(b',');
        }
        self.first = false;
    }
}

impl<'a> ser::Serializer for &'a mut Serializer {
    type Ok = ();
    type Error = Error;

    type SerializeSeq = SeqSerializer<'a>;
    type SerializeTuple = TupleSerializer<'a>;
    type SerializeTupleStruct = TupleSerializer<'a>;
    type SerializeTupleVariant = TupleSerializer<'a>;
    type SerializeMap = MapSerializer<'a>;
    type SerializeStruct = StructSerializer<'a>;
    type SerializeStructVariant = StructSerializer<'a>;

    #[inline]
    fn serialize_bool(self, v: bool) -> Result<()> {
        self.push_separator();
        if self.current_type_hint.is_none() && self.typed {
            self.current_type_hint = Some("bool");
        }
        self.buf
            .extend_from_slice(if v { b"true" } else { b"false" });
        Ok(())
    }

    #[inline]
    fn serialize_i8(self, v: i8) -> Result<()> {
        self.serialize_i64(v as i64)
    }
    #[inline]
    fn serialize_i16(self, v: i16) -> Result<()> {
        self.serialize_i64(v as i64)
    }
    #[inline]
    fn serialize_i32(self, v: i32) -> Result<()> {
        self.serialize_i64(v as i64)
    }

    #[inline]
    fn serialize_i64(self, v: i64) -> Result<()> {
        self.push_separator();
        if self.current_type_hint.is_none() && self.typed {
            self.current_type_hint = Some("int");
        }
        write_i64(&mut self.buf, v);
        Ok(())
    }

    #[inline]
    fn serialize_u8(self, v: u8) -> Result<()> {
        self.serialize_u64(v as u64)
    }
    #[inline]
    fn serialize_u16(self, v: u16) -> Result<()> {
        self.serialize_u64(v as u64)
    }
    #[inline]
    fn serialize_u32(self, v: u32) -> Result<()> {
        self.serialize_u64(v as u64)
    }

    #[inline]
    fn serialize_u64(self, v: u64) -> Result<()> {
        self.push_separator();
        if self.current_type_hint.is_none() && self.typed {
            self.current_type_hint = Some("int");
        }
        write_u64(&mut self.buf, v);
        Ok(())
    }

    #[inline]
    fn serialize_f32(self, v: f32) -> Result<()> {
        self.serialize_f64(v as f64)
    }

    #[inline]
    fn serialize_f64(self, v: f64) -> Result<()> {
        self.push_separator();
        if self.current_type_hint.is_none() && self.typed {
            self.current_type_hint = Some("float");
        }
        write_f64(&mut self.buf, v);
        Ok(())
    }

    #[inline]
    fn serialize_char(self, v: char) -> Result<()> {
        self.push_separator();
        if self.current_type_hint.is_none() && self.typed {
            self.current_type_hint = Some("str");
        }
        let mut tmp = [0u8; 4];
        let s = v.encode_utf8(&mut tmp);
        self.buf.extend_from_slice(s.as_bytes());
        Ok(())
    }

    #[inline]
    fn serialize_str(self, v: &str) -> Result<()> {
        self.push_separator();
        if self.current_type_hint.is_none() && self.typed {
            self.current_type_hint = Some("str");
        }
        if needs_quoting(v) {
            write_escaped(&mut self.buf, v);
        } else {
            self.buf.extend_from_slice(v.as_bytes());
        }
        Ok(())
    }

    fn serialize_bytes(self, v: &[u8]) -> Result<()> {
        self.push_separator();
        self.buf.push(b'[');
        for (i, &b) in v.iter().enumerate() {
            if i > 0 {
                self.buf.push(b',');
            }
            write_u64(&mut self.buf, b as u64);
        }
        self.buf.push(b']');
        Ok(())
    }

    #[inline]
    fn serialize_none(self) -> Result<()> {
        self.push_separator();
        // For typed mode: None doesn't set a type hint (the Some branch will)
        Ok(())
    }

    #[inline]
    fn serialize_some<T: ?Sized + Serialize>(self, value: &T) -> Result<()> {
        value.serialize(self)
    }

    #[inline]
    fn serialize_unit(self) -> Result<()> {
        self.push_separator();
        self.buf.extend_from_slice(b"()");
        Ok(())
    }

    #[inline]
    fn serialize_unit_struct(self, _name: &'static str) -> Result<()> {
        self.serialize_unit()
    }

    fn serialize_unit_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        variant: &'static str,
    ) -> Result<()> {
        self.serialize_str(variant)
    }

    fn serialize_newtype_struct<T: ?Sized + Serialize>(
        self,
        _name: &'static str,
        value: &T,
    ) -> Result<()> {
        value.serialize(self)
    }

    fn serialize_newtype_variant<T: ?Sized + Serialize>(
        self,
        _name: &'static str,
        _variant_index: u32,
        variant: &'static str,
        value: &T,
    ) -> Result<()> {
        self.push_separator();
        self.buf.push(b'(');
        self.buf.extend_from_slice(variant.as_bytes());
        self.buf.push(b',');
        self.first = true;
        value.serialize(&mut *self)?;
        self.buf.push(b')');
        Ok(())
    }

    fn serialize_seq(self, _len: Option<usize>) -> Result<SeqSerializer<'a>> {
        self.push_separator();
        // For typed mode: mark as generic array if no specific hint set
        if self.current_type_hint.is_none() && self.typed {
            // We'll refine with element types if possible, but default to generic
            // The hint will stay None and StructSerializer will not emit a type
        }
        self.buf.push(b'[');
        Ok(SeqSerializer {
            ser: self,
            first: true,
        })
    }

    fn serialize_tuple(self, _len: usize) -> Result<TupleSerializer<'a>> {
        self.push_separator();
        self.buf.push(b'(');
        Ok(TupleSerializer {
            ser: self,
            first: true,
        })
    }

    fn serialize_tuple_struct(
        self,
        _name: &'static str,
        _len: usize,
    ) -> Result<TupleSerializer<'a>> {
        self.push_separator();
        self.buf.push(b'(');
        Ok(TupleSerializer {
            ser: self,
            first: true,
        })
    }

    fn serialize_tuple_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        variant: &'static str,
        _len: usize,
    ) -> Result<TupleSerializer<'a>> {
        self.push_separator();
        self.buf.push(b'(');
        self.buf.extend_from_slice(variant.as_bytes());
        Ok(TupleSerializer {
            ser: self,
            first: false,
        })
    }

    fn serialize_map(self, _len: Option<usize>) -> Result<MapSerializer<'a>> {
        self.push_separator();
        if self.current_type_hint.is_none() && self.typed {
            self.current_type_hint = Some("map");
        }
        self.buf.push(b'[');
        Ok(MapSerializer {
            ser: self,
            first: true,
        })
    }

    fn serialize_struct(self, _name: &'static str, len: usize) -> Result<StructSerializer<'a>> {
        let is_top = !self.in_tuple;
        if is_top {
            let data_start = self.buf.len();
            self.buf.push(b'(');
            self.in_tuple = true;
            Ok(StructSerializer {
                ser: self,
                fields: Vec::with_capacity(len),
                field_types: Vec::with_capacity(len),
                is_top: true,
                first: true,
                data_start,
            })
        } else {
            self.push_separator();
            self.buf.push(b'(');
            Ok(StructSerializer {
                ser: self,
                fields: Vec::new(),
                field_types: Vec::new(),
                is_top: false,
                first: true,
                data_start: 0,
            })
        }
    }

    fn serialize_struct_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        variant: &'static str,
        _len: usize,
    ) -> Result<StructSerializer<'a>> {
        self.push_separator();
        self.buf.push(b'(');
        self.buf.extend_from_slice(variant.as_bytes());
        self.buf.push(b',');
        Ok(StructSerializer {
            ser: self,
            fields: Vec::new(),
            field_types: Vec::new(),
            is_top: false,
            first: true,
            data_start: 0,
        })
    }
}

// ---------------------------------------------------------------------------
// SeqSerializer
// ---------------------------------------------------------------------------

pub struct SeqSerializer<'a> {
    ser: &'a mut Serializer,
    first: bool,
}

impl<'a> ser::SerializeSeq for SeqSerializer<'a> {
    type Ok = ();
    type Error = Error;

    #[inline]
    fn serialize_element<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<()> {
        if !self.first {
            self.ser.buf.push(b',');
        }
        self.first = false;
        self.ser.first = true;
        value.serialize(&mut *self.ser)
    }

    #[inline]
    fn end(self) -> Result<()> {
        self.ser.buf.push(b']');
        self.ser.first = false;
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// TupleSerializer
// ---------------------------------------------------------------------------

pub struct TupleSerializer<'a> {
    ser: &'a mut Serializer,
    first: bool,
}

impl<'a> ser::SerializeTuple for TupleSerializer<'a> {
    type Ok = ();
    type Error = Error;

    #[inline]
    fn serialize_element<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<()> {
        if !self.first {
            self.ser.buf.push(b',');
        }
        self.first = false;
        self.ser.first = true;
        value.serialize(&mut *self.ser)
    }

    #[inline]
    fn end(self) -> Result<()> {
        self.ser.buf.push(b')');
        self.ser.first = false;
        Ok(())
    }
}

impl<'a> ser::SerializeTupleStruct for TupleSerializer<'a> {
    type Ok = ();
    type Error = Error;

    #[inline]
    fn serialize_field<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<()> {
        ser::SerializeTuple::serialize_element(self, value)
    }

    #[inline]
    fn end(self) -> Result<()> {
        ser::SerializeTuple::end(self)
    }
}

impl<'a> ser::SerializeTupleVariant for TupleSerializer<'a> {
    type Ok = ();
    type Error = Error;

    #[inline]
    fn serialize_field<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<()> {
        ser::SerializeTuple::serialize_element(self, value)
    }

    #[inline]
    fn end(self) -> Result<()> {
        ser::SerializeTuple::end(self)
    }
}

// ---------------------------------------------------------------------------
// MapSerializer
// ---------------------------------------------------------------------------

pub struct MapSerializer<'a> {
    ser: &'a mut Serializer,
    first: bool,
}

impl<'a> ser::SerializeMap for MapSerializer<'a> {
    type Ok = ();
    type Error = Error;

    fn serialize_key<T: ?Sized + Serialize>(&mut self, key: &T) -> Result<()> {
        if !self.first {
            self.ser.buf.push(b',');
        }
        self.first = false;
        self.ser.buf.push(b'(');
        self.ser.first = true;
        key.serialize(&mut *self.ser)
    }

    fn serialize_value<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<()> {
        self.ser.buf.push(b',');
        self.ser.first = true;
        value.serialize(&mut *self.ser)?;
        self.ser.buf.push(b')');
        Ok(())
    }

    fn end(self) -> Result<()> {
        self.ser.buf.push(b']');
        self.ser.first = false;
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// StructSerializer
// ---------------------------------------------------------------------------

pub struct StructSerializer<'a> {
    ser: &'a mut Serializer,
    fields: Vec<&'static str>,
    /// Type hints collected for each field (only when typed mode is on)
    field_types: Vec<Option<&'static str>>,
    is_top: bool,
    first: bool,
    data_start: usize,
}

impl<'a> ser::SerializeStruct for StructSerializer<'a> {
    type Ok = ();
    type Error = Error;

    #[inline]
    fn serialize_field<T: ?Sized + Serialize>(
        &mut self,
        key: &'static str,
        value: &T,
    ) -> Result<()> {
        if self.is_top {
            self.fields.push(key);
            // Clear the type hint slot before serializing the value
            if self.ser.typed {
                self.ser.current_type_hint = None;
            }
        }
        if !self.first {
            self.ser.buf.push(b',');
        }
        self.first = false;
        self.ser.first = true;
        self.ser.in_tuple = true;
        value.serialize(&mut *self.ser)?;
        if self.is_top && self.ser.typed {
            self.field_types.push(self.ser.current_type_hint.take());
        }
        Ok(())
    }

    fn end(self) -> Result<()> {
        if self.is_top {
            self.ser.buf.push(b')');
            // Split data, prepend schema, re-append
            let data = self.ser.buf.split_off(self.data_start);
            self.ser.buf.push(b'{');
            for (i, f) in self.fields.iter().enumerate() {
                if i > 0 {
                    self.ser.buf.push(b',');
                }
                self.ser.buf.extend_from_slice(f.as_bytes());
                // Append type hint if available (typed mode)
                if self.ser.typed {
                    if let Some(type_hint) = self.field_types.get(i).and_then(|t| *t) {
                        self.ser.buf.push(b':');
                        self.ser.buf.extend_from_slice(type_hint.as_bytes());
                    }
                }
            }
            self.ser.buf.extend_from_slice(b"}:");
            self.ser.buf.extend_from_slice(&data);
        } else {
            self.ser.buf.push(b')');
            self.ser.first = false;
            // Clear any leaked type hint from nested struct fields
            if self.ser.typed {
                self.ser.current_type_hint = None;
            }
        }
        Ok(())
    }
}

impl<'a> ser::SerializeStructVariant for StructSerializer<'a> {
    type Ok = ();
    type Error = Error;

    #[inline]
    fn serialize_field<T: ?Sized + Serialize>(
        &mut self,
        _key: &'static str,
        value: &T,
    ) -> Result<()> {
        if !self.first {
            self.ser.buf.push(b',');
        }
        self.first = false;
        self.ser.first = true;
        value.serialize(&mut *self.ser)
    }

    fn end(self) -> Result<()> {
        self.ser.buf.push(b')');
        self.ser.first = false;
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// Vec<T: StructSchema> — direct serialization, no per-row allocation
// ---------------------------------------------------------------------------

pub fn to_string_vec<T: Serialize>(values: &[T]) -> Result<String>
where
    T: StructSchema,
{
    to_string_vec_inner(values, false)
}

/// Serialize a Vec of structs to ASON string with type-annotated schema.
///
/// Output example: `{id:int,name:str,active:bool}:(1,Alice,true),(2,Bob,false)`
///
/// Requires `StructSchema` to be implemented with `field_types()`.
pub fn to_string_vec_typed<T: Serialize>(values: &[T]) -> Result<String>
where
    T: StructSchema,
{
    to_string_vec_inner(values, true)
}

fn to_string_vec_inner<T: Serialize>(values: &[T], typed: bool) -> Result<String>
where
    T: StructSchema,
{
    let mut ser = Serializer {
        buf: Vec::with_capacity(values.len() * 48 + 64),
        in_tuple: true,
        first: true,
        typed,
        current_type_hint: None,
    };

    // Write schema header
    ser.buf.push(b'{');
    let fields = T::field_names();
    let types = if typed { Some(T::field_types()) } else { None };
    for (i, f) in fields.iter().enumerate() {
        if i > 0 {
            ser.buf.push(b',');
        }
        ser.buf.extend_from_slice(f.as_bytes());
        if let Some(ref type_list) = types {
            if let Some(type_hint) = type_list.get(i).copied() {
                if !type_hint.is_empty() {
                    ser.buf.push(b':');
                    ser.buf.extend_from_slice(type_hint.as_bytes());
                }
            }
        }
    }
    ser.buf.extend_from_slice(b"}:");

    // Write data rows directly — no per-row allocation
    for (i, val) in values.iter().enumerate() {
        if i > 0 {
            ser.buf.push(b',');
        }
        ser.buf.push(b'(');
        ser.first = true;
        val.serialize_fields(&mut ser)?;
        ser.buf.push(b')');
    }

    Ok(unsafe { String::from_utf8_unchecked(ser.buf) })
}

/// Trait for structs to provide their schema for optimized vec serialization.
pub trait StructSchema {
    fn field_names() -> &'static [&'static str];
    fn serialize_fields(&self, ser: &mut Serializer) -> Result<()>;

    /// Return type annotations for each field. Used by `to_string_vec_typed`.
    ///
    /// Each entry should be an ASON type hint string (e.g. `"int"`, `"str"`,
    /// `"float"`, `"bool"`), or `""` to omit the type for that field.
    ///
    /// Default implementation returns empty (no types), which means
    /// `to_string_vec_typed` will fall back to unannotated schema for
    /// fields without explicit types.
    fn field_types() -> &'static [&'static str] {
        &[]
    }
}
