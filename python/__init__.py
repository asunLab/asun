"""
ASON (Array-Schema Object Notation) Python Library

A pure Python implementation of ASON format.
"""

from parser import parse
from serializer import serialize, serialize_with_schema
from value import Value

__version__ = "0.1.0"
__all__ = ["parse", "serialize", "serialize_with_schema", "Value"]

