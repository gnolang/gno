import { defineConfig } from '@pandacss/dev'

import { globalCss } from './theme/global-css'
import { keyframes } from './theme/keyframes'
import { recipes } from './theme/recipes'
import { semanticTokens } from './theme/semantic-tokens'
import { tokens } from './theme/tokens'

export default defineConfig({
  outdir: 'styled-system',
  preflight: true,
  separator: '=',

  include: ['./src/**/*.{ts,tsx}'],
  exclude: [],

  jsxFramework: 'react',
  conditions: {
    extend: {
      dark: '.dark &, [data-theme="dark"] &',
      light: '.light &',
    },
  },

  theme: {
    extend: {
      tokens,
      semanticTokens,
      keyframes,
    },
    recipes,
  },

  utilities: {
    extend: {
      pawpalName: {
        className: 'pawpal-event-name',
        shorthand: 'paw',
        transform() {
          return {}
        },
      },
    },
  },

  globalCss,
})
