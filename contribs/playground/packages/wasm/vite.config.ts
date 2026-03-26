import { resolve } from 'path'
import { defineConfig, type Plugin } from 'vite'
import dts from 'vite-plugin-dts'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig({
  base: './', // This is required for the worker asset to be embedded
  plugins: [dts(), nodePolyfills({})] as Plugin[],
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'GnoStudioWasm',
      formats: ['es', 'umd'],
      fileName: (format) => `index.${format}.js`,
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
})
