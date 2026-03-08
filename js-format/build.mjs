// build.mjs — esbuild bundler for ason-format
import * as esbuild from 'esbuild';
import { cpSync, mkdirSync } from 'fs';

mkdirSync('dist', { recursive: true });

// ESM bundle
await esbuild.build({
  entryPoints: ['src/index.ts'],
  bundle: true,
  platform: 'neutral',
  format: 'esm',
  outfile: 'dist/ason-format.js',
  minify: false,
});

// CJS bundle
await esbuild.build({
  entryPoints: ['src/index.ts'],
  bundle: true,
  platform: 'node',
  format: 'cjs',
  outfile: 'dist/ason-format.cjs',
  minify: false,
});

// Minified IIFE for CDN (<script> tag usage)
await esbuild.build({
  entryPoints: ['src/index.ts'],
  bundle: true,
  platform: 'browser',
  format: 'iife',
  globalName: 'AsonFormat',
  outfile: 'dist/ason-format.min.js',
  minify: true,
});

// Copy CSS
cpSync('src/ason-format.css', 'dist/ason-format.css');

console.log('✓ dist/ason-format.js  dist/ason-format.cjs  dist/ason-format.min.js  dist/ason-format.css');
