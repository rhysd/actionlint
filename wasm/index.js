(function() {
    const editor = CodeMirror(document.getElementById('editor'), {
        mode: 'yaml',
        lineNumbers: true,
        lineWrapping: true,
        autofocus: true,
        styleActiveLine: true,
        extraKeys: {
            Tab(cm) {
                if (cm.somethingSelected()) {
                    cm.execCommand('indentMore');
                } else {
                    cm.execCommand('insertSoftTab');
                }
            },
        },
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
            errorMessage.style.display = 'none';
            window.runActionlint(editor.getValue());
        }, debounceInterval);
    });

    const body = document.getElementById('lint-result-body');
    const errorMessage = document.getElementById('error-msg');

    function getSource() {
        return editor.getValue();
    }

    function showError(message) {
        console.error('Check failed!:', message);
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
    }

    function onCheckCompleted(errors) {
        body.textContent = '';
        for (const error of errors) {
            const row = document.createElement('tr');
            row.addEventListener('click', () => {
                console.log(editor.getCursor(), error);
                editor.setCursor({line: error.line-1, ch: error.column-1});
                editor.focus();
            });

            const pos = document.createElement('td');
            const tag = document.createElement('span');
            tag.className = 'tag is-info is-light';
            tag.textContent = `line:${error.line}, col:${error.column}`;
            pos.appendChild(tag);
            row.appendChild(pos);

            const desc = document.createElement('td');
            const msg = document.createElement('span');
            msg.textContent = error.message;
            desc.appendChild(msg);
            const kind = document.createElement('span');
            kind.className = 'tag is-light';
            kind.textContent = error.kind;
            kind.style.marginLeft = '4px';
            desc.appendChild(kind);
            row.appendChild(desc);

            body.appendChild(row);
        }
    }

    window.getYamlSource = getSource;
    window.showError = showError;
    window.onCheckCompleted = onCheckCompleted;

    async function main() {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject);
        await go.run(result.instance); // This function will never return
    }

    main().catch(err => console.error('ERROR!:', err));
})();
