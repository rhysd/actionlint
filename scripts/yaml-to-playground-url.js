#!/usr/bin/env node

// Usage:
//   pbpaste| node ./scripts/yaml-to-playground-url.js
//   node ./scripts/yaml-to-playground-url.js < test.yaml

const fs = require('fs');
const pako = require('../playground/node_modules/pako');

const re = /^\s*#/;
const stdin = fs.readFileSync(process.stdin.fd, 'utf8').trim();
const lines = stdin.split('\n').filter(l => !re.test(l)); // remove comment lines
const src = lines.join('\n');
const compressed = pako.deflate(new TextEncoder().encode(src));
const b64 = Buffer.from(compressed).toString('base64');
console.log(`https://rhysd.github.io/actionlint#${b64}`);
