package io.ason.examples;

import io.ason.Ason;
import java.util.*;

public class CrossCompatExample {

    public static class Detail {
        public long ID;
        public String Name;
        public int Age;
        public boolean Gender;

        public Detail() {
        }

        public Detail(long ID, String Name, int Age, boolean Gender) {
            this.ID = ID;
            this.Name = Name;
            this.Age = Age;
            this.Gender = Gender;
        }
    }

    public static class User {
        public List<Detail> details;

        public User() {
            this.details = new ArrayList<>();
        }
    }

    public static class Person {
        public long ID;
        public String Name;
        public int Age;

        public Person() {
        }

        @Override
        public String toString() {
            return "Person{ID=" + ID + ", Name='" + Name + "', Age=" + Age + "}";
        }
    }

    public static class Human {
        public List<Person> details;

        public Human() {
            this.details = new ArrayList<>();
        }

        @Override
        public String toString() {
            return "Human{details=" + details + "}";
        }
    }

    public static void main(String[] args) {
        // Create User data
        User u1 = new User();
        u1.details.add(new Detail(1, "Alice", 30, true));
        u1.details.add(new Detail(2, "Bob", 25, false));
        List<User> users = List.of(u1);

        // Encode
        String asonStr = Ason.encode(new ArrayList<>(users));
        System.out.println("Encoded ASON:");
        System.out.println(asonStr);

        // Decode into Human
        List<Human> decoded = Ason.decodeList(asonStr, Human.class);
        System.out.println("\nDecoded into Human list:");
        for (Human h : decoded) {
            System.out.println(h);
        }
    }
}
