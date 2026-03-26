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

export default defineConfig(() => ({
  server: {
    hmr: {
      timeout: 1000,
    },
  },
  build: {
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
    svgr(),
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
    minifyPlugin(process.env.SKIP_OBFUSCATION === 'true'),
  ] as PluginOption[],
  worker: {
    plugins: () => [
      nodePolyfills({
        globals: {
          Buffer: true,
          process: true,
        },
      }),
      minifyPlugin(process.env.SKIP_OBFUSCATION === 'true'),
    ],
  },
}))
