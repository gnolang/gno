import { types } from 'mobx-state-tree'
import { persist } from 'mst-persist'

const colorSchemeQuery = window.matchMedia('(prefers-color-scheme: dark)')
export const Settings = types
  .model({
    theme: types.optional(types.enumeration(['light', 'dark']), colorSchemeQuery.matches ? 'dark' : 'light'),
    fontSize: types.optional(types.number, 16),
    editorMode: types.optional(types.enumeration(['vim', 'emacs', 'classic']), 'classic'),
    enableFunctionGutter: types.optional(types.boolean, false),
  })
  .views((self) => ({
    get isDark() {
      return self.theme === 'dark'
    },
  }))
  .actions((self) => ({
    setEditorMode(adapter: 'vim' | 'emacs' | 'classic') {
      self.editorMode = adapter
    },

    setTheme(theme: 'light' | 'dark') {
      self.theme = theme
    },

    setFontSize(size: number) {
      self.fontSize = size
    },

    setFunctionGutterEnabled(value: boolean) {
      self.enableFunctionGutter = value
    },

    afterCreate() {
      persist('settings', self).catch(console.error)
    },
  }))
  .actions((self) => ({
    afterCreate() {
      colorSchemeQuery.onchange = (e) => self.setTheme(e.matches ? 'dark' : 'light')
    },
  }))
