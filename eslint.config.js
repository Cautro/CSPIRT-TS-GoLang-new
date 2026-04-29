import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import security from 'eslint-plugin-security'
import noUnsanitized from 'eslint-plugin-no-unsanitized'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    plugins: {
      security,
      'no-unsanitized': noUnsanitized,
    },
    languageOptions: {
      globals: globals.browser,
    },
    rules: {
      'security/detect-eval-with-expression': 'error',
      'security/detect-new-buffer': 'error',
      'security/detect-non-literal-regexp': 'warn',
      'no-unsanitized/method': 'error',
      'no-unsanitized/property': 'error',
      'no-console': import.meta?.env?.PROD ? 'error' : 'warn',
    },
  },
])