import { resolve } from 'path'
import { defineConfig } from 'vite'
import dts from 'vite-plugin-dts'

export default defineConfig({
  base: './', // This is required for the worker asset to be embedded
  plugins: [dts()],
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'GnoStudioCodemirrorLsp',
      formats: ['es', 'umd'],
      fileName: (format) => `index.${format}.js`,
    },
    rollupOptions: {
      external: [
        '@codemirror/autocomplete',
        '@codemirror/lint',
        '@codemirror/state',
        '@codemirror/view',
        '@open-rpc/client-js',
        'vscode-languageserver-protocol',
      ],
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
})
