import { createPluginTheme, languageServerPlugin, type LanguageServerPluginOptions } from '@gnostudio/codemirror-lsp'

import { Compartment } from '@codemirror/state'

import { codeLensHighlightStyles } from './themes'

const EMPTY_COMPARTMENT: [] = []

const pluginTheme = createPluginTheme({
  highlightStyles: codeLensHighlightStyles,
})

export const defaultPluginOptions: Partial<LanguageServerPluginOptions> = {
  languageId: 'gno',
  theme: pluginTheme,
}

export const lspCompartment = new Compartment()

export const newLspCompartment = (opts?: LanguageServerPluginOptions) =>
  lspCompartment.of(opts ? languageServerPlugin(opts) : EMPTY_COMPARTMENT)

export const updateLspExtensionEffect = (opts: LanguageServerPluginOptions) =>
  lspCompartment.reconfigure(languageServerPlugin(opts))

export const resetLspExtensionEffect = () => lspCompartment.reconfigure(EMPTY_COMPARTMENT)
