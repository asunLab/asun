# Python 指南

Python ASON 实现是单文件库（`ason.py`），零外部依赖，仅需 Python 标准库。

## 安装

将 `ason.py` 复制到你的项目：

```bash
cp /path/to/ason/python/ason.py ./ason.py
```

或从仓库安装：

```bash
pip install git+https://github.com/your-org/ason.git#subdirectory=python
```

## 编码

```python
from ason import encode, encode_rows
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    active: bool

# 单个结构体
user = User(1, "Alice", True)
text = encode(user)
# → "{id,name,active}:(1,Alice,true)"

# 列表
users = [User(1, "Alice", True), User(2, "Bob", False)]
text = encode_rows(users)
# → "{id,name,active}:\n  (1,Alice,true),\n  (2,Bob,false)"
```

## 解码

```python
from ason import decode, decode_rows

# 单个结构体
user = decode(text, User)

# 列表
users = decode_rows(text, User)
```

## 类型支持

| Python 类型 | ASON 类型 |
|-------------|-----------|
| `int` | `int` |
| `float` | `float` |
| `bool` | `bool` |
| `str` | `str` |
| `None` | 空槽 |
| `list` | `[type]` |
| `dataclass` | 嵌套 schema |
| `Optional[T]` | 可空槽 |

## 使用字典

如果没有 dataclass，使用 `encode_dict_rows`：

```python
from ason import encode_dict_rows, decode_dict_rows

rows = [
    {"id": 1, "name": "Alice", "active": True},
    {"id": 2, "name": "Bob",   "active": False},
]
text = encode_dict_rows(rows)
back = decode_dict_rows(text)
```

## 运行测试

```bash
cd python && python -m pytest test_ason.py -v
```

## 运行示例

```bash
cd python/examples
python basic.py
python bench.py
python complex.py
```
