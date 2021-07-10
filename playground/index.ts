(function () {
    function ensureNonNull<T>(x: T | null): T {
        if (x === null) {
            throw new Error('Unexpected null value');
        }
        return x;
    }

    const body = ensureNonNull(document.getElementById('lint-result-body'));
    const errorMessage = ensureNonNull(document.getElementById('error-msg'));
    const successMessage = ensureNonNull(document.getElementById('success-msg'));
    const nowLoading = ensureNonNull(document.getElementById('loading'));

    function getDefaultSource(): string {
        const p = new URLSearchParams(window.location.search).get('s');
        if (p !== null) {
            return p;
        }

        return `# Paste your workflow YAML to this code editor

on:
  push:
    branch: main
    tags:
      - 'v\\d+'

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
      - run: npm install && npm test`;
    }

    const editor = CodeMirror(ensureNonNull(document.getElementById('editor')), {
        mode: 'yaml',
        theme: 'material-darker',
        lineNumbers: true,
        lineWrapping: true,
        autofocus: true,
        styleActiveLine: true,
        gutters: ['CodeMirror-linenumbers', 'error-marker'],
        extraKeys: {
            Tab(cm) {
                if (cm.somethingSelected()) {
                    cm.execCommand('indentMore');
                } else {
                    cm.execCommand('insertSoftTab');
                }
            },
        },
        value: getDefaultSource(),
    } as CodeMirror.EditorConfiguration);

    const debounceInterval = isMobile.phone ? 1000 : 300;
    let debounceId: number | null = null;
    editor.on('change', function (_, e) {
        if (typeof window.runActionlint !== 'function') {
            showError('Preparing Wasm file is not completed yet. Please wait for a while and try again.');
            return;
        }

        if (debounceId !== null) {
            window.clearTimeout(debounceId);
        }

        function startActionlint(): void {
            debounceId = null;
            errorMessage.style.display = 'none';
            successMessage.style.display = 'none';
            editor.clearGutter('error-marker');
            // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            window.runActionlint!(editor.getValue());
        }

        if (e.origin === 'paste') {
            startActionlint(); // When pasting some code, apply actionlint instantly
            return;
        }

        debounceId = window.setTimeout(() => startActionlint(), debounceInterval);
    });

    function getSource(): string {
        return editor.getValue();
    }

    function showError(message: string): void {
        console.error('Check failed!:', message);
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
    }

    function dismissLoading(): void {
        nowLoading.style.display = 'none';
    }

    function onCheckCompleted(errors: ActionlintError[]): void {
        body.textContent = '';

        if (errors.length === 0) {
            successMessage.style.display = 'block';
            return;
        }

        for (const error of errors) {
            const row = document.createElement('tr');
            row.className = 'is-size-5';
            row.addEventListener('click', () => {
                editor.setCursor({ line: error.line - 1, ch: error.column - 1 });
                editor.focus();
            });

            const pos = document.createElement('td');
            const tag = document.createElement('span');
            tag.className = 'tag is-primary is-dark';
            tag.textContent = `line:${error.line}, col:${error.column}`;
            pos.appendChild(tag);
            row.appendChild(pos);

            const desc = document.createElement('td');
            const msg = document.createElement('span');
            msg.textContent = error.message;
            desc.appendChild(msg);
            const kind = document.createElement('span');
            kind.className = 'tag is-dark';
            kind.textContent = error.kind;
            kind.style.marginLeft = '4px';
            desc.appendChild(kind);
            row.appendChild(desc);

            body.appendChild(row);

            const marker = document.createElement('div');
            marker.style.color = '#ff5370';
            marker.textContent = '●';
            editor.setGutterMarker(error.line - 1, 'error-marker', marker);
        }
    }

    window.getYamlSource = getSource;
    window.showError = showError;
    window.onCheckCompleted = onCheckCompleted;
    window.dismissLoading = dismissLoading;

    async function main(): Promise<void> {
        const go = new Go();
        let result;
        // Note: WebAssembly.instantiateStreaming is not implemented on Safari yet
        if (typeof WebAssembly.instantiateStreaming == 'function') {
            result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
        } else {
            const response = await fetch('main.wasm');
            const mod = await response.arrayBuffer();
            result = await WebAssembly.instantiate(mod, go.importObject);
        }
        await go.run(result.instance);
    }

    main().catch(err => {
        console.error('ERROR!:', err);
        alert(`${err.name}: ${err.message}\n\n${err.stack}`);
    });
})();
