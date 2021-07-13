(async function () {
    function getElementById(id: string): HTMLElement {
        const e = document.getElementById(id);
        if (e === null) {
            throw new Error(`#${id} element does not exist`);
        }
        return e;
    }

    const body = getElementById('lint-result-body');
    const errorMessage = getElementById('error-msg');
    const successMessage = getElementById('success-msg');
    const nowLoading = getElementById('loading');
    const checkUrlButton = getElementById('check-url-btn');
    const checkUrlInput = getElementById('check-url-input') as HTMLInputElement;

    async function getRemoteSource(url: string): Promise<string> {
        function getUrlToFetch(u: string): string {
            const url = new URL(u);

            // Convert repository URL to raw source URL
            if (url.host === 'github.com') {
                // Convert /owner/repo/blob/branch/path/to to /owner/repo/branch/path/to
                const s = url.pathname.split('/blob/');
                if (s.length === 2) {
                    url.pathname = s.join('/');
                    url.host = 'raw.githubusercontent.com';
                    return url.toString();
                }
            }

            // Convert Gist URL to raw source URL
            if (url.host === 'gist.github.com' && /\/[0-9a-f]+$/.test(url.pathname)) {
                url.host = 'gist.githubusercontent.com';
                url.pathname += '/raw';
                return url.toString();
            }

            return u;
        }

        const res = await fetch(getUrlToFetch(url));
        if (!res.ok) {
            throw new Error(`Fetching ${url} failed with status ${res.status}: ${res.statusText}`);
        }
        const src = await res.text();
        return src.trim();
    }

    async function getDefaultSource(): Promise<string> {
        const params = new URLSearchParams(window.location.search);

        const s = params.get('s');
        if (s !== null) {
            return s;
        }

        const u = params.get('u');
        if (u !== null) {
            return getRemoteSource(u);
        }

        const src = `# Paste your workflow YAML to this code editor

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

        return src;
    }

    const editor = CodeMirror(getElementById('editor'), {
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
        value: await getDefaultSource(),
    } as CodeMirror.EditorConfiguration);

    const debounceInterval = isMobile.phone ? 1000 : 300;
    let debounceId: number | null = null;
    let contentChanged = false;
    editor.on('change', function (_, e) {
        contentChanged = true;

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
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
    }

    function dismissLoading(): void {
        nowLoading.style.display = 'none';
    }

    const reUrl = /https?:\/\/\S+/;
    function linkifyMessage(text: string): HTMLElement[] {
        function span(text: string): HTMLSpanElement {
            const e = document.createElement('span');
            e.textContent = text;
            return e;
        }

        const ret: HTMLElement[] = [];
        let rest = text;
        while (true) {
            const m = rest.match(reUrl);
            if (m === null || m.index === undefined || m[0] === undefined) {
                if (rest.length > 0) {
                    ret.push(span(rest));
                }
                return ret;
            }

            const idx = m.index;
            const url = m[0];

            const s = rest.slice(0, idx);
            if (s.length > 0) {
                ret.push(span(s));
            }

            const a = document.createElement('a');
            a.href = url;
            a.rel = 'noopener';
            a.textContent = url;
            a.className = 'has-text-info-light is-underlined';
            a.addEventListener('click', e => e.stopPropagation());
            ret.push(a);

            rest = rest.slice(idx + url.length);
        }
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
            for (const elem of linkifyMessage(error.message)) {
                desc.appendChild(elem);
            }
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

    window.addEventListener('beforeunload', e => {
        if (contentChanged) {
            e.preventDefault();
            e.returnValue = '';
        }
    });

    checkUrlButton.addEventListener('click', async e => {
        e.preventDefault();
        const input = checkUrlInput.value;
        let src;
        try {
            src = await getRemoteSource(input);
        } catch (err) {
            showError(`Incorrect input "${input}": ${err.message}`);
            return;
        }
        editor.setValue(src);
        contentChanged = false;
    });

    const go = new Go();

    let result;
    // Note: WebAssembly.instantiateStreaming is not implemented on Safari yet
    if (typeof WebAssembly.instantiateStreaming === 'function') {
        result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
    } else {
        const response = await fetch('main.wasm');
        const mod = await response.arrayBuffer();
        result = await WebAssembly.instantiate(mod, go.importObject);
    }

    await go.run(result.instance);
})().catch(err => {
    console.error('ERROR!:', err);
    alert(`${err.name}: ${err.message}\n\n${err.stack}`);
});
