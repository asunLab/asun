//! Serde support for ASON Value type.
//!
//! This module provides Serialize and Deserialize implementations for Value,
//! allowing conversion between Rust types and ASON values.

use crate::ast::Value;
use indexmap::IndexMap;
use serde::de::{self, IntoDeserializer, MapAccess, SeqAccess, Visitor};
use serde::ser::{SerializeMap, SerializeSeq};
use serde::{Deserialize, Deserializer, Serialize, Serializer};
use std::fmt;

// ============================================================================
// Serialize implementation
// ============================================================================

impl Serialize for Value {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        match self {
            Value::Null => serializer.serialize_none(),
            Value::Bool(b) => serializer.serialize_bool(*b),
            Value::Integer(n) => serializer.serialize_i64(*n),
            Value::Float(n) => serializer.serialize_f64(*n),
            Value::String(s) => serializer.serialize_str(s),
            Value::Array(arr) => {
                let mut seq = serializer.serialize_seq(Some(arr.len()))?;
                for item in arr {
                    seq.serialize_element(item)?;
                }
                seq.end()
            }
            Value::Object(obj) => {
                let mut map = serializer.serialize_map(Some(obj.len()))?;
                for (key, value) in obj {
                    map.serialize_entry(key, value)?;
                }
                map.end()
            }
        }
    }
}

// ============================================================================
// Deserialize implementation
// ============================================================================

impl<'de> Deserialize<'de> for Value {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        deserializer.deserialize_any(ValueVisitor)
    }
}

struct ValueVisitor;

impl<'de> Visitor<'de> for ValueVisitor {
    type Value = Value;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("any valid ASON value")
    }

    fn visit_bool<E>(self, v: bool) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::Bool(v))
    }

    fn visit_i64<E>(self, v: i64) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::Integer(v))
    }

    fn visit_u64<E>(self, v: u64) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::Integer(v as i64))
    }

    fn visit_f64<E>(self, v: f64) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::Float(v))
    }

    fn visit_str<E>(self, v: &str) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::String(v.to_string()))
    }

    fn visit_string<E>(self, v: String) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::String(v))
    }

    fn visit_none<E>(self) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::Null)
    }

    fn visit_unit<E>(self) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        Ok(Value::Null)
    }

    fn visit_seq<A>(self, mut seq: A) -> Result<Self::Value, A::Error>
    where
        A: SeqAccess<'de>,
    {
        let mut arr = Vec::new();
        while let Some(elem) = seq.next_element()? {
            arr.push(elem);
        }
        Ok(Value::Array(arr))
    }

    fn visit_map<A>(self, mut map: A) -> Result<Self::Value, A::Error>
    where
        A: MapAccess<'de>,
    {
        let mut obj = IndexMap::new();
        while let Some((key, value)) = map.next_entry()? {
            obj.insert(key, value);
        }
        Ok(Value::Object(obj))
    }
}

// ============================================================================
// Helper functions for converting Rust types to/from Value
// ============================================================================

/// Convert a serializable Rust value to an ASON Value.
///
/// # Example
/// ```
/// use ason::serde_impl::to_value;
/// use serde::Serialize;
///
/// #[derive(Serialize)]
/// struct Person {
///     name: String,
///     age: i64,
/// }
///
/// let person = Person { name: "Alice".to_string(), age: 30 };
/// let value = to_value(&person).unwrap();
/// ```
pub fn to_value<T: Serialize>(value: &T) -> Result<Value, ToValueError> {
    value.serialize(ValueSerializer)
}

/// Convert an ASON Value to a deserializable Rust type.
///
/// # Example
/// ```
/// use ason::serde_impl::from_value;
/// use ason::Value;
/// use serde::Deserialize;
///
/// #[derive(Deserialize)]
/// struct Person {
///     name: String,
///     age: i64,
/// }
///
/// let value = ason::parse("{name,age}:(Alice,30)").unwrap();
/// let person: Person = from_value(&value).unwrap();
/// ```
pub fn from_value<'de, T: Deserialize<'de>>(value: &'de Value) -> Result<T, FromValueError> {
    T::deserialize(value)
}

