// Conformance runner for asun-cpp.
// Loads ../../cases.json (untyped decode) and ../../encode-cases.json
// (round-trip), feeding each case through asun::decode_value / encode_value.
//
// JSON parsing is done with a small inline parser — keeping the runner
// self-contained avoids pulling external dependencies and matches the
// approach used by the asun-java runner.

#include "asun.hpp"
#include "asun_value.hpp"

#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <memory>
#include <stdexcept>
#include <variant>
#include <map>

// ============================================================================
// Minimal JSON parser (just enough for cases.json / encode-cases.json)
// ============================================================================
namespace js {

struct JVal;
using JObj = std::map<std::string, JVal>;
using JArr = std::vector<JVal>;

struct JVal {
    enum T { N, B, I, D, S, A, O } t = N;
    bool b = false;
    int64_t i = 0;
    double d = 0.0;
    std::string s;
    std::shared_ptr<JArr> a;
    std::shared_ptr<JObj> o;
};

struct Parser {
    const std::string& src;
    size_t p = 0;
    explicit Parser(const std::string& s) : src(s) {}

    void skip() {
        while (p < src.size()) {
            char c = src[p];
            if (c == ' ' || c == '\t' || c == '\n' || c == '\r') { p++; continue; }
            break;
        }
    }
    JVal parse() { skip(); JVal v = val(); skip(); return v; }
    JVal val() {
        skip();
        if (p >= src.size()) throw std::runtime_error("unexpected EOF");
        char c = src[p];
        if (c == '{') return object();
        if (c == '[') return array();
        if (c == '"') { JVal v; v.t = JVal::S; v.s = str(); return v; }
        if (c == 't') { p += 4; JVal v; v.t = JVal::B; v.b = true; return v; }
        if (c == 'f') { p += 5; JVal v; v.t = JVal::B; v.b = false; return v; }
        if (c == 'n') { p += 4; JVal v; v.t = JVal::N; return v; }
        return num();
    }
    JVal object() {
        JVal v; v.t = JVal::O; v.o = std::make_shared<JObj>();
        p++;
        skip();
        if (p < src.size() && src[p] == '}') { p++; return v; }
        for (;;) {
            skip();
            std::string k = str();
            skip();
            if (p >= src.size() || src[p] != ':') throw std::runtime_error("expected ':'");
            p++;
            (*v.o)[k] = val();
            skip();
            if (p >= src.size()) throw std::runtime_error("unexpected EOF in object");
            char c = src[p++];
            if (c == ',') continue;
            if (c == '}') return v;
            throw std::runtime_error("expected ',' or '}'");
        }
    }
    JVal array() {
        JVal v; v.t = JVal::A; v.a = std::make_shared<JArr>();
        p++;
        skip();
        if (p < src.size() && src[p] == ']') { p++; return v; }
        for (;;) {
            v.a->push_back(val());
            skip();
            if (p >= src.size()) throw std::runtime_error("unexpected EOF in array");
            char c = src[p++];
            if (c == ',') continue;
            if (c == ']') return v;
            throw std::runtime_error("expected ',' or ']'");
        }
    }
    std::string str() {
        if (p >= src.size() || src[p] != '"') throw std::runtime_error("expected '\"'");
        p++;
        std::string out;
        while (p < src.size()) {
            char c = src[p++];
            if (c == '"') return out;
            if (c == '\\') {
                if (p >= src.size()) throw std::runtime_error("bad escape");
                char e = src[p++];
                switch (e) {
                    case '"':  out.push_back('"'); break;
                    case '\\': out.push_back('\\'); break;
                    case '/':  out.push_back('/'); break;
                    case 'n':  out.push_back('\n'); break;
                    case 't':  out.push_back('\t'); break;
                    case 'r':  out.push_back('\r'); break;
                    case 'b':  out.push_back('\b'); break;
                    case 'f':  out.push_back('\f'); break;
                    case 'u': {
                        if (p + 4 > src.size()) throw std::runtime_error("bad \\u");
                        char hex[5] = {src[p], src[p+1], src[p+2], src[p+3], 0};
                        p += 4;
                        unsigned long cp = std::strtoul(hex, nullptr, 16);
                        if (cp < 0x80) out.push_back(static_cast<char>(cp));
                        else if (cp < 0x800) {
                            out.push_back(static_cast<char>(0xC0 | (cp >> 6)));
                            out.push_back(static_cast<char>(0x80 | (cp & 0x3F)));
                        } else {
                            out.push_back(static_cast<char>(0xE0 | (cp >> 12)));
                            out.push_back(static_cast<char>(0x80 | ((cp >> 6) & 0x3F)));
                            out.push_back(static_cast<char>(0x80 | (cp & 0x3F)));
                        }
                        break;
                    }
                    default: throw std::runtime_error("bad escape");
                }
            } else {
                out.push_back(c);
            }
        }
        throw std::runtime_error("unterminated string");
    }
    JVal num() {
        size_t start = p;
        if (p < src.size() && src[p] == '-') p++;
        while (p < src.size() && src[p] >= '0' && src[p] <= '9') p++;
        bool is_float = false;
        if (p < src.size() && src[p] == '.') { is_float = true; p++; while (p < src.size() && src[p] >= '0' && src[p] <= '9') p++; }
        if (p < src.size() && (src[p] == 'e' || src[p] == 'E')) {
            is_float = true; p++;
            if (p < src.size() && (src[p] == '+' || src[p] == '-')) p++;
            while (p < src.size() && src[p] >= '0' && src[p] <= '9') p++;
        }
        std::string tok = src.substr(start, p - start);
        JVal v;
        if (is_float) {
            v.t = JVal::D;
            v.d = std::strtod(tok.c_str(), nullptr);
        } else {
            v.t = JVal::I;
            v.i = std::strtoll(tok.c_str(), nullptr, 10);
        }
        return v;
    }
};

// Convert JSON expected value -> asun::Value for comparison
inline asun::Value jv_to_av(const JVal& j) {
    switch (j.t) {
        case JVal::N: return asun::Value::null_();
        case JVal::B: return asun::Value::boolean(j.b);
        case JVal::I: return asun::Value::integer(j.i);
        case JVal::D: return asun::Value::floating(j.d);
        case JVal::S: return asun::Value::string(j.s);
        case JVal::A: {
            std::vector<asun::Value> items;
            items.reserve(j.a->size());
            for (auto& e : *j.a) items.push_back(jv_to_av(e));
            return asun::Value::array(std::move(items));
        }
        case JVal::O:
            // not expected in conformance data, but be safe
            return asun::Value::null_();
    }
    return asun::Value::null_();
}

inline std::string read_file(const char* path) {
    std::ifstream f(path);
    if (!f) { std::fprintf(stderr, "cannot open %s\n", path); std::exit(2); }
    std::stringstream ss;
    ss << f.rdbuf();
    return ss.str();
}

} // namespace js

