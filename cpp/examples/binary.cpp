/**
 * ASON Binary Format Example (C++)
 *
 * Demonstrates dump_bin<T> / load_bin<T> for fast binary serialization.
 * Wire format: little-endian fixed-width scalars, 4-byte length-prefixed
 * strings and arrays — no field names, purely positional.
 *
 *   std::string  ason::dump_bin(const T&)
 *   T            ason::load_bin<T>(std::string_view)
 */

#include <cassert>
#include <chrono>
#include <iostream>
#include <optional>
#include <string>
#include <string_view>
#include <unordered_map>
#include <vector>
#include "ason.hpp"

/* ------------------------------------------------------------------ */
/* Example structs (ASON_FIELDS must be at file scope)                  */
/* ------------------------------------------------------------------ */

struct User {
    std::string name;
    int64_t     age    = 0;
    double      score  = 0.0;
    bool        active = false;
};
ASON_FIELDS(User,
    (name,   "name",   "str"),
    (age,    "age",    "int"),
    (score,  "score",  "float"),
    (active, "active", "bool"))

struct Movie {
    std::string              title;
    int64_t                  year   = 0;
    double                   rating = 0.0;
    std::vector<std::string> tags;
};
ASON_FIELDS(Movie,
    (title,  "title",  "str"),
    (year,   "year",   "int"),
    (rating, "rating", "float"),
    (tags,   "tags",   "[str]"))

struct Config {
    std::string                              service;
    std::optional<int64_t>                   timeout_ms;
    std::unordered_map<std::string, int64_t> limits;
    std::vector<int64_t>                     ports;
};
ASON_FIELDS(Config,
    (service,    "service",    "str"),
    (timeout_ms, "timeout_ms", "int?"),
    (limits,     "limits",     "{int}"),
    (ports,      "ports",      "[int]"))

/* Zero-copy structs */
struct PacketOwned {
    std::string id;
    int64_t     seq = 0;
    std::string payload;
};
ASON_FIELDS(PacketOwned,
    (id,      "id",      "str"),
    (seq,     "seq",     "int"),
    (payload, "payload", "str"))

/* ------------------------------------------------------------------ */

static double now_ms() {
    using clock = std::chrono::high_resolution_clock;
    using ns    = std::chrono::nanoseconds;
    return std::chrono::duration_cast<ns>(
               clock::now().time_since_epoch()).count() * 1e-6;
}

