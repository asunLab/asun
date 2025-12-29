"""Tests for ASON Python library."""

import sys
import os
import unittest

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from parser import parse
from serializer import serialize, serialize_with_schema


class TestParser(unittest.TestCase):
    """Parser tests"""

    def test_parse_simple_object(self):
        result = parse("{name,age}:(Alice,30)")
        self.assertEqual(result["name"], "Alice")
        self.assertEqual(result["age"], 30)

    def test_parse_multiple_objects(self):
        result = parse("{name,age}:(Alice,30),(Bob,25)")
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0]["name"], "Alice")
        self.assertEqual(result[1]["name"], "Bob")

    def test_parse_nested_object(self):
        result = parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
        self.assertEqual(result["name"], "Alice")
        self.assertEqual(result["addr"]["city"], "NYC")
        self.assertEqual(result["addr"]["zip"], 10001)

    def test_parse_array_field(self):
        result = parse("{name,scores[]}:(Alice,[90,85,95])")
        self.assertEqual(result["name"], "Alice")
        self.assertEqual(result["scores"], [90, 85, 95])

    def test_parse_object_array(self):
        result = parse("{users[{id,name}]}:([(1,Alice),(2,Bob)])")
        self.assertEqual(len(result["users"]), 2)
        self.assertEqual(result["users"][0]["id"], 1)
        self.assertEqual(result["users"][1]["name"], "Bob")


class TestSerializer(unittest.TestCase):
    """Serializer tests"""

    def test_serialize(self):
        obj = {"name": "Alice", "age": 30}
        result = serialize(obj)
        self.assertEqual(result, "(Alice,30)")

    def test_serialize_with_schema(self):
        obj = {"name": "Alice", "age": 30}
        result = serialize_with_schema(obj)
        self.assertEqual(result, "{name,age}:(Alice,30)")


class TestUnicode(unittest.TestCase):
    """Unicode tests"""

    def test_unicode(self):
        result = parse("{name,city}:(小明,北京)")
        self.assertEqual(result["name"], "小明")
        self.assertEqual(result["city"], "北京")


class TestTypes(unittest.TestCase):
    """Value type tests"""

    def test_types(self):
        result = parse("{a,b,c,d,e}:(null,true,42,3.14,hello)")
        self.assertIsNone(result["a"])
        self.assertTrue(result["b"])
        self.assertEqual(result["c"], 42)
        self.assertAlmostEqual(result["d"], 3.14, places=2)
        self.assertEqual(result["e"], "hello")


if __name__ == "__main__":
    unittest.main(verbosity=2)

