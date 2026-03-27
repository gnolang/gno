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
    rollupOptions: {
      // Externalize all dependencies so Rollup doesn't try to bundle them.
      // This is needed because workspace packages like @gnostudio/pkg expose
      // raw TypeScript source ("main": "src/index.ts") which pulls in CJS
      // dependencies (browserfs, path-browserify) whose named exports Rollup
      // can't auto-detect, causing "X is not exported" build errors.
      // Only local source files (relative, absolute, or aliased paths) are bundled.
      external: (id) => {
        if (id.startsWith('.') || id.startsWith('/') || id.startsWith('@/') || id.startsWith('~')) return false
        return true
      },
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
})
