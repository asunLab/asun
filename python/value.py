"""ASON Value types."""

from typing import Any, Dict, List, Optional, Union

AsonValue = Union[None, bool, int, float, str, List["AsonValue"], Dict[str, "AsonValue"]]


class Value:
    """Represents an ASON value."""
    
    def __init__(self, data: AsonValue = None):
        self._data = data
    
    @classmethod
    def null(cls) -> "Value":
        return cls(None)
    
    @classmethod
    def boolean(cls, v: bool) -> "Value":
        return cls(v)
    
    @classmethod
    def integer(cls, v: int) -> "Value":
        return cls(v)
    
    @classmethod
    def floating(cls, v: float) -> "Value":
        return cls(v)
    
    @classmethod
    def string(cls, v: str) -> "Value":
        return cls(v)
    
    @classmethod
    def array(cls, items: Optional[List["Value"]] = None) -> "Value":
        return cls([v._data if isinstance(v, Value) else v for v in (items or [])])
    
    @classmethod
    def object(cls, fields: Optional[Dict[str, "Value"]] = None) -> "Value":
        return cls({k: v._data if isinstance(v, Value) else v for k, v in (fields or {}).items()})
    
    @property
    def data(self) -> AsonValue:
        return self._data
    
    def is_null(self) -> bool:
        return self._data is None
    
    def is_bool(self) -> bool:
        return isinstance(self._data, bool)
    
    def is_int(self) -> bool:
        return isinstance(self._data, int) and not isinstance(self._data, bool)
    
    def is_float(self) -> bool:
        return isinstance(self._data, float)
    
    def is_number(self) -> bool:
        return self.is_int() or self.is_float()
    
    def is_string(self) -> bool:
        return isinstance(self._data, str)
    
    def is_array(self) -> bool:
        return isinstance(self._data, list)
    
    def is_object(self) -> bool:
        return isinstance(self._data, dict)
    
    def __getitem__(self, key):
        return self._data[key]
    
    def __len__(self):
        return len(self._data) if self._data else 0
    
    def __repr__(self):
        return f"Value({self._data!r})"

