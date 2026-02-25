# Java 指南

Java ASON 库通过反射和注解处理，提供 Java 开发者熟悉的类 JSON API。

## 安装

### Gradle

```groovy
// build.gradle
dependencies {
    implementation 'io.ason:ason:0.1.0'
}
```

### Maven

```xml
<dependency>
    <groupId>io.ason</groupId>
    <artifactId>ason</artifactId>
    <version>0.1.0</version>
</dependency>
```

## 编码与解码

```java
import io.ason.Ason;
import io.ason.AsonField;
import java.util.List;

public class Main {
    public static class User {
        @AsonField("id")     public long    id;
        @AsonField("name")   public String  name;
        @AsonField("active") public boolean active;

        public User(long id, String name, boolean active) {
            this.id = id; this.name = name; this.active = active;
        }
    }

    public static void main(String[] args) throws Exception {
        List<User> users = List.of(
            new User(1, "Alice", true),
            new User(2, "Bob",   false)
        );

        // 序列化
        String text = Ason.encodeList(users, User.class);
        System.out.println(text);
        // {id:int,name:str,active:bool}:
        //   (1,Alice,true),
        //   (2,Bob,false)

        // 反序列化
        List<User> restored = Ason.decodeList(text, User.class);
    }
}
```

## 二进制格式

```java
import io.ason.AsonBinary;

byte[]     bytes    = AsonBinary.encode(user);
User       restored = AsonBinary.decode(bytes, User.class);

byte[]     binList   = AsonBinary.encodeList(users, User.class);
List<User> restored2 = AsonBinary.decodeList(binList, User.class);
```

## 字段注解

| 注解 | 说明 |
|------|------|
| `@AsonField("name")` | 覆盖 ASON Schema 中的字段名 |
| `@AsonIgnore` | 排除该字段，不参与编解码 |
| `@AsonType("int")` | 覆盖输出中的类型注解 |

## 构建

```bash
cd java
./gradlew build
./gradlew test
./gradlew run
```
