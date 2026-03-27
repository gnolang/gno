import { createRequire } from 'node:module'
import { minifyPlugin } from '@gnostudio/vite-plugins'

import mdx from '@mdx-js/rollup'
import react from '@vitejs/plugin-react-swc'
import { defineConfig, PluginOption } from 'vite'
import istanbul from 'vite-plugin-istanbul'
import { nodePolyfills } from 'vite-plugin-node-polyfills'
import pluginRewriteAll from 'vite-plugin-rewrite-all'
import { viteStaticCopy } from 'vite-plugin-static-copy'
import svgr from 'vite-plugin-svgr'
import tsConfigPaths from 'vite-tsconfig-paths'

const require = createRequire(import.meta.url)

export default defineConfig(() => ({
  server: {
    hmr: {
      timeout: 1000,
    },
  },
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: (id: string) => {
          if (id.includes('browserfs')) {
            return 'browserfs'
          }
        },
      },
    },
  },
  plugins: [
    pluginRewriteAll(),
    viteStaticCopy({
      targets: [
        {
          src: './node_modules/@gnoide/files/src/img/favicon/playground/*',
          dest: './',
        },
        {
          src: './node_modules/@gnoide/files/src/img/og/og-playground-2.png',
          dest: './',
        },
      ],
    }),
    mdx({
      providerImportSource: '@mdx-js/react',
    }),
    react(),
    tsConfigPaths(),
    // Use a custom include to allow processing SVG files from monorepo
    // packages outside the app root (e.g. @gnoide/files).
    svgr({ include: '**/*.svg?react' }),
    nodePolyfills({
      globals: {
        Buffer: true,
        process: true,
      },
    }),
    istanbul({
      include: 'src/*',
      exclude: ['node_modules'],
      extension: ['.js', '.jsx', '.ts', '.tsx'],
      requireEnv: true,
      forceBuildInstrument: process.env.VITE_COVERAGE === 'true',
    }),
    minifyPlugin(true),
  ] as PluginOption[],
  worker: {
    plugins: () => [
      nodePolyfills({
        globals: {
          Buffer: true,
          process: true,
        },
      }),
      minifyPlugin(true),
    ],
  },
  resolve: {
    // Ensure react is always resolved from the app's node_modules.
    // Without this, SVG files transformed by svgr in monorepo packages
    // (e.g. @gnoide/files) fail to resolve react during build.
    dedupe: ['react', 'react-dom'],
    alias: {
      'vite-plugin-node-polyfills/shims/buffer': require.resolve('vite-plugin-node-polyfills/shims/buffer'),
      'vite-plugin-node-polyfills/shims/global': require.resolve('vite-plugin-node-polyfills/shims/global'),
      'vite-plugin-node-polyfills/shims/process': require.resolve('vite-plugin-node-polyfills/shims/process'),
    },
  },
}))
