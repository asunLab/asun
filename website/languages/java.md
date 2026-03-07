# Java Guide

The Java ASON library uses reflection and annotation processing to provide a JSON-style API familiar to Java developers.

## Installation

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

## Encoding & Decoding

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

        // Serialize
        String text = Ason.encodeList(users, User.class);
        System.out.println(text);
        // {id:int,name:str,active:bool}:
        //   (1,Alice,true),
        //   (2,Bob,false)

        // Deserialize
        List<User> restored = Ason.decodeList(text, User.class);
    }
}
```

## Binary Format

```java
import io.ason.AsonBinary;

byte[] bytes   = AsonBinary.encode(user);
User   restored = AsonBinary.decode(bytes, User.class);

byte[]      binList  = AsonBinary.encodeList(users, User.class);
List<User>  restored2 = AsonBinary.decodeList(binList, User.class);
```

## Field Annotations

| Annotation | Description |
|-----------|-------------|
| `@AsonField("name")` | Override field name in ASON schema |
| `@AsonIgnore` | Exclude field from encoding/decoding |
| `@AsonType("int")` | Override type annotation in output |

## Building

```bash
cd java
./gradlew build
./gradlew test
./gradlew run
```
