import type { Options } from 'tsup'

export const defaultTSUPOptions = {
  entry: ['src/index.tsx', 'src/index.ts'],
  name: 'tsup',
  target: 'es2019',
  dts: true,
  clean: true,
  sourcemap: true,
  format: ['esm', 'cjs'],
  treeshake: 'smallest',
} as const satisfies Options
