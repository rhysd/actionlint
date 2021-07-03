Playground for actionlint
=========================

This is a development directory for [actionlint playground](https://rhysd.github.io/actionlint/).

The playground is built with vanilla HTML/CSS/JavaScript/Wasm. All dependencies are defined in `package.json` and managed
by `npm`. Tasks for development are defined in [`Makefile`](./Makefile).

## Tasks

```sh
# Install dependencies, build main.wasm, start serving the app at localhost:1234 using Python
make

# Install dependencies, build main.wasm
make build

# Install dependencies
make install

# Clean all built files and dependencies
make clean
```

## Deployment

Deployment is automated by [`deploy.bash`](./deploy.bash). See [CONTRIBUTING.md](../CONTRIBUTING.md) for more details.

