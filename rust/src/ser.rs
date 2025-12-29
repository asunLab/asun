//! ASON Serializer implementation.

use crate::error::{Error, Result};
use serde::ser::{self, Serialize};

/// Serialize a value to an ASON string with schema.
pub fn to_string<T: Serialize>(value: &T) -> Result<String> {
    let mut serializer = Serializer::new();
    value.serialize(&mut serializer)?;
    Ok(serializer.output)
}

/// Serialize a value to an ASON string with schema and indentation.
pub fn to_string_pretty<T: Serialize>(value: &T) -> Result<String> {
    let mut serializer = Serializer::pretty();
    value.serialize(&mut serializer)?;
    Ok(serializer.output)
}

/// ASON Serializer.
pub struct Serializer {
    output: String,
    #[allow(dead_code)]
    schema: String,
    indent: Option<String>,
    depth: usize,
    #[allow(dead_code)]
    in_struct: bool,
    #[allow(dead_code)]
    field_index: usize,
}

impl Serializer {
    pub fn new() -> Self {
        Serializer {
            output: String::new(),
            schema: String::new(),
            indent: None,
            depth: 0,
            in_struct: false,
            field_index: 0,
        }
    }

    pub fn pretty() -> Self {
        Serializer {
            output: String::new(),
            schema: String::new(),
            indent: Some("  ".to_string()),
            depth: 0,
            in_struct: false,
            field_index: 0,
        }
    }

    #[allow(dead_code)]
    fn write_indent(&mut self) {
        if let Some(ref indent) = self.indent {
            self.output.push('\n');
            for _ in 0..self.depth {
                self.output.push_str(indent);
            }
        }
    }

    fn needs_quote(s: &str) -> bool {
        if s.is_empty() {
            return true;
        }
        if s == "null" || s == "true" || s == "false" {
            return true;
        }
        let first = s.chars().next().unwrap();
        if first == '-' || first == '+' || first.is_ascii_digit() {
            return true;
        }
        s.chars().any(|c| {
            c == '"'
                || c == '\\'
                || c == '('
                || c == ')'
                || c == '['
                || c == ']'
                || c == '{'
                || c == '}'
                || c == ','
                || c == ':'
                || c.is_whitespace()
        })
    }

    fn write_string(&mut self, s: &str) {
        if Self::needs_quote(s) {
            self.output.push('"');
            for c in s.chars() {
                match c {
                    '"' => self.output.push_str("\\\""),
                    '\\' => self.output.push_str("\\\\"),
                    '\n' => self.output.push_str("\\n"),
                    '\r' => self.output.push_str("\\r"),
                    '\t' => self.output.push_str("\\t"),
                    c if c < '\x20' => {
                        self.output.push_str(&format!("\\u{:04x}", c as u32));
                    }
                    c => self.output.push(c),
                }
            }
            self.output.push('"');
        } else {
            self.output.push_str(s);
        }
    }
}

impl<'a> ser::Serializer for &'a mut Serializer {
    type Ok = ();
    type Error = Error;
    type SerializeSeq = SeqSerializer<'a>;
    type SerializeTuple = SeqSerializer<'a>;
    type SerializeTupleStruct = SeqSerializer<'a>;
    type SerializeTupleVariant = SeqSerializer<'a>;
    type SerializeMap = MapSerializer<'a>;
    type SerializeStruct = StructSerializer<'a>;
    type SerializeStructVariant = StructSerializer<'a>;

    fn serialize_bool(self, v: bool) -> Result<()> {
        self.output.push_str(if v { "true" } else { "false" });
        Ok(())
    }

    fn serialize_i8(self, v: i8) -> Result<()> {
        self.serialize_i64(v as i64)
    }
    fn serialize_i16(self, v: i16) -> Result<()> {
        self.serialize_i64(v as i64)
    }
    fn serialize_i32(self, v: i32) -> Result<()> {
        self.serialize_i64(v as i64)
    }
    fn serialize_i64(self, v: i64) -> Result<()> {
        self.output.push_str(&v.to_string());
        Ok(())
    }

    fn serialize_u8(self, v: u8) -> Result<()> {
        self.serialize_u64(v as u64)
    }
    fn serialize_u16(self, v: u16) -> Result<()> {
        self.serialize_u64(v as u64)
    }
    fn serialize_u32(self, v: u32) -> Result<()> {
        self.serialize_u64(v as u64)
    }
    fn serialize_u64(self, v: u64) -> Result<()> {
        self.output.push_str(&v.to_string());
        Ok(())
    }

    fn serialize_f32(self, v: f32) -> Result<()> {
        self.serialize_f64(v as f64)
    }
    fn serialize_f64(self, v: f64) -> Result<()> {
        self.output.push_str(&v.to_string());
        Ok(())
    }

    fn serialize_char(self, v: char) -> Result<()> {
        self.serialize_str(&v.to_string())
    }

    fn serialize_str(self, v: &str) -> Result<()> {
        self.write_string(v);
        Ok(())
    }

