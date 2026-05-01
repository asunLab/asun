<?php
// Conformance runner for asun-php.
//
// Loads ../../cases.json and decodes each case via asun_decode().
// Then runs ../../encode-cases.json round-trip (encode → decode → equal).

declare(strict_types=1);

$root      = realpath(__DIR__ . '/../../..');
$casesPath = realpath(__DIR__ . '/../../cases.json');
$encPath   = realpath(__DIR__ . '/../../encode-cases.json');
$ext       = $root . '/asun-php/modules/asun.so';

if (!extension_loaded('asun')) {
    if (!is_file($ext)) {
        fwrite(STDERR, "asun.so not found at $ext — build asun-php first\n");
        exit(2);
    }
    if (!dl(basename($ext))) {
        // dl() is usually disabled — caller must `php -dextension=…`.
        fwrite(STDERR, "asun extension is not loaded.\n");
        fwrite(STDERR, "Re-run with:  php -dextension=$ext " . __FILE__ . "\n");
        exit(2);
    }
}

/**
 * JSON-decode produces:
 *   - assoc arrays when the JSON shape is an object
 *   - numeric-indexed arrays when the JSON shape is a list
 *
 * `asun_decode` should return the same PHP shapes for tuples/structs/lists.
 *
 * We compare them structurally with the same numeric tolerance the other runners use.
 */
function deep_equal($a, $b): bool {
    if (is_bool($a) || is_bool($b)) {
        return is_bool($a) && is_bool($b) && $a === $b;
    }
    if ($a === null || $b === null) {
        return $a === null && $b === null;
    }
    if (is_int($a) && is_int($b))     return $a === $b;
    if ((is_int($a) || is_float($a)) && (is_int($b) || is_float($b))) {
        $fa = (float)$a; $fb = (float)$b;
        if ($fa === $fb) return true;
        // Tolerant numeric compare matching the other runners: combine an
        // absolute floor (for tiny / near-zero values) with a relative
        // tolerance scaled by the larger magnitude (for large / extreme floats).
        $tol = max(1e-12, max(abs($fa), abs($fb)) * 1e-12);
        return abs($fa - $fb) <= $tol;
    }
    if (is_string($a) && is_string($b)) return $a === $b;
    if (is_array($a) && is_array($b)) {
        $aIsList = array_is_list($a);
        $bIsList = array_is_list($b);
        if ($aIsList !== $bIsList) return false;
        if ($aIsList) {
            if (count($a) !== count($b)) return false;
            foreach ($a as $i => $v) if (!deep_equal($v, $b[$i])) return false;
            return true;
        }
        if (count($a) !== count($b)) return false;
        foreach ($a as $k => $v) {
            if (!array_key_exists($k, $b)) return false;
            if (!deep_equal($v, $b[$k])) return false;
        }
        return true;
    }
    return $a === $b;
}

function pp($v): string {
    return json_encode($v, JSON_UNESCAPED_UNICODE | JSON_PARTIAL_OUTPUT_ON_ERROR);
}

// ---------- Decode cases ----------
$manifest = json_decode(file_get_contents($casesPath), true, flags: JSON_THROW_ON_ERROR);
$cases    = $manifest['cases'];
echo "loaded " . count($cases) . " cases from $casesPath\n";

$total = $passed = $failed = $errPassed = $errFailed = $skipped = 0;
$failures = [];

foreach ($cases as $c) {
    $total++;
    if (!empty($c['schemaDriven'])) { $skipped++; continue; }

    $okRun = true;
    $errMsg = '';
    $got = null;
    try {
        $got = asun_decode($c['input']);
    } catch (\Throwable $e) {
        $okRun = false;
        $errMsg = get_class($e) . ': ' . $e->getMessage();
    }

    if ($c['kind'] === 'ok') {
        if (!$okRun) {
            $failed++;
            if (count($failures) < 25) {
                $failures[] = [$c['id'], "expected ok, got error: $errMsg\n    input: " . pp($c['input'])];
            }
            continue;
        }
        $expected = $c['expected'];
        if (!deep_equal($got, $expected)) {
            $failed++;
            if (count($failures) < 25) {
                $failures[] = [$c['id'], "value mismatch\n    input:    " . pp($c['input']) . "\n    expected: " . pp($expected) . "\n    actual:   " . pp($got)];
            }
            continue;
        }
        $passed++;
    } else {
        if (!$okRun) {
            $errPassed++;
        } else {
            $errFailed++;
            if (count($failures) < 25) {
                $failures[] = [$c['id'], "expected error, got ok: " . pp($got) . "\n    input: " . pp($c['input'])];
            }
        }
    }
}

echo "\n";
echo "================ ASUN-PHP conformance ================\n";
echo "total                : $total\n";
echo "untyped ok-cases pass: $passed\n";
echo "untyped ok-cases fail: $failed\n";
echo "untyped err-cases pass: $errPassed\n";
echo "untyped err-cases fail: $errFailed\n";
echo "skipped (needs typed): $skipped\n";
$executed = $total - $skipped;
$pct = $executed ? (($passed + $errPassed) / $executed * 100.0) : 0.0;
printf("untyped pass rate    : %d/%d (%.1f%%)\n", $passed + $errPassed, $executed, $pct);
echo "======================================================\n";

foreach ($failures as [$fid, $msg]) {
    echo "\n[$fid]\n    $msg\n";
}

// ---------- Encode round-trip ----------
$encFailed = 0;
if (is_string($encPath) && is_file($encPath)) {
    $em = json_decode(file_get_contents($encPath), true, flags: JSON_THROW_ON_ERROR);
    $encCases = $em['cases'];
    echo "\nloaded " . count($encCases) . " encode cases from $encPath\n";
    $encPassed = 0;
    $encFailures = [];
    foreach ($encCases as $c) {
        $val = $c['value'];
        try {
            $text = asun_encode($val);
        } catch (\Throwable $e) {
            $encFailed++;
            if (count($encFailures) < 25) {
                $encFailures[] = [$c['id'], "encode failed: " . get_class($e) . ': ' . $e->getMessage() . "\n    value: " . pp($val)];
            }
            continue;
        }
        try {
            $got = asun_decode($text);
        } catch (\Throwable $e) {
            $encFailed++;
            if (count($encFailures) < 25) {
                $encFailures[] = [$c['id'], "decode-after-encode failed: " . get_class($e) . ': ' . $e->getMessage() . "\n    value:   " . pp($val) . "\n    encoded: " . pp($text)];
            }
            continue;
        }
        if (!deep_equal($val, $got)) {
            $encFailed++;
            if (count($encFailures) < 25) {
                $encFailures[] = [$c['id'], "round-trip mismatch\n    value:   " . pp($val) . "\n    encoded: " . pp($text) . "\n    decoded: " . pp($got)];
            }
            continue;
        }
        $encPassed++;
    }
    $encTotal = $encPassed + $encFailed;
    $encPct = $encTotal ? ($encPassed / $encTotal * 100.0) : 0.0;
    echo "\n";
    echo "============ ASUN-PHP encode round-trip ==============\n";
    echo "total : $encTotal\n";
    echo "pass  : $encPassed\n";
    echo "fail  : $encFailed\n";
    printf("rate  : %d/%d (%.1f%%)\n", $encPassed, $encTotal, $encPct);
    echo "======================================================\n";
    foreach ($encFailures as [$fid, $msg]) {
        echo "\n[$fid]\n    $msg\n";
    }
}

exit(($failed === 0 && $errFailed === 0 && $encFailed === 0) ? 0 : 1);
