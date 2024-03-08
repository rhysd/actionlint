declare module 'eslint-plugin-mocha' {
    import type { ESLint, Linter } from 'eslint';

    interface Configs {
        configs: {
            recommended: Linter.Config;
            all: Linter.Config;
            flat: {
                recommended: Linter.FlatConfig;
                all: Linter.FlatConfig;
            };
        };
    }

    const plugin: Configs & ESLint.Plugin;
    export default plugin;
}
