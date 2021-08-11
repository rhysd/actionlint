const path = require('path');
const fs = require('fs');
const assert = require('assert').strict;
const pattern = require('./object.js').problemMatcher[0].pattern[0];
const regexp = new RegExp(pattern.regexp);

{
    const escaped = path.join(__dirname, 'test', 'escape.txt');
    const lines = fs.readFileSync(escaped, 'utf8').split('\n');
    const m = lines[0].match(regexp);
    assert.ok(m); // not null
    assert.equal(m[pattern.file], 'test.yaml');
    assert.equal(m[pattern.line], '6');
    assert.equal(m[pattern.column], '41');
    assert.equal(m[pattern.message], '"github.event.head_commit.message" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://securitylab.github.com/research/github-actions-untrusted-input for more details');
    assert.equal(m[pattern.code], 'expression');
}

{
    const escaped = path.join(__dirname, 'test', 'no_escape.txt');
    const lines = fs.readFileSync(escaped, 'utf8').split('\n');
    const m = lines[0].match(regexp);
    assert.ok(m); // not null
    assert.equal(m[pattern.file], 'test.yaml');
    assert.equal(m[pattern.line], '6');
    assert.equal(m[pattern.column], '41');
    assert.equal(m[pattern.message], '"github.event.head_commit.message" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://securitylab.github.com/research/github-actions-untrusted-input for more details');
    assert.equal(m[pattern.code], 'expression');
}
