import path from 'node:path';
import fs from 'node:fs';
import assert from 'node:assert';
import { fileURLToPath } from 'node:url';
import object from './object.mjs';

const pattern = object.problemMatcher[0].pattern[0];
const regexp = new RegExp(pattern.regexp);
const dirname = path.dirname(fileURLToPath(import.meta.url));

{
    const file = path.join(dirname, 'test', 'want.json');
    const want = JSON.parse(fs.readFileSync(file, 'utf8'))[0];

    for (const file of ['escape.txt', 'no_escape.txt']) {
        console.log(`Testing test/${file}`);
        const escaped = path.join(dirname, 'test', file);
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

for (const parent of ['examples', 'err']) {
    const dir = path.join(dirname, '..', '..', 'testdata', parent);
    for (const name of fs.readdirSync(dir)) {
        if (!name.endsWith('.out')) {
            continue;
        }
        console.log(`Testing testdata/${parent}/${name}`);
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
        console.log(`Success testdata/${parent}/${name}`);
    }
}
