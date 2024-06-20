// @ts-check

import eslint from '@eslint/js';
import ts from 'typescript-eslint';
import mocha from 'eslint-plugin-mocha';

export default ts.config(
    eslint.configs.recommended,
    ...ts.configs.recommendedTypeChecked,
    {
        files: ['*.ts'],
        languageOptions: {
            parserOptions: {
                project: 'tsconfig.json',
            },
        },
    },
    {
        files: ['*.ts', '*.mjs'],
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
    mocha.configs.flat.recommended,
    {
        files: ['test.ts'],
        rules: {
            '@typescript-eslint/unbound-method': 'off', // For checking `window.runActionlint`
            'mocha/no-exclusive-tests': 'error',
            'mocha/no-pending-tests': 'error',
            'mocha/no-skipped-tests': 'error',
            'mocha/no-top-level-hooks': 'error',
        },
    },
    {
        files: ['eslint.config.mjs'],
        languageOptions: {
            parserOptions: {
                project: 'tsconfig.eslint.json',
            },
        },
    }
);