// ============================================================================
// Error types
// ============================================================================

/// Error type for to_value conversion.
#[derive(Debug, Clone)]
pub struct ToValueError(String);

impl std::fmt::Display for ToValueError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl std::error::Error for ToValueError {}

impl serde::ser::Error for ToValueError {
    fn custom<T: std::fmt::Display>(msg: T) -> Self {
        ToValueError(msg.to_string())
    }
}

/// Error type for from_value conversion.
#[derive(Debug, Clone)]
pub struct FromValueError(String);

impl std::fmt::Display for FromValueError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl std::error::Error for FromValueError {}

impl serde::de::Error for FromValueError {
    fn custom<T: std::fmt::Display>(msg: T) -> Self {
        FromValueError(msg.to_string())
    }
}

// ============================================================================
// ValueSerializer - serialize Rust types to Value
// ============================================================================

struct ValueSerializer;

impl Serializer for ValueSerializer {
    type Ok = Value;
    type Error = ToValueError;
    type SerializeSeq = SerializeValueSeq;
    type SerializeTuple = SerializeValueSeq;
    type SerializeTupleStruct = SerializeValueSeq;
    type SerializeTupleVariant = SerializeValueSeq;
    type SerializeMap = SerializeValueMap;
    type SerializeStruct = SerializeValueMap;
    type SerializeStructVariant = SerializeValueMap;

    fn serialize_bool(self, v: bool) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Bool(v))
    }

    fn serialize_i8(self, v: i8) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v as i64))
    }

    fn serialize_i16(self, v: i16) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v as i64))
    }

    fn serialize_i32(self, v: i32) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v as i64))
    }

    fn serialize_i64(self, v: i64) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v))
    }

    fn serialize_u8(self, v: u8) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v as i64))
    }

    fn serialize_u16(self, v: u16) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v as i64))
    }

    fn serialize_u32(self, v: u32) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v as i64))
    }

    fn serialize_u64(self, v: u64) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Integer(v as i64))
    }

    fn serialize_f32(self, v: f32) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Float(v as f64))
    }

    fn serialize_f64(self, v: f64) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Float(v))
    }

    fn serialize_char(self, v: char) -> Result<Self::Ok, Self::Error> {
        Ok(Value::String(v.to_string()))
    }

    fn serialize_str(self, v: &str) -> Result<Self::Ok, Self::Error> {
        Ok(Value::String(v.to_string()))
    }

    fn serialize_bytes(self, v: &[u8]) -> Result<Self::Ok, Self::Error> {
        let arr: Vec<Value> = v.iter().map(|&b| Value::Integer(b as i64)).collect();
        Ok(Value::Array(arr))
    }

    fn serialize_none(self) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Null)
    }

    fn serialize_some<T: ?Sized + Serialize>(self, value: &T) -> Result<Self::Ok, Self::Error> {
        value.serialize(self)
    }

    fn serialize_unit(self) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Null)
    }

    fn serialize_unit_struct(self, _name: &'static str) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Null)
    }

    fn serialize_unit_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        variant: &'static str,
    ) -> Result<Self::Ok, Self::Error> {
        Ok(Value::String(variant.to_string()))
    }

    fn serialize_newtype_struct<T: ?Sized + Serialize>(
        self,
        _name: &'static str,
        value: &T,
    ) -> Result<Self::Ok, Self::Error> {
        value.serialize(self)
    }

    fn serialize_newtype_variant<T: ?Sized + Serialize>(
        self,
        _name: &'static str,
        _variant_index: u32,
        variant: &'static str,
        value: &T,
    ) -> Result<Self::Ok, Self::Error> {
        let mut obj = IndexMap::new();
        obj.insert(variant.to_string(), value.serialize(ValueSerializer)?);
        Ok(Value::Object(obj))
    }

    fn serialize_seq(self, len: Option<usize>) -> Result<Self::SerializeSeq, Self::Error> {
        Ok(SerializeValueSeq {
            items: Vec::with_capacity(len.unwrap_or(0)),
        })
    }

    fn serialize_tuple(self, len: usize) -> Result<Self::SerializeTuple, Self::Error> {
        self.serialize_seq(Some(len))
    }

    fn serialize_tuple_struct(
        self,
        _name: &'static str,
        len: usize,
    ) -> Result<Self::SerializeTupleStruct, Self::Error> {
        self.serialize_seq(Some(len))
    }

    fn serialize_tuple_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        _variant: &'static str,
        len: usize,
    ) -> Result<Self::SerializeTupleVariant, Self::Error> {
        self.serialize_seq(Some(len))
    }

    fn serialize_map(self, len: Option<usize>) -> Result<Self::SerializeMap, Self::Error> {
        Ok(SerializeValueMap {
            map: IndexMap::with_capacity(len.unwrap_or(0)),
            next_key: None,
        })
    }

    fn serialize_struct(
        self,
        _name: &'static str,
        len: usize,
    ) -> Result<Self::SerializeStruct, Self::Error> {
        self.serialize_map(Some(len))
    }

    fn serialize_struct_variant(
        self,
        _name: &'static str,
        _variant_index: u32,
        _variant: &'static str,
        len: usize,
    ) -> Result<Self::SerializeStructVariant, Self::Error> {
        self.serialize_map(Some(len))
    }
}

