# LLM 集成

ASON 从设计之初就考虑了在 LLM（大语言模型）工作流中的应用。本页解释它为何比 JSON 更适合结构化模型输出。

## JSON + LLM 的痛点

要求 LLM 以 JSON 返回结构化数据时，常见的失败模式有：

1. **尾逗号** —— JSON 不合法，但极易写出
2. **括号不匹配** —— 深层嵌套时尤为常见
3. **忘记引号** —— `"name": Alice` 而非 `"name": "Alice"`
4. **字段名冗余** —— 模型为每行重复输出成千上万个 Key，浪费大量 Token

即使使用 Function Calling / Tool Use，基于 JSON 的结构化输出在复杂 Schema 下依然错误率偏高。

## ASON 的优势

### 1. 更少出错点

ASON 需要更少的特殊字符。大多数字符串值**无需引号**：

```ason
[{name:str, role:str}]:
  (Alice, admin),
  (Bob,   viewer)
```

`Alice`、`admin`、`viewer` 不需要引号。字符串中不含 `,` 就不需要转义。需要违反的语法规则大幅减少。

### 2. Schema 作为提示词 Header

在系统提示或 Header 中包含一次 Schema：

```
请用 ASON 格式返回数据，Schema 如下：
[{id:int, name:str, sentiment:str, score:float}]

示例：
[{id:int, name:str, sentiment:str, score:float}]:
  (1, 产品A, 正面, 0.92),
  (2, 产品B, 中性, 0.51)

请分析以下评论：
...
```

模型只需输出数据行 —— Schema 已确立，字段名幻觉彻底消失。

### 3. Token 预算

ASON 比等价 JSON 少 65% Token，同样的上下文窗口可以多放 3 倍数据行。这对批量分析、少样本示例和 RAG 上下文注入都有极大价值。

### 100 条记录的 Token 粗略对比

| 格式 | 约 Token 数 |
|------|------------|
| JSON | ~3 000 |
| ASON | ~1 050 |

### 4. 流式处理友好

ASON 的行导向格式使每个元组 `(...)` 都是独立的完整单元。客户端可以在流式接收时立即开始解析，无需等待 JSON 数组的最终 `}`。

## 推荐提示词模板

```
## 输出格式
使用 ASON 格式，Schema 如下。只输出 ASON 块，不要任何解释。

Schema: {field1:type, field2:type, ...}

## 数据
...你的输入数据...
```

对于带校验的结构化提取，推荐封装库：
1. 在提示词中发送 Schema
2. 解析 ASON 响应
3. 解析失败时重试一次（ASON 格式错误率极低）

## 示例：情感分析 Pipeline

```python
import openai
from ason import decode_rows
from dataclasses import dataclass

@dataclass
class Review:
    id: int
    sentiment: str
    score: float

SCHEMA = "{id:int, sentiment:str, score:float}"

def analyze(reviews: list[str]) -> list[Review]:
    numbered = "\n".join(f"{i+1}. {r}" for i, r in enumerate(reviews))
    prompt = f"""分析每条评论的情感倾向。
只用 ASON 格式返回，Schema：{SCHEMA}

评论：
{numbered}"""

    resp = openai.chat.completions.create(
        model="gpt-4o",
        messages=[{"role": "user", "content": prompt}]
    )
    return decode_rows(resp.choices[0].message.content, Review)
```
