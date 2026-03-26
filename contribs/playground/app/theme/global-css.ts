import { defineGlobalStyles } from '@pandacss/dev'

export const globalCss = defineGlobalStyles({
  '*, *::before, *::after': {
    borderColor: 'border',
  },
  'html, body, #root': {
    height: '100%',
  },
  ':root': {
    fontSize: '0.875em',
  },
  html: {
    fontFamily: '"Inter Variable", system-ui, sans-serif',
    fontWeight: '500',
    textRendering: 'optimizeLegibility',
    fontFeatureSettings: `'kern', 'liga', 'clig', 'calt'`,
    WebkitFontSmoothing: 'antialiased',
    MozOsxFontSmoothing: 'grayscale',
  },
  body: {
    bg: 'background',
    color: 'foreground',
  },
})