    fn serialize_bytes(self, v: &[u8]) -> Result<()> {
        use serde::ser::SerializeSeq;
        let mut seq = self.serialize_seq(Some(v.len()))?;
        for b in v {
            seq.serialize_element(b)?;
        }
        seq.end()
    }

    fn serialize_none(self) -> Result<()> {
        self.serialize_unit()
    }
    fn serialize_some<T: ?Sized + Serialize>(self, v: &T) -> Result<()> {
        v.serialize(self)
    }
    fn serialize_unit(self) -> Result<()> {
        self.output.push_str("null");
        Ok(())
    }
    fn serialize_unit_struct(self, _: &'static str) -> Result<()> {
        self.serialize_unit()
    }

    fn serialize_unit_variant(self, _: &'static str, _: u32, variant: &'static str) -> Result<()> {
        self.serialize_str(variant)
    }

    fn serialize_newtype_struct<T: ?Sized + Serialize>(self, _: &'static str, v: &T) -> Result<()> {
        v.serialize(self)
    }

    fn serialize_newtype_variant<T: ?Sized + Serialize>(
        self,
        _: &'static str,
        _: u32,
        variant: &'static str,
        v: &T,
    ) -> Result<()> {
        self.output.push_str(variant);
        self.output.push(':');
        v.serialize(self)
    }

    fn serialize_seq(self, _len: Option<usize>) -> Result<Self::SerializeSeq> {
        self.depth += 1;
        Ok(SeqSerializer {
            ser: self,
            elements: Vec::new(),
            schema: None,
        })
    }

    fn serialize_tuple(self, len: usize) -> Result<Self::SerializeTuple> {
        self.serialize_seq(Some(len))
    }

    fn serialize_tuple_struct(
        self,
        _: &'static str,
        len: usize,
    ) -> Result<Self::SerializeTupleStruct> {
        self.serialize_seq(Some(len))
    }

    fn serialize_tuple_variant(
        self,
        _: &'static str,
        _: u32,
        variant: &'static str,
        _len: usize,
    ) -> Result<Self::SerializeTupleVariant> {
        self.output.push_str(variant);
        self.output.push(':');
        self.depth += 1;
        Ok(SeqSerializer {
            ser: self,
            elements: Vec::new(),
            schema: None,
        })
    }

    fn serialize_map(self, _len: Option<usize>) -> Result<Self::SerializeMap> {
        Ok(MapSerializer {
            ser: self,
            first: true,
            keys: Vec::new(),
        })
    }

    fn serialize_struct(self, _: &'static str, _len: usize) -> Result<Self::SerializeStruct> {
        Ok(StructSerializer {
            ser: self,
            first: true,
            schema: String::new(),
            data: String::new(),
        })
    }

    fn serialize_struct_variant(
        self,
        _: &'static str,
        _: u32,
        variant: &'static str,
        _len: usize,
    ) -> Result<Self::SerializeStructVariant> {
        self.output.push_str(variant);
        self.output.push(':');
        Ok(StructSerializer {
            ser: self,
            first: true,
            schema: String::new(),
            data: String::new(),
        })
    }
}

/// Sequence serializer.
pub struct SeqSerializer<'a> {
    ser: &'a mut Serializer,
    elements: Vec<String>,
    schema: Option<String>,
}

impl<'a> ser::SerializeSeq for SeqSerializer<'a> {
    type Ok = ();
    type Error = Error;

    fn serialize_element<T: ?Sized + Serialize>(&mut self, v: &T) -> Result<()> {
        // Always serialize without indent to get clean output
        let mut elem_ser = Serializer::new();
        v.serialize(&mut elem_ser)?;

        // Check if element is a struct with schema
        if elem_ser.output.starts_with('{') && elem_ser.output.contains("}:(") {
            let colon_pos = elem_ser.output.find("}:(").unwrap();
            let elem_schema = &elem_ser.output[1..colon_pos];
            let elem_data = &elem_ser.output[colon_pos + 2..]; // includes :(...)

            if self.schema.is_none() {
                self.schema = Some(elem_schema.to_string());
            }
            self.elements.push(elem_data.to_string());
        } else {
            self.elements.push(elem_ser.output);
        }
        Ok(())
    }

    fn end(self) -> Result<()> {
        self.ser.depth -= 1;
        if let Some(schema) = self.schema {
            // Array of structs: {fields}:(v1),(v2),...
            self.ser.output.push('{');
            self.ser.output.push_str(&schema);
            self.ser.output.push_str("}:");

            if let Some(ref indent) = self.ser.indent {
                // Pretty print: each tuple on new line
                self.ser.output.push('\n');
                for (i, elem) in self.elements.iter().enumerate() {
                    if i > 0 {
                        self.ser.output.push_str(",\n");
                    }
                    self.ser.output.push_str(indent);
                    self.ser.output.push_str(elem);
                }
            } else {
                for (i, elem) in self.elements.iter().enumerate() {
                    if i > 0 {
                        self.ser.output.push(',');
                    }
                    self.ser.output.push_str(elem);
                }
            }
        } else {
            // Simple array: [v1,v2,...]
            self.ser.output.push('[');
            for (i, elem) in self.elements.iter().enumerate() {
                if i > 0 {
                    self.ser.output.push(',');
                }
                self.ser.output.push_str(elem);
            }
            self.ser.output.push(']');
        }
        Ok(())
    }
}

