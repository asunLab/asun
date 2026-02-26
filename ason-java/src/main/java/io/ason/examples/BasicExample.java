package io.ason.examples;

import io.ason.Ason;
import java.util.*;

public class BasicExample {

    public static class User {
        public long id;
        public String name;
        public boolean active;
        public User() {}
        public User(long id, String name, boolean active) {
            this.id = id; this.name = name; this.active = active;
        }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof User u)) return false;
            return id == u.id && active == u.active && Objects.equals(name, u.name);
        }
        @Override public String toString() { return "User{id=" + id + ", name=" + name + ", active=" + active + "}"; }
    }

    public static class Item {
        public long id;
        public Optional<String> label;
        public Item() { label = Optional.empty(); }
    }

    public static class Tagged {
        public String name;
        public List<String> tags;
        public Tagged() { tags = new ArrayList<>(); }
        @Override public String toString() { return "Tagged{name=" + name + ", tags=" + tags + "}"; }
    }

    public static void main(String[] args) {
        System.out.println("=== ASON Basic Examples (Java) ===\n");

        // 1. Serialize a single struct
        User user = new User(1, "Alice", true);
        String asonStr = Ason.encode(user);
        System.out.println("1. Serialize single struct:");
        System.out.println("  " + asonStr + "\n");

        // 2. Serialize with type annotations (encodeTyped)
        String typedStr = Ason.encodeTyped(user);
        System.out.println("2. Serialize with type annotations:");
        System.out.println("  " + typedStr + "\n");
        assert typedStr.startsWith("{id:int,name:str,active:bool}:");

        // 3. Deserialize from ASON (accepts both annotated and unannotated)
        String input = "{id:int,name:str,active:bool}:(1,Alice,true)";
        User decoded = Ason.decode(input, User.class);
        System.out.println("3. Deserialize single struct:");
        System.out.println("  " + decoded + "\n");

        // 4. Serialize a list of structs (schema-driven)
        List<User> users = List.of(
            new User(1, "Alice", true),
            new User(2, "Bob", false),
            new User(3, "Carol Smith", true)
        );
        String asonVec = Ason.encode(new ArrayList<>(users));
        System.out.println("4. Serialize list (schema-driven):");
        System.out.println("  " + asonVec + "\n");

        // 5. Serialize list with type annotations (encodeTyped)
        String typedVec = Ason.encodeTyped(new ArrayList<>(users));
        System.out.println("5. Serialize list with type annotations:");
        System.out.println("  " + typedVec + "\n");
        assert typedVec.startsWith("[{id:int,name:str,active:bool}]:");

        // 6. Deserialize list
        String listInput = "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false),(3,\"Carol Smith\",true)";
        List<User> decodedUsers = Ason.decodeList(listInput, User.class);
        System.out.println("6. Deserialize list:");
        for (User u : decodedUsers) System.out.println("  " + u);

        // 7. Multiline format
        System.out.println("\n7. Multiline format:");
        String multiline = """
            [{id:int, name:str, active:bool}]:
              (1, Alice, true),
              (2, Bob, false),
              (3, "Carol Smith", true)""";
        List<User> multiUsers = Ason.decodeList(multiline, User.class);
        for (User u : multiUsers) System.out.println("  " + u);

        // 8. Roundtrip (ASON-text + ASON-bin + JSON)
        System.out.println("\n8. Roundtrip (ASON-text vs ASON-bin):");
        User original = new User(42, "Test User", true);
        String asonText = Ason.encode(original);
        User fromAson = Ason.decode(asonText, User.class);
        assert original.equals(fromAson);
        byte[] asonBin = Ason.encodeBinary(original);
        User fromBin = Ason.decodeBinary(asonBin, User.class);
        assert original.equals(fromBin);
        System.out.println("  original:     " + original);
        System.out.println("  ASON text:    " + asonText + " (" + asonText.length() + " B)");
        System.out.println("  ASON binary:  " + asonBin.length + " B");
        System.out.println("  ✓ all formats roundtrip OK");

        // 9. List roundtrip (ASON-text + ASON-bin)
        System.out.println("\n9. List roundtrip (ASON-text vs ASON-bin):");
        String vecAson = Ason.encode(new ArrayList<>(users));
        byte[] vecBin = Ason.encodeBinary(new ArrayList<>(users));
        List<User> v1 = Ason.decodeList(vecAson, User.class);
        List<User> v2 = Ason.decodeBinaryList(vecBin, User.class);
        assert v1.size() == users.size();
        assert v2.size() == users.size();
        System.out.println("  ASON text:   " + vecAson.length() + " B");
        System.out.println("  ASON binary: " + vecBin.length + " B");
        System.out.println("  ✓ list roundtrip OK (all formats)");

        // 10. Optional fields
        System.out.println("\n10. Optional fields:");
        Item item1 = Ason.decode("{id,label}:(1,hello)", Item.class);
        System.out.println("  with value: Item{id=" + item1.id + ", label=" + item1.label + "}");
        Item item2 = Ason.decode("{id,label}:(2,)", Item.class);
        System.out.println("  with null:  Item{id=" + item2.id + ", label=" + item2.label + "}");

        // 11. Array fields
        System.out.println("\n11. Array fields:");
        Tagged tagged = Ason.decode("{name,tags}:(Alice,[rust,go,python])", Tagged.class);
        System.out.println("  " + tagged);

        // 12. Comments
        System.out.println("\n12. With comments:");
        User commented = Ason.decode("/* user list */ {id,name,active}:(1,Alice,true)", User.class);
        System.out.println("  " + commented);

        System.out.println("\n=== All examples passed! ===");
    }
}
