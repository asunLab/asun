// Conformance runner for asun-go.
//
// Loads ../../cases.json and decodes each non-schema-driven case into
// a generic interface{}. Reports pass/fail counts.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	asun "github.com/asunLab/asun-go"
)

type Case struct {
	ID            string      `json:"id"`
	Category      string      `json:"category"`
	Desc          string      `json:"desc"`
	Input         string      `json:"input"`
	Kind          string      `json:"kind"`
	SchemaDriven  bool        `json:"schemaDriven"`
	Expected      interface{} `json:"expected"`
	ErrorHint     string      `json:"errorHint"`
}

type Manifest struct {
	Version int    `json:"version"`
	Count   int    `json:"count"`
	Cases   []Case `json:"cases"`
}

func deepEqual(a, b interface{}) bool {
	switch av := a.(type) {
	case nil:
		return b == nil
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		if !ok {
			return false
		}
		// integer-equal first
		if av == math.Trunc(av) && bv == math.Trunc(bv) {
			return av == bv
		}
		// tolerant
		if math.Abs(av-bv) <= 1e-12 {
			return true
		}
		return av == bv
	case int64:
		switch bv := b.(type) {
		case int64:
			return av == bv
		case float64:
			return float64(av) == bv
		}
		return false
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !deepEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			w, ok := bv[k]
			if !ok || !deepEqual(v, w) {
				return false
			}
		}
		return true
	}
	// Fallback via JSON
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return string(ja) == string(jb)
}

func main() {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..", "..", "cases.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read cases.json:", err)
		os.Exit(2)
	}
	var m Manifest
	dec := json.NewDecoder(jsonReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil {
		fmt.Fprintln(os.Stderr, "parse cases.json:", err)
		os.Exit(2)
	}

	fmt.Printf("loaded %d cases from %s\n", m.Count, path)

	var (
		total, passed, failed, errPassed, errFailed, skipped int
	)
	type Failure struct {
		ID, Msg string
	}
	var failures []Failure

	for _, c := range m.Cases {
		total++
		if c.SchemaDriven {
			skipped++
			continue
		}
		var got interface{}
		err := asun.Decode([]byte(c.Input), &got)
		if c.Kind == "ok" {
			if err != nil {
				failed++
				if len(failures) < 25 {
					failures = append(failures, Failure{c.ID, fmt.Sprintf("expected ok, got error: %v\n    input: %q", err, c.Input)})
				}
				continue
			}
			expected := normaliseJSON(c.Expected)
			actual := normaliseJSON(got)
			if !deepEqual(expected, actual) {
				failed++
				if len(failures) < 25 {
					failures = append(failures, Failure{c.ID, fmt.Sprintf("value mismatch\n    input:    %q\n    expected: %s\n    actual:   %s",
						c.Input, jsonStr(expected), jsonStr(actual))})
				}
				continue
			}
			passed++
		} else {
			if err != nil {
				errPassed++
			} else {
				errFailed++
				if len(failures) < 25 {
					failures = append(failures, Failure{c.ID, fmt.Sprintf("expected error, got ok: %s\n    input: %q",
						jsonStr(got), c.Input)})
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("================ ASUN-GO conformance ================")
	fmt.Printf("total                : %d\n", total)
	fmt.Printf("untyped ok-cases pass: %d\n", passed)
	fmt.Printf("untyped ok-cases fail: %d\n", failed)
	fmt.Printf("untyped err-cases pass: %d\n", errPassed)
	fmt.Printf("untyped err-cases fail: %d\n", errFailed)
	fmt.Printf("skipped (needs typed): %d\n", skipped)
	executed := total - skipped
	pct := 0.0
	if executed > 0 {
		pct = float64(passed+errPassed) / float64(executed) * 100.0
	}
	fmt.Printf("untyped pass rate    : %d/%d (%.1f%%)\n", passed+errPassed, executed, pct)
	fmt.Println("=====================================================")

	for _, f := range failures {
		fmt.Printf("\n[%s]\n    %s\n", f.ID, f.Msg)
	}

	// ---------- Encode (round-trip) ----------
	encPath := filepath.Join(wd, "..", "..", "encode-cases.json")
	encRaw, encErr := os.ReadFile(encPath)
	encFailed := 0
	if encErr == nil {
		var em struct {
			Cases []struct {
				ID    string      `json:"id"`
				Value interface{} `json:"value"`
			} `json:"cases"`
		}
		dec := json.NewDecoder(jsonReader(encRaw))
		dec.UseNumber()
		if err := dec.Decode(&em); err != nil {
			fmt.Fprintln(os.Stderr, "parse encode-cases.json:", err)
		} else {
			fmt.Printf("\nloaded %d encode cases from %s\n", len(em.Cases), encPath)
			encPassed := 0
			var encFailures []Failure
			for _, c := range em.Cases {
				val := normaliseJSON(c.Value)
				out, err := asun.Encode(val)
				if err != nil {
					encFailed++
					if len(encFailures) < 25 {
						encFailures = append(encFailures, Failure{c.ID, fmt.Sprintf("encode failed: %v\n    value: %s", err, jsonStr(val))})
					}
					continue
				}
				var got interface{}
				if err := asun.Decode(out, &got); err != nil {
					encFailed++
					if len(encFailures) < 25 {
						encFailures = append(encFailures, Failure{c.ID, fmt.Sprintf("decode-after-encode failed: %v\n    value:   %s\n    encoded: %s", err, jsonStr(val), string(out))})
					}
					continue
				}
				gotN := normaliseJSON(got)
				if !deepEqual(val, gotN) {
					encFailed++
					if len(encFailures) < 25 {
						encFailures = append(encFailures, Failure{c.ID, fmt.Sprintf("round-trip mismatch\n    value:   %s\n    encoded: %s\n    decoded: %s", jsonStr(val), string(out), jsonStr(gotN))})
					}
					continue
				}
				encPassed++
			}
			encTotal := encPassed + encFailed
			encPct := 0.0
			if encTotal > 0 {
				encPct = float64(encPassed) / float64(encTotal) * 100.0
			}
			fmt.Println()
			fmt.Println("============ ASUN-GO encode round-trip ==============")
			fmt.Printf("total : %d\n", encTotal)
			fmt.Printf("pass  : %d\n", encPassed)
			fmt.Printf("fail  : %d\n", encFailed)
			fmt.Printf("rate  : %d/%d (%.1f%%)\n", encPassed, encTotal, encPct)
			fmt.Println("=====================================================")
			for _, f := range encFailures {
				fmt.Printf("\n[%s]\n    %s\n", f.ID, f.Msg)
			}
		}
	}

	if failed > 0 || errFailed > 0 || encFailed > 0 {
		os.Exit(1)
	}
}

func jsonStr(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// normaliseJSON converts json.Number to float64/int64 and recurses
func normaliseJSON(v interface{}) interface{} {
	switch t := v.(type) {
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return i
		}
		f, _ := t.Float64()
		return f
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, vv := range t {
			out[k] = normaliseJSON(vv)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, vv := range t {
			out[i] = normaliseJSON(vv)
		}
		return out
	case int:
		return int64(t)
	}
	return v
}

// jsonReader wraps []byte as io.Reader without extra deps
type byteReader struct{ b []byte; i int }

func (r *byteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
func jsonReader(b []byte) *byteReader { return &byteReader{b: b} }
