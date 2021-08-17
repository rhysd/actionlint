const path = require('path');
const fs = require('fs');
const assert = require('assert').strict;
const pattern = require('./object.js').problemMatcher[0].pattern[0];
const regexp = new RegExp(pattern.regexp);

{
    const want = require('./test/want.json')[0];

    for (const file of ['escape.txt', 'no_escape.txt']) {
        console.log(`Testing test/${file}`);
        const escaped = path.join(__dirname, 'test', file);
        const lines = fs.readFileSync(escaped, 'utf8').split('\n');
        const m = lines[0].match(regexp);
        assert.ok(m); // not null
        assert.equal(want.filepath, m[pattern.file]);
        assert.equal(want.line.toString(), m[pattern.line]);
        assert.equal(want.column.toString(), m[pattern.column]);
        assert.equal(want.message, m[pattern.message]);
        assert.equal(want.kind, m[pattern.code]);
        console.log(`Success test/${file}`);
    }
}

{
    const dir = path.join(__dirname, '..', '..', 'testdata', 'examples');
    for (const name of fs.readdirSync(dir)) {
        if (!name.endsWith('.out')) {
            continue;
        }
        console.log(`Testing testdata/examples/${name}`);
        for (const line of fs.readFileSync(path.join(dir, name), 'utf8').split('\n')) {
            if (line.length === 0 || line.startsWith('/')) {
                continue
            }
            const msg = `Line '${line}' did not match to the regex`
            const m = line.match(regexp);
            assert.ok(m, msg); // not null
            assert.equal('test.yaml', m[pattern.file], msg);
            assert.match(m[pattern.line], /^\d+$/, msg);
            assert.match(m[pattern.column], /^\d+$/, msg);
            assert.ok(m[pattern.message].length > 0, msg);
            assert.ok(m[pattern.code].length > 0, msg);
        }
        console.log(`Success testdata/examples/${name}`);
    }
}
