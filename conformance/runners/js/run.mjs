// Conformance runner for asun-js.
// Loads ../../cases.json and decodes each case via decode().

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '..', '..', '..');
const distEntry = path.join(repoRoot, 'asun-js', 'dist', 'index.js');
const { decode, encode } = await import(pathToFileURL(distEntry).href);

const casesPath = path.resolve(__dirname, '..', '..', 'cases.json');
const manifest = JSON.parse(fs.readFileSync(casesPath, 'utf-8'));
console.log(`loaded ${manifest.cases.length} cases from ${casesPath}`);

function deepEqual(a, b) {
  if (a === b) return true;
  if (a === null || b === null) return a === b;
  if (typeof a === 'number' && typeof b === 'number') {
    if (Number.isInteger(a) && Number.isInteger(b)) return a === b;
    if (Math.abs(a - b) <= 1e-12) return true;
    return a === b;
  }
  if (typeof a !== typeof b) return false;
  if (Array.isArray(a)) {
    if (!Array.isArray(b) || a.length !== b.length) return false;
    return a.every((x, i) => deepEqual(x, b[i]));
  }
  if (typeof a === 'object') {
    const ak = Object.keys(a), bk = Object.keys(b);
    if (ak.length !== bk.length) return false;
    return ak.every((k) => deepEqual(a[k], b[k]));
  }
  return false;
}

let total = 0, passed = 0, failed = 0, errPassed = 0, errFailed = 0, skipped = 0;
const failures = [];

for (const c of manifest.cases) {
  total++;
  if (c.schemaDriven) { skipped++; continue; }
  let got, threw = false, errMsg = '';
  try {
    got = decode(c.input);
  } catch (e) {
    threw = true;
    errMsg = `${e.name}: ${e.message}`;
  }
  if (c.kind === 'ok') {
    if (threw) {
      failed++;
      if (failures.length < 25) failures.push([c.id, `expected ok, got error: ${errMsg}\n    input: ${JSON.stringify(c.input)}`]);
      continue;
    }
    if (!deepEqual(got, c.expected)) {
      failed++;
      if (failures.length < 25) failures.push([c.id, `value mismatch\n    input:    ${JSON.stringify(c.input)}\n    expected: ${JSON.stringify(c.expected)}\n    actual:   ${JSON.stringify(got)}`]);
      continue;
    }
    passed++;
  } else {
    if (threw) errPassed++;
    else {
      errFailed++;
      if (failures.length < 25) failures.push([c.id, `expected error, got ok: ${JSON.stringify(got)}\n    input: ${JSON.stringify(c.input)}`]);
    }
  }
}

console.log();
console.log('================ ASUN-JS conformance ================');
console.log(`total                : ${total}`);
console.log(`untyped ok-cases pass: ${passed}`);
console.log(`untyped ok-cases fail: ${failed}`);
console.log(`untyped err-cases pass: ${errPassed}`);
console.log(`untyped err-cases fail: ${errFailed}`);
console.log(`skipped (needs typed): ${skipped}`);
const exec = total - skipped;
const pct = exec > 0 ? ((passed + errPassed) / exec * 100).toFixed(1) : '0.0';
console.log(`untyped pass rate    : ${passed + errPassed}/${exec} (${pct}%)`);
console.log('=====================================================');

for (const [id, msg] of failures) console.log(`\n[${id}]\n    ${msg}`);

// ---------- Encode (round-trip) ----------
const encPath = path.resolve(__dirname, '..', '..', 'encode-cases.json');
let encFailed = 0;
if (fs.existsSync(encPath)) {
  const em = JSON.parse(fs.readFileSync(encPath, 'utf-8'));
  console.log(`\nloaded ${em.cases.length} encode cases from ${encPath}`);
  let encPassed = 0;
  const encFailures = [];
  for (const c of em.cases) {
    const val = c.value;
    let text;
    try {
      text = encode(val);
    } catch (e) {
      encFailed++;
      if (encFailures.length < 25) encFailures.push([c.id, `encode failed: ${e.name}: ${e.message}\n    value: ${JSON.stringify(val)}`]);
      continue;
    }
    let got;
    try {
      got = decode(text);
    } catch (e) {
      encFailed++;
      if (encFailures.length < 25) encFailures.push([c.id, `decode-after-encode failed: ${e.name}: ${e.message}\n    value:   ${JSON.stringify(val)}\n    encoded: ${JSON.stringify(text)}`]);
      continue;
    }
    if (!deepEqual(val, got)) {
      encFailed++;
      if (encFailures.length < 25) encFailures.push([c.id, `round-trip mismatch\n    value:   ${JSON.stringify(val)}\n    encoded: ${JSON.stringify(text)}\n    decoded: ${JSON.stringify(got)}`]);
      continue;
    }
    encPassed++;
  }
  const encTotal = encPassed + encFailed;
  const encPct = encTotal > 0 ? (encPassed / encTotal * 100).toFixed(1) : '0.0';
  console.log();
  console.log('============ ASUN-JS encode round-trip ==============');
  console.log(`total : ${encTotal}`);
  console.log(`pass  : ${encPassed}`);
  console.log(`fail  : ${encFailed}`);
  console.log(`rate  : ${encPassed}/${encTotal} (${encPct}%)`);
  console.log('=====================================================');
  for (const [id, msg] of encFailures) console.log(`\n[${id}]\n    ${msg}`);
}

process.exit((failed > 0 || errFailed > 0 || encFailed > 0) ? 1 : 0);
