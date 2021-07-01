(function() {
    function getSource() {
        return document.getElementById('editor').value;
    }

    function onCheckFailed(message) {
        console.error('Check failed!:', message);
        document.getElementById('error-msg').textContent = message;
    }

    function onCheckCompleted(errors) {
        console.log('Checked!:', errors);

        function td(row, text) {
            const e = document.createElement('td');
            e.textContent = text;
            row.appendChild(e);
        }

        const body = document.getElementById('lint-result-body');
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
