const path = require('path');
const fs = require('fs');
const assert = require('assert').strict;
const pattern = require('./object.js').problemMatcher[0].pattern[0];
const regexp = new RegExp(pattern.regexp);

const filepath = './testdata/err/one_error.yaml';
const line = '6';
const column = '41';
const message = '"github.event.head_commit.message" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://securitylab.github.com/research/github-actions-untrusted-input for more details';
const code = 'expression';

{
    const escaped = path.join(__dirname, 'test', 'escape.txt');
    const lines = fs.readFileSync(escaped, 'utf8').split('\n');
    const m = lines[0].match(regexp);
    assert.ok(m); // not null
    assert.equal(filepath, m[pattern.file]);
    assert.equal(line, m[pattern.line]);
    assert.equal(column, m[pattern.column]);
    assert.equal(message, m[pattern.message]);
    assert.equal(code, m[pattern.code]);
}

{
    const escaped = path.join(__dirname, 'test', 'no_escape.txt');
    const lines = fs.readFileSync(escaped, 'utf8').split('\n');
    const m = lines[0].match(regexp);
    assert.ok(m); // not null
    assert.equal(filepath, m[pattern.file]);
    assert.equal(line, m[pattern.line]);
    assert.equal(column, m[pattern.column]);
    assert.equal(message, m[pattern.message]);
    assert.equal(code, m[pattern.code]);
}
