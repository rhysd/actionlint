(function() {
    const editor = CodeMirror(document.getElementById('editor'), {
        mode: 'yaml',
        lineNumbers: true,
        lineWrapping: true,
        autofocus: true,
        styleActiveLine: true,
        value:
`on:
  push:
    branch: main

jobs:
  test:
    strategy:
      matrix:
        os: [macos-latest, linux-latest]
    runs-on: \${{ matrix.os }}
    steps:
      - uses: actions/checkout@v2
      - uses: actions/cache@v2
        with:
          path: ~/.npm
          key: \${{ matrix.platform }}-node-\${{ hashFiles('**/package-lock.json') }}
        if: \${{ github.repository.permissions.admin == true }}
      - run: npm install && npm test`,
    });
    const debounceInterval = 300; // TODO: Change interval looking at desktop or mobile
    let debounceId = null;
    editor.on('change', function() {
        if (typeof window.runActionlint !== 'function') {
            return;
        }

        if (debounceId !== null) {
            window.clearTimeout(debounceId);
        }

        debounceId = window.setTimeout(() => {
            debounceId = null;
            const src = editor.getValue();
            window.runActionlint(src);
        }, debounceInterval);
    });

    const body = document.getElementById('lint-result-body');
    const errorMessage = document.getElementById('error-msg');

    function getSource() {
        return editor.getValue();
    }

    function onCheckFailed(message) {
        console.error('Check failed!:', message);
        errorMessage.textContent = message;
    }

    function onCheckCompleted(errors) {
        function td(row, text) {
            const e = document.createElement('td');
            e.textContent = text;
            row.appendChild(e);
        }

        body.textContent = '';
        for (const error of errors) {
            const row = document.createElement('tr');
            td(row, `${error.line}:${error.column}`);
            td(row, error.message);
            td(row, error.kind);
            body.appendChild(row);
        }
    }

    window.getYamlSource = getSource;
    window.onCheckFailed = onCheckFailed;
    window.onCheckCompleted = onCheckCompleted;

    async function main() {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject);
        await go.run(result.instance); // This function will never return
    }

    main().catch(err => console.error('ERROR!:', err));
})();