// ============================================================================
// Runner
// ============================================================================
struct Failure {
    std::string id;
    std::string msg;
};

int main() {
    using namespace js;

    // ------- decode (cases.json, untyped) --------------------------------
    auto decode_text = read_file("../../cases.json");
    Parser dp(decode_text);
    JVal decode_root = dp.parse();
    auto& dcases = *decode_root.o->at("cases").a;
    std::printf("loaded %zu cases from conformance/cases.json\n", dcases.size());

    size_t d_total = 0, d_ok_pass = 0, d_ok_fail = 0;
    size_t d_err_pass = 0, d_err_fail = 0, d_skipped = 0;
    std::vector<Failure> d_failures;

    for (auto& cv : dcases) {
        auto& c = *cv.o;
        d_total++;
        std::string id = c.at("id").s;
        if (auto it = c.find("schemaDriven"); it != c.end() && it->second.t == JVal::B && it->second.b) {
            d_skipped++; continue;
        }
        std::string input = c.at("input").s;
        std::string kind  = c.at("kind").s;

        asun::Value got;
        bool ok = true;
        std::string err_msg;
        try {
            got = asun::decode_value(input);
        } catch (const std::exception& e) {
            ok = false;
            err_msg = e.what();
        }
        if (!ok) {
            if (kind == "error") d_err_pass++;
            else {
                d_ok_fail++;
                if (d_failures.size() < 25)
                    d_failures.push_back({id, "expected ok, got error: " + err_msg + "\n    input: " + input});
            }
            continue;
        }
        if (kind == "ok") {
            asun::Value expected = jv_to_av(c.at("expected"));
            if (got == expected) d_ok_pass++;
            else {
                d_ok_fail++;
                if (d_failures.size() < 25)
                    d_failures.push_back({id,
                        "value mismatch\n    input:    " + input +
                        "\n    expected: " + expected.to_diagnostic() +
                        "\n    actual:   " + got.to_diagnostic()});
            }
        } else {
            d_err_fail++;
            if (d_failures.size() < 25)
                d_failures.push_back({id, "expected error, got ok: " + got.to_diagnostic() + "\n    input: " + input});
        }
    }

    std::printf("\n================ ASUN-CPP conformance ================\n");
    std::printf("total                : %zu\n", d_total);
    std::printf("untyped ok-cases pass: %zu\n", d_ok_pass);
    std::printf("untyped ok-cases fail: %zu\n", d_ok_fail);
    std::printf("untyped err-cases pass: %zu\n", d_err_pass);
    std::printf("untyped err-cases fail: %zu\n", d_err_fail);
    std::printf("skipped (needs typed): %zu\n", d_skipped);
    size_t executed = d_total - d_skipped;
    double dpct = executed > 0 ? 100.0 * (d_ok_pass + d_err_pass) / executed : 0.0;
    std::printf("untyped pass rate    : %zu/%zu (%.1f%%)\n", d_ok_pass + d_err_pass, executed, dpct);
    std::printf("=======================================================\n");
    for (auto& f : d_failures) std::printf("\n[%s]\n    %s\n", f.id.c_str(), f.msg.c_str());

    // ------- encode (encode-cases.json, round-trip) ----------------------
    auto enc_text = read_file("../../encode-cases.json");
    Parser ep(enc_text);
    JVal enc_root = ep.parse();
    auto& ecases = *enc_root.o->at("cases").a;
    std::printf("loaded %zu encode cases from conformance/encode-cases.json\n", ecases.size());

    size_t e_pass = 0, e_fail = 0;
    std::vector<Failure> e_failures;

    for (auto& cv : ecases) {
        auto& c = *cv.o;
        std::string id = c.at("id").s;
        asun::Value value = jv_to_av(c.at("value"));

        std::string encoded;
        try {
            encoded = asun::encode_value(value);
        } catch (const std::exception& e) {
            e_fail++;
            if (e_failures.size() < 25)
                e_failures.push_back({id, std::string("encode error: ") + e.what()});
            continue;
        }

        asun::Value decoded;
        try {
            decoded = asun::decode_value(encoded);
        } catch (const std::exception& e) {
            e_fail++;
            if (e_failures.size() < 25)
                e_failures.push_back({id, "decode error after encode: " + std::string(e.what()) + "\n    encoded: " + encoded});
            continue;
        }

        if (decoded == value) e_pass++;
        else {
            e_fail++;
            if (e_failures.size() < 25)
                e_failures.push_back({id,
                    "round-trip mismatch\n    encoded:  " + encoded +
                    "\n    expected: " + value.to_diagnostic() +
                    "\n    actual:   " + decoded.to_diagnostic()});
        }
    }

    std::printf("\n================ ASUN-CPP encode conformance ================\n");
    std::printf("total : %zu\n", ecases.size());
    std::printf("pass  : %zu\n", e_pass);
    std::printf("fail  : %zu\n", e_fail);
    double ept = ecases.size() > 0 ? 100.0 * e_pass / ecases.size() : 0.0;
    std::printf("rate  : %zu/%zu (%.1f%%)\n", e_pass, ecases.size(), ept);
    std::printf("==============================================================\n");
    for (auto& f : e_failures) std::printf("\n[%s]\n    %s\n", f.id.c_str(), f.msg.c_str());

    if (d_ok_fail || d_err_fail || e_fail) return 1;
    return 0;
}