struct SerializeValueSeq {
    items: Vec<Value>,
}

impl serde::ser::SerializeSeq for SerializeValueSeq {
    type Ok = Value;
    type Error = ToValueError;

    fn serialize_element<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<(), Self::Error> {
        self.items.push(value.serialize(ValueSerializer)?);
        Ok(())
    }

    fn end(self) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Array(self.items))
    }
}

impl serde::ser::SerializeTuple for SerializeValueSeq {
    type Ok = Value;
    type Error = ToValueError;

    fn serialize_element<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<(), Self::Error> {
        serde::ser::SerializeSeq::serialize_element(self, value)
    }

    fn end(self) -> Result<Self::Ok, Self::Error> {
        serde::ser::SerializeSeq::end(self)
    }
}

impl serde::ser::SerializeTupleStruct for SerializeValueSeq {
    type Ok = Value;
    type Error = ToValueError;

    fn serialize_field<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<(), Self::Error> {
        serde::ser::SerializeSeq::serialize_element(self, value)
    }

    fn end(self) -> Result<Self::Ok, Self::Error> {
        serde::ser::SerializeSeq::end(self)
    }
}

impl serde::ser::SerializeTupleVariant for SerializeValueSeq {
    type Ok = Value;
    type Error = ToValueError;

    fn serialize_field<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<(), Self::Error> {
        serde::ser::SerializeSeq::serialize_element(self, value)
    }

    fn end(self) -> Result<Self::Ok, Self::Error> {
        serde::ser::SerializeSeq::end(self)
    }
}

struct SerializeValueMap {
    map: IndexMap<String, Value>,
    next_key: Option<String>,
}

impl serde::ser::SerializeMap for SerializeValueMap {
    type Ok = Value;
    type Error = ToValueError;

    fn serialize_key<T: ?Sized + Serialize>(&mut self, key: &T) -> Result<(), Self::Error> {
        let key_value = key.serialize(ValueSerializer)?;
        self.next_key = Some(match key_value {
            Value::String(s) => s,
            _ => return Err(ToValueError("map key must be a string".to_string())),
        });
        Ok(())
    }

    fn serialize_value<T: ?Sized + Serialize>(&mut self, value: &T) -> Result<(), Self::Error> {
        let key = self.next_key.take().ok_or_else(|| {
            ToValueError("serialize_value called before serialize_key".to_string())
        })?;
        self.map.insert(key, value.serialize(ValueSerializer)?);
        Ok(())
    }

    fn end(self) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Object(self.map))
    }
}

impl serde::ser::SerializeStruct for SerializeValueMap {
    type Ok = Value;
    type Error = ToValueError;

