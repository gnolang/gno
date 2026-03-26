import { resolve } from 'path'
import { defineConfig } from 'tsup'

export default defineConfig((options) => ({
  entry: ['src/**/*.ts(x)?', '!src/**/*.test.ts'],
  splitting: false,
  bundle: false,
  sourcemap: true,
  clean: true,
  dts: true,
  minify: !options.watch,
  format: ['esm'],
  alias: {
    '~': resolve('src'),
  },
  target: 'esnext',
}))
