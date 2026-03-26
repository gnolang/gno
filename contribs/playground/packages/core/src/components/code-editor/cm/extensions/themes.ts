import { Compartment } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { vscodeDark, vscodeDarkStyle, vscodeLight } from '@uiw/codemirror-theme-vscode'

import type { ThemeName } from '../types'

/**
 * Highlight styles for inline snippets in CodeLens and completion docs.
 *
 * Always dark mode as popups are also dark.
 */
export const codeLensHighlightStyles = vscodeDarkStyle

export const highlightClasses = {
  line: 'line--highlighted',
  gutter: 'gutter--highlighted', // unused atm, reserved when gutter highlight will be implemented.
}

export const defaultThemeStyles = EditorView.theme({
  '&': {
    height: '100%',
    flex: '1 1 auto',
  },
  '& .cm-scroller': {
    height: '100% !important',
  },
  '&:not(.cm-focused) .cm-activeLine': {
    backgroundColor: 'transparent!important',
  },
  '&:not(.cm-focused) .cm-activeLineGutter': {
    backgroundColor: 'transparent!important',
  },
  [`& .${highlightClasses.line}`]: {
    background: 'var(--gno-highlighted-line-bg, rgba(255, 255, 0, 0.25))',
  },
})

export const themeCompartment = new Compartment()

const themeExtensionFromName = (themeName?: ThemeName) => (themeName === 'dark' ? vscodeDark : vscodeLight)

/**
 * Returns theme compartment instance with initial theme.
 */
export const newThemeCompartment = (themeName?: ThemeName) => themeCompartment.of(themeExtensionFromName(themeName))

/**
 * Returns a new state effect to update theme.
 */
export const updateThemeEffect = (newTheme?: ThemeName) =>
  themeCompartment.reconfigure(themeExtensionFromName(newTheme))