    fn serialize_field<T: ?Sized + Serialize>(
        &mut self,
        key: &'static str,
        value: &T,
    ) -> Result<(), Self::Error> {
        self.map
            .insert(key.to_string(), value.serialize(ValueSerializer)?);
        Ok(())
    }

    fn end(self) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Object(self.map))
    }
}

impl serde::ser::SerializeStructVariant for SerializeValueMap {
    type Ok = Value;
    type Error = ToValueError;

    fn serialize_field<T: ?Sized + Serialize>(
        &mut self,
        key: &'static str,
        value: &T,
    ) -> Result<(), Self::Error> {
        self.map
            .insert(key.to_string(), value.serialize(ValueSerializer)?);
        Ok(())
    }

    fn end(self) -> Result<Self::Ok, Self::Error> {
        Ok(Value::Object(self.map))
    }
}

// ============================================================================
// ValueDeserializer - deserialize Value to Rust types
// ============================================================================

impl<'de> Deserializer<'de> for &'de Value {
    type Error = FromValueError;

    fn deserialize_any<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Null => visitor.visit_none(),
            Value::Bool(b) => visitor.visit_bool(*b),
            Value::Integer(n) => visitor.visit_i64(*n),
            Value::Float(n) => visitor.visit_f64(*n),
            Value::String(s) => visitor.visit_str(s),
            Value::Array(arr) => visitor.visit_seq(SeqDeserializer::new(arr)),
            Value::Object(obj) => visitor.visit_map(MapDeserializer::new(obj)),
        }
    }

    fn deserialize_bool<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Bool(b) => visitor.visit_bool(*b),
            _ => Err(FromValueError("expected bool".to_string())),
        }
    }

    fn deserialize_i8<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_i64(visitor)
    }

    fn deserialize_i16<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_i64(visitor)
    }

    fn deserialize_i32<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_i64(visitor)
    }

    fn deserialize_i64<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Integer(n) => visitor.visit_i64(*n),
            _ => Err(FromValueError("expected integer".to_string())),
        }
    }

    fn deserialize_u8<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_u64(visitor)
    }

    fn deserialize_u16<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_u64(visitor)
    }

    fn deserialize_u32<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_u64(visitor)
    }

    fn deserialize_u64<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Integer(n) => visitor.visit_u64(*n as u64),
            _ => Err(FromValueError("expected integer".to_string())),
        }
    }

    fn deserialize_f32<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_f64(visitor)
    }

    fn deserialize_f64<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Float(n) => visitor.visit_f64(*n),
            Value::Integer(n) => visitor.visit_f64(*n as f64),
            _ => Err(FromValueError("expected float".to_string())),
        }
    }

    fn deserialize_char<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_str(visitor)
    }

    fn deserialize_str<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::String(s) => visitor.visit_str(s),
            _ => Err(FromValueError("expected string".to_string())),
        }
    }

    fn deserialize_string<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_str(visitor)
    }

    fn deserialize_bytes<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_seq(visitor)
    }

    fn deserialize_byte_buf<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_seq(visitor)
    }

    fn deserialize_option<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Null => visitor.visit_none(),
            _ => visitor.visit_some(self),
        }
    }

    fn deserialize_unit<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Null => visitor.visit_unit(),
            _ => Err(FromValueError("expected null".to_string())),
        }
    }

    fn deserialize_unit_struct<V>(
        self,
        _name: &'static str,
        visitor: V,
    ) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_unit(visitor)
    }

    fn deserialize_newtype_struct<V>(
        self,
        _name: &'static str,
        visitor: V,
    ) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        visitor.visit_newtype_struct(self)
    }

    fn deserialize_seq<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Array(arr) => visitor.visit_seq(SeqDeserializer::new(arr)),
            _ => Err(FromValueError("expected array".to_string())),
        }
    }

    fn deserialize_tuple<V>(self, _len: usize, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_seq(visitor)
    }

    fn deserialize_tuple_struct<V>(
        self,
        _name: &'static str,
        _len: usize,
        visitor: V,
    ) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_seq(visitor)
    }

    fn deserialize_map<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::Object(obj) => visitor.visit_map(MapDeserializer::new(obj)),
            _ => Err(FromValueError("expected object".to_string())),
        }
    }

    fn deserialize_struct<V>(
        self,
        _name: &'static str,
        _fields: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_map(visitor)
    }

    fn deserialize_enum<V>(
        self,
        _name: &'static str,
        _variants: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self {
            Value::String(s) => visitor.visit_enum(s.as_str().into_deserializer()),
            Value::Object(obj) => {
                if obj.len() != 1 {
                    return Err(FromValueError(
                        "enum must have exactly one field".to_string(),
                    ));
                }
                let (variant, value) = obj.iter().next().unwrap();
                visitor.visit_enum(EnumDeserializer { variant, value })
            }
            _ => Err(FromValueError(
                "expected string or object for enum".to_string(),
            )),
        }
    }

    fn deserialize_identifier<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        self.deserialize_str(visitor)
    }

    fn deserialize_ignored_any<V>(self, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        visitor.visit_unit()
    }
}

