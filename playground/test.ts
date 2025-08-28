import { JSDOM } from 'jsdom';
import { promises as fs } from 'fs';
import { strict as assert } from 'assert';
import { Crypto } from '@peculiar/webcrypto';

// This polyfill is necessary for Node.js v18 or earlier. `global.crypto` was added at v19.
// https://github.com/nodejs/node/pull/42083/files
if (typeof globalThis.crypto === 'undefined') {
    globalThis.crypto = new Crypto();
}

// Inject global.Go for testing `main.wasm`.
require('./lib/js/wasm_exec.js'); // eslint-disable-line @typescript-eslint/no-require-imports

class CheckResults {
    errors: ActionlintError[] | null = null;
    resolve: ((errs: ActionlintError[]) => void) | null = null;

    onCheckCompleted(errs: ActionlintError[]) {
        this.errors = errs;
        if (this.resolve !== null) {
            this.resolve(errs);
            this.resolve = null;
        }
    }

    waitCheckCompleted(): Promise<ActionlintError[]> {
        return new Promise(resolve => {
            if (this.errors !== null) {
                resolve(this.errors);
                return;
            }
            this.resolve = resolve;
        });
    }

    reset(): void {
        this.errors = null;
    }
}

describe('main.wasm', function () {
    const results = new CheckResults();

    before(async function () {
        const dom = new JSDOM('');
        dom.window.dismissLoading = function () {
            /*do nothing*/
        };
        dom.window.getYamlSource = function () {
            return `
on: push

jobs:
  test:
    steps:
      - run: echo 'hi'`;
        };
        dom.window.onCheckCompleted = results.onCheckCompleted.bind(results);

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        global.window = dom.window as any;

        const go = new Go();
        const bin = await fs.readFile('./main.wasm');
        const result = await WebAssembly.instantiate(bin, go.importObject);

        // Do not `await` this method call since it will never be settled
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        go.run('instance' in result ? (result as { instance: WebAssembly.Instance }).instance : result);
    });

    it('shows first result on loading', async function () {
        const errors = await results.waitCheckCompleted();

        const json = JSON.stringify(errors);
        assert.equal(errors.length, 1, json);

        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        const err = errors[0]!;
        assert.equal(err.message, '"runs-on" section is missing in job "test"', `message is unexpected: ${json}`);
        assert.equal(err.line, 5, `line is unexpected: ${json}`);
        assert.equal(err.column, 3, `column is unexpected: ${json}`);
        assert.equal(err.kind, 'syntax-check', `kind is unexpected: ${json}`);
    });

    it('reports some errors by running actionlint with runActionlint', async function () {
        assert.ok(window.runActionlint);
        results.reset();

        const source = `
on: foo

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hi'`;

        window.runActionlint(source);
        const errors = await results.waitCheckCompleted();
        const json = JSON.stringify(errors);
        assert.equal(errors.length, 1, json);

        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        const err = errors[0]!;
        assert.ok(err.message.includes('unknown Webhook event "foo"'), `message is unexpected: ${json}`);
        assert.equal(err.line, 2, `line is unexpected: ${json}`);
        assert.equal(err.column, 5, `column is unexpected: ${json}`);
        assert.equal(err.kind, 'events', `kind is unexpected: ${json}`);
    });

    it('reports no error by running actionlint with runActionlint', async function () {
        assert.ok(window.runActionlint);
        results.reset();

        const source = `
on: push

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hi'`;

        window.runActionlint(source);
        const errors = await results.waitCheckCompleted();
        const json = JSON.stringify(errors);
        assert.equal(errors.length, 0, json);
    });
});