impl<'a> ser::SerializeTuple for SeqSerializer<'a> {
    type Ok = ();
    type Error = Error;
    fn serialize_element<T: ?Sized + Serialize>(&mut self, v: &T) -> Result<()> {
        ser::SerializeSeq::serialize_element(self, v)
    }
    fn end(self) -> Result<()> {
        ser::SerializeSeq::end(self)
    }
}

impl<'a> ser::SerializeTupleStruct for SeqSerializer<'a> {
    type Ok = ();
    type Error = Error;
    fn serialize_field<T: ?Sized + Serialize>(&mut self, v: &T) -> Result<()> {
        ser::SerializeSeq::serialize_element(self, v)
    }
    fn end(self) -> Result<()> {
        ser::SerializeSeq::end(self)
    }
}

impl<'a> ser::SerializeTupleVariant for SeqSerializer<'a> {
    type Ok = ();
    type Error = Error;
    fn serialize_field<T: ?Sized + Serialize>(&mut self, v: &T) -> Result<()> {
        ser::SerializeSeq::serialize_element(self, v)
    }
    fn end(self) -> Result<()> {
        ser::SerializeSeq::end(self)
    }
}

/// Map serializer.
pub struct MapSerializer<'a> {
    ser: &'a mut Serializer,
    first: bool,
    keys: Vec<String>,
}

impl<'a> ser::SerializeMap for MapSerializer<'a> {
    type Ok = ();
    type Error = Error;

    fn serialize_key<T: ?Sized + Serialize>(&mut self, key: &T) -> Result<()> {
        let mut key_ser = Serializer::new();
        key.serialize(&mut key_ser)?;
        self.keys.push(key_ser.output);
        Ok(())
    }

    fn serialize_value<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<()> {
        if !self.first {
            self.ser.output.push(',');
        }
        self.first = false;
        value.serialize(&mut *self.ser)
    }

    fn end(self) -> Result<()> {
        let schema = format!("{{{}}}", self.keys.join(","));
        let data = std::mem::take(&mut self.ser.output);
        self.ser.output = format!("{}:({})", schema, data);
        Ok(())
    }
}

/// Struct serializer.
pub struct StructSerializer<'a> {
    ser: &'a mut Serializer,
    first: bool,
    schema: String,
    data: String,
}

impl<'a> ser::SerializeStruct for StructSerializer<'a> {
    type Ok = ();
    type Error = Error;

    fn serialize_field<T: ?Sized + Serialize>(
        &mut self,
        key: &'static str,
        value: &T,
    ) -> Result<()> {
        // Build schema
        if !self.first {
            self.schema.push(',');
            self.data.push(',');
        }
        self.first = false;

        // Always serialize without indent to get clean schema extraction
        let mut value_ser = Serializer::new();
        value.serialize(&mut value_ser)?;

        // Check if the value has its own schema (struct or array of structs)
        if value_ser.output.starts_with('{') && value_ser.output.contains("}:(") {
            // Nested struct: {fields}:(values)
            let colon_pos = value_ser.output.find("}:(").unwrap();
            let nested_schema = &value_ser.output[1..colon_pos];
            let nested_data = &value_ser.output[colon_pos + 3..value_ser.output.len() - 1];
            self.schema
                .push_str(&format!("{}{{{}}}", key, nested_schema));
            self.data.push_str(&format!("({})", nested_data));
        } else if value_ser.output.starts_with("{") && value_ser.output.contains("}:") {
            // Array of structs: {fields}:(v1),(v2)
            let colon_pos = value_ser.output.find("}:").unwrap();
            let nested_schema = &value_ser.output[1..colon_pos];
            let nested_data = &value_ser.output[colon_pos + 2..];
            self.schema
                .push_str(&format!("{}[]{{{}}}", key, nested_schema));
            self.data.push_str(&format!("[{}]", nested_data));
        } else if value_ser.output.starts_with('[') {
            // Simple array
            self.schema.push_str(&format!("{}[]", key));
            self.data.push_str(&value_ser.output);
        } else {
            self.schema.push_str(key);
            self.data.push_str(&value_ser.output);
        }
        Ok(())
    }

    fn end(self) -> Result<()> {
        // Single struct: always compact, pretty print only affects arrays
        self.ser.output = format!("{{{}}}:({})", self.schema, self.data);
        Ok(())
    }
}

impl<'a> ser::SerializeStructVariant for StructSerializer<'a> {
    type Ok = ();
    type Error = Error;

    fn serialize_field<T: ?Sized + Serialize>(
        &mut self,
        key: &'static str,
        value: &T,
    ) -> Result<()> {
        ser::SerializeStruct::serialize_field(self, key, value)
    }

    fn end(self) -> Result<()> {
        ser::SerializeStruct::end(self)
    }
}