struct SeqDeserializer<'de> {
    iter: std::slice::Iter<'de, Value>,
}

impl<'de> SeqDeserializer<'de> {
    fn new(arr: &'de [Value]) -> Self {
        SeqDeserializer { iter: arr.iter() }
    }
}

impl<'de> SeqAccess<'de> for SeqDeserializer<'de> {
    type Error = FromValueError;

    fn next_element_seed<T>(&mut self, seed: T) -> Result<Option<T::Value>, Self::Error>
    where
        T: de::DeserializeSeed<'de>,
    {
        match self.iter.next() {
            Some(value) => seed.deserialize(value).map(Some),
            None => Ok(None),
        }
    }
}

struct MapDeserializer<'de> {
    iter: indexmap::map::Iter<'de, String, Value>,
    value: Option<&'de Value>,
}

impl<'de> MapDeserializer<'de> {
    fn new(obj: &'de IndexMap<String, Value>) -> Self {
        MapDeserializer {
            iter: obj.iter(),
            value: None,
        }
    }
}

impl<'de> MapAccess<'de> for MapDeserializer<'de> {
    type Error = FromValueError;

    fn next_key_seed<K>(&mut self, seed: K) -> Result<Option<K::Value>, Self::Error>
    where
        K: de::DeserializeSeed<'de>,
    {
        match self.iter.next() {
            Some((key, value)) => {
                self.value = Some(value);
                seed.deserialize(de::value::BorrowedStrDeserializer::new(key))
                    .map(Some)
            }
            None => Ok(None),
        }
    }

    fn next_value_seed<V>(&mut self, seed: V) -> Result<V::Value, Self::Error>
    where
        V: de::DeserializeSeed<'de>,
    {
        match self.value.take() {
            Some(value) => seed.deserialize(value),
            None => Err(FromValueError(
                "next_value_seed called before next_key_seed".to_string(),
            )),
        }
    }
}

struct EnumDeserializer<'de> {
    variant: &'de str,
    value: &'de Value,
}

impl<'de> de::EnumAccess<'de> for EnumDeserializer<'de> {
    type Error = FromValueError;
    type Variant = VariantDeserializer<'de>;

    fn variant_seed<V>(self, seed: V) -> Result<(V::Value, Self::Variant), Self::Error>
    where
        V: de::DeserializeSeed<'de>,
    {
        let variant = seed.deserialize(de::value::BorrowedStrDeserializer::new(self.variant))?;
        Ok((variant, VariantDeserializer { value: self.value }))
    }
}

struct VariantDeserializer<'de> {
    value: &'de Value,
}

