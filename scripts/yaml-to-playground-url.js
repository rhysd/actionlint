#!/usr/bin/env node

const fs = require('fs');

const re = /^\s*#/;
const stdin = fs.readFileSync(process.stdin.fd, 'utf8').trim();
const lines = stdin.split('\n').filter(l => !re.test(l)); // remove comment lines
const params = new URLSearchParams();
params.set('s', lines.join('\n'));
console.log(`https://rhysd.github.io/actionlint/?${params.toString()}`);
