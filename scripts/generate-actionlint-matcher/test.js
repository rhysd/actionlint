const path = require('path');
const fs = require('fs');
const assert = require('assert').strict;
const pattern = require('./object.js').problemMatcher[0].pattern[0];
const regexp = new RegExp(pattern.regexp);
const want = require('./test/want.json')[0];

for (const file of ['escape.txt', 'no_escape.txt']) {
    console.log(`Testing ${file}`);
    const escaped = path.join(__dirname, 'test', file);
    const lines = fs.readFileSync(escaped, 'utf8').split('\n');
    const m = lines[0].match(regexp);
    assert.ok(m); // not null
    assert.equal(want.filepath, m[pattern.file]);
    assert.equal(want.line.toString(), m[pattern.line]);
    assert.equal(want.column.toString(), m[pattern.column]);
    assert.equal(want.message, m[pattern.message]);
    assert.equal(want.kind, m[pattern.code]);
    console.log(`Success! ${file}`);
}