int main() {
    std::cout << "=== ASON Binary Format Example (C++) ===\n\n";

    /* 1. Single struct roundtrip */
    std::cout << "1. Single struct roundtrip\n";
    {
        User u{"Alice", 30, 95.5, true};
        std::string bin = ason::dump_bin(u);
        std::cout << "   Serialized " << bin.size() << " bytes\n";

        User u2 = ason::load_bin<User>(bin);
        std::cout << "   User { name=\"" << u2.name
                  << "\", age=" << u2.age
                  << ", score=" << u2.score
                  << ", active=" << (u2.active ? "true" : "false") << " }\n";

        assert(u2.name == u.name && u2.age == u.age);
        assert(u2.score == u.score && u2.active == u.active);
    }
    std::cout << "   OK\n\n";

    /* 2. Struct with optional + map + vec */
    std::cout << "2. Config with optional/map/vector fields\n";
    {
        Config cfg{"gateway", 5000,
                   {{"rps", 1000}, {"burst", 200}},
                   {8080, 8443, 9090}};

        std::string bin = ason::dump_bin(cfg);
        std::cout << "   Serialized " << bin.size() << " bytes\n";

        Config cfg2 = ason::load_bin<Config>(bin);
        std::cout << "   Config { service=\"" << cfg2.service
                  << "\", timeout_ms="
                  << (cfg2.timeout_ms ? std::to_string(*cfg2.timeout_ms) : "null")
                  << ", ports=[";
        for (size_t i = 0; i < cfg2.ports.size(); i++) {
            if (i) std::cout << ',';
            std::cout << cfg2.ports[i];
        }
        std::cout << "], limits.size=" << cfg2.limits.size() << " }\n";

        assert(cfg2.service == cfg.service);
        assert(cfg2.timeout_ms == cfg.timeout_ms);
        assert(cfg2.limits == cfg.limits);
        assert(cfg2.ports == cfg.ports);
    }
    std::cout << "   OK\n\n";

    /* 3. Vector of structs */
    std::cout << "3. Vector of structs (batch roundtrip)\n";
    {
        std::vector<Movie> movies = {
            {"The Dark Knight", 2008, 9.0, {"action", "crime"}},
            {"Interstellar",    2014, 8.6, {"sci-fi", "drama"}},
            {"Parasite",        2019, 8.5, {"thriller", "drama"}},
        };

        std::string bin = ason::dump_bin(movies);
        std::cout << "   Serialized " << movies.size() << " movies → "
                  << bin.size() << " bytes\n";

        auto movies2 = ason::load_bin<std::vector<Movie>>(bin);
        assert(movies2.size() == movies.size());
        for (const auto& m : movies2)
            std::cout << "   Movie { title=\"" << m.title
                      << "\", year=" << m.year
                      << ", rating=" << m.rating << " }\n";
        assert(movies2[0].title == movies[0].title);
        assert(movies2[1].tags  == movies[1].tags);
    }
    std::cout << "   OK\n\n";

    /* 4. Zero-copy load via string_view */
    std::cout << "4. Zero-copy deserialization (string_view fields)\n";
    {
        // Binary buffer stays alive; string_view fields point into it (no copy)
        PacketOwned src{"pkt-001", 42, "Hello, zero-copy world!"};
        std::string bin = ason::dump_bin(src);

        // Load back as owned struct (each load_bin<T> with string_view fields
        // returns views into the input buffer — zero allocation for strings)
        const char* pos = bin.data();
        const char* end = bin.data() + bin.size();
        std::string_view id_view, payload_view;
        int64_t seq = 0;
        ason::load_bin_value(pos, end, id_view);
        ason::load_bin_value(pos, end, seq);
        ason::load_bin_value(pos, end, payload_view);

        std::cout << "   Packet { id=\"" << id_view
                  << "\", seq=" << seq
                  << ", payload=\"" << payload_view << "\" }\n";
        const bool zero_copy = (id_view.data() >= bin.data() &&
                                id_view.data() < bin.data() + (ptrdiff_t)bin.size());
        std::cout << "   id points into bin buffer: "
                  << (zero_copy ? "yes (zero-copy)" : "no") << "\n";

        assert(id_view == src.id);
        assert(seq == src.seq);
        assert(payload_view == src.payload);
    }
    std::cout << "   OK\n\n";

    /* 5. Size + speed comparison */
    std::cout << "5. Size + speed comparison (1000 users, 500 iterations)\n";
    {
        const int N = 1000, ITER = 500;
        std::vector<User> users(N);
        for (int i = 0; i < N; i++)
            users[i] = {"User" + std::to_string(i), 20 + i % 50,
                        80.0 + (i % 20) * 0.5, (i % 2 == 0)};

        std::string text_out;
        double t0 = now_ms();
        for (int i = 0; i < ITER; i++) text_out = ason::dump_vec(users);
        double text_ser = now_ms() - t0;

        std::string bin_out;
        t0 = now_ms();
        for (int i = 0; i < ITER; i++) bin_out = ason::dump_bin(users);
        double bin_ser = now_ms() - t0;

        t0 = now_ms();
        for (int i = 0; i < ITER; i++) ason::load_vec<User>(text_out);
        double text_de = now_ms() - t0;

        t0 = now_ms();
        for (int i = 0; i < ITER; i++) ason::load_bin<std::vector<User>>(bin_out);
        double bin_de = now_ms() - t0;

        double saving = (1.0 - (double)bin_out.size() / text_out.size()) * 100.0;
        std::cout << "   Serialize:   ASON-text " << text_ser << " ms | BIN "
                  << bin_ser << " ms (" << (text_ser / bin_ser) << "x faster)\n";
        std::cout << "   Deserialize: ASON-text " << text_de  << " ms | BIN "
                  << bin_de  << " ms (" << (text_de  / bin_de ) << "x faster)\n";
        std::cout << "   Size:        ASON-text " << text_out.size()
                  << " B | BIN " << bin_out.size()
                  << " B (" << saving << "% smaller)\n";
    }
    std::cout << "   OK\n\n";

    std::cout << "All examples passed.\n";
    return 0;
}

