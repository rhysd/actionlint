// @ts-check

import eslint from '@eslint/js';
import ts from 'typescript-eslint';

export default ts.config(
    eslint.configs.recommended,
    ...ts.configs.recommendedTypeChecked,
    {
        languageOptions: {
            parserOptions: {
                project: true,
                tsconfigRootDir: import.meta.dirname,
            },
        },
    },
    {
        files: ['index.ts', 'test.ts'],
        rules: {
            indent: ['error', 4],
            quotes: ['error', 'single'],
            'linebreak-style': ['error', 'unix'],
            semi: ['error', 'always'],
            eqeqeq: ['error', 'always'],
            'no-constant-condition': ['error', { checkLoops: false }],
            '@typescript-eslint/no-unsafe-member-access': 'off',
            '@typescript-eslint/no-unsafe-argument': 'off',
            '@typescript-eslint/no-unsafe-assignment': 'off',
        },
    },
    {
        files: ['test.ts'],
        rules: {
            '@typescript-eslint/unbound-method': 'off', // For checking `window.runActionlint`
        },
    }
);