impl<'de> de::VariantAccess<'de> for VariantDeserializer<'de> {
    type Error = FromValueError;

    fn unit_variant(self) -> Result<(), Self::Error> {
        Ok(())
    }

    fn newtype_variant_seed<T>(self, seed: T) -> Result<T::Value, Self::Error>
    where
        T: de::DeserializeSeed<'de>,
    {
        seed.deserialize(self.value)
    }

    fn tuple_variant<V>(self, _len: usize, visitor: V) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self.value {
            Value::Array(arr) => visitor.visit_seq(SeqDeserializer::new(arr)),
            _ => Err(FromValueError(
                "expected array for tuple variant".to_string(),
            )),
        }
    }

    fn struct_variant<V>(
        self,
        _fields: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value, Self::Error>
    where
        V: Visitor<'de>,
    {
        match self.value {
            Value::Object(obj) => visitor.visit_map(MapDeserializer::new(obj)),
            _ => Err(FromValueError(
                "expected object for struct variant".to_string(),
            )),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde::{Deserialize, Serialize};

    #[derive(Debug, Serialize, Deserialize, PartialEq)]
    struct Person {
        name: String,
        age: i64,
    }

    #[derive(Debug, Serialize, Deserialize, PartialEq)]
    struct Company {
        name: String,
        employees: Vec<Person>,
        active: bool,
    }

    #[test]
    fn test_to_value_struct() {
        let person = Person {
            name: "Alice".to_string(),
            age: 30,
        };
        let value = to_value(&person).unwrap();

        assert!(value.is_object());
        assert_eq!(value.get("name"), Some(&Value::String("Alice".to_string())));
        assert_eq!(value.get("age"), Some(&Value::Integer(30)));
    }

    #[test]
    fn test_from_value_struct() {
        let mut obj = IndexMap::new();
        obj.insert("name".to_string(), Value::String("Bob".to_string()));
        obj.insert("age".to_string(), Value::Integer(25));
        let value = Value::Object(obj);

        let person: Person = from_value(&value).unwrap();
        assert_eq!(person.name, "Bob");
        assert_eq!(person.age, 25);
    }

    #[test]
    fn test_roundtrip_struct() {
        let original = Person {
            name: "Charlie".to_string(),
            age: 35,
        };
        let value = to_value(&original).unwrap();
        let restored: Person = from_value(&value).unwrap();
        assert_eq!(original, restored);
    }

    #[test]
    fn test_nested_struct() {
        let company = Company {
            name: "ACME".to_string(),
            employees: vec![
                Person {
                    name: "Alice".to_string(),
                    age: 30,
                },
                Person {
                    name: "Bob".to_string(),
                    age: 25,
                },
            ],
            active: true,
        };

        let value = to_value(&company).unwrap();
        let restored: Company = from_value(&value).unwrap();
        assert_eq!(company, restored);
    }

    #[test]
    fn test_parse_then_deserialize() {
        let ason_str = "{name,age}:(Alice,30)";
        let value = crate::parse(ason_str).unwrap();

        let person: Person = from_value(&value).unwrap();
        assert_eq!(person.name, "Alice");
        assert_eq!(person.age, 30);
    }

    #[test]
    fn test_serialize_then_to_string() {
        let person = Person {
            name: "Alice".to_string(),
            age: 30,
        };
        let value = to_value(&person).unwrap();
        let ason_str = crate::to_string(&value);

        // Object is serialized as data only (no schema)
        assert_eq!(ason_str, "(Alice,30)");
    }

    #[test]
    fn test_option_some() {
        #[derive(Serialize, Deserialize, PartialEq, Debug)]
        struct OptPerson {
            name: String,
            age: Option<i64>,
        }

        let p = OptPerson {
            name: "Alice".to_string(),
            age: Some(30),
        };
        let value = to_value(&p).unwrap();
        let restored: OptPerson = from_value(&value).unwrap();
        assert_eq!(p, restored);
    }

    #[test]
    fn test_option_none() {
        #[derive(Serialize, Deserialize, PartialEq, Debug)]
        struct OptPerson {
            name: String,
            age: Option<i64>,
        }

        let p = OptPerson {
            name: "Bob".to_string(),
            age: None,
        };
        let value = to_value(&p).unwrap();
        let restored: OptPerson = from_value(&value).unwrap();
        assert_eq!(p, restored);
    }

    #[test]
    fn test_vec() {
        let nums = vec![1i64, 2, 3, 4, 5];
        let value = to_value(&nums).unwrap();
        let restored: Vec<i64> = from_value(&value).unwrap();
        assert_eq!(nums, restored);
    }
}
