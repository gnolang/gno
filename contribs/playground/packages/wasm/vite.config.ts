import { createRequire } from 'node:module'
import { resolve } from 'path'
import { defineConfig, type Plugin } from 'vite'
import dts from 'vite-plugin-dts'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

const require = createRequire(import.meta.url)

export default defineConfig({
  base: './', // This is required for the worker asset to be embedded
  plugins: [dts(), nodePolyfills({})] as Plugin[],
  worker: {
    // Make Node pollyfills available to workers
    plugins: () => [nodePolyfills({})],
  },
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
      // Resolve polyfill shims from this package's node_modules so the worker
      // build can find them even when the importing file lives in a different
      // workspace package (pnpm strict resolution would otherwise fail).
      'vite-plugin-node-polyfills/shims/buffer': require.resolve('vite-plugin-node-polyfills/shims/buffer'),
      'vite-plugin-node-polyfills/shims/global': require.resolve('vite-plugin-node-polyfills/shims/global'),
      'vite-plugin-node-polyfills/shims/process': require.resolve('vite-plugin-node-polyfills/shims/process'),
    },
  },
})
