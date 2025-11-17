// @ts-check

import eslint from '@eslint/js';
import { defineConfig } from 'eslint/config';
import ts from 'typescript-eslint';
import mocha from 'eslint-plugin-mocha';

export default defineConfig(
    eslint.configs.recommended,
    ts.configs.strictTypeChecked,
    {
        files: ['*.ts'],
        languageOptions: {
            parserOptions: {
                projectService: true,
                tsconfigRootDir: import.meta.dirname,
            },
        },
    },
    {
        files: ['*.ts', '*.mjs'],
        rules: {
            eqeqeq: ['error', 'always'],
            '@typescript-eslint/no-unnecessary-condition': ['error', { allowConstantLoopConditions: true }],
            '@typescript-eslint/no-unsafe-member-access': 'off',
            '@typescript-eslint/no-unsafe-argument': 'off',
            '@typescript-eslint/no-unsafe-assignment': 'off',
            '@typescript-eslint/restrict-template-expressions': 'off',
        },
    },
    {
        files: ['test.ts'],
        // The cast is workaround for https://github.com/lo1tuma/eslint-plugin-mocha/issues/392
        .../** @type {{recommended: import('eslint').Linter.Config}} */ (mocha.configs).recommended,
    },
    {
        files: ['test.ts'],
        rules: {
            '@typescript-eslint/unbound-method': 'off', // For checking `window.runActionlint`
            'mocha/no-exclusive-tests': 'error',
            'mocha/no-pending-tests': 'error',
            'mocha/no-top-level-hooks': 'error',
            'mocha/consistent-interface': ['error', { interface: 'BDD' }],
        },
    },
    {
        files: ['eslint.config.mjs'],
        languageOptions: {
            parserOptions: {
                projectService: false,
                project: 'tsconfig.eslint.json',
            },
        },
    }
);
