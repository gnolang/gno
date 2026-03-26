import { type LanguageServerClient } from '@gnostudio/codemirror-lsp'

import type { EditorState, StateEffect } from '@codemirror/state'

import { clearHighlightsEffect } from '../extensions/highlight'
import { updateInputModeEffect } from '../extensions/input'
import { defaultPluginOptions, resetLspExtensionEffect, updateLspExtensionEffect } from '../extensions/lsp'
import { readOnlyEffect } from '../extensions/readonly'
import { Syntax, syntaxFromFileName, updateSyntaxEffect } from '../extensions/syntax'
import { updateThemeEffect } from '../extensions/themes'
import type { EditorProps } from '../types'
import { stateDataField, updateStateDataEffect } from './field'
import { defaultStateData, type StateData } from './types'

export type StateProps = Pick<EditorProps, 'theme' | 'inputMode' | 'readonly' | 'value'>

/**
 * Creates new uninitialized StateData from React component props.
 */
export const stateDataFromProps = (
  { theme, inputMode, readonly: readOnly = false, value }: StateProps,
  defaults = defaultStateData,
): StateData => ({
  isInitialised: false,
  theme: theme ?? defaults.theme,
  inputMode: inputMode ?? defaults.inputMode,
  fileName: value?.path?.toString(), // Force copy to avoid use-after-free of mobx state nodes.
  syntax: syntaxFromFileName(value?.path),
  readOnly,
})

/**
 * Creates a new extension instance to use StateData in editor with defaults from React props.
 * Any custom state field is essentially an extension.
 *
 * Props function invoked any time when trying to get StateData on an empty EditorState.
 * Check `isInitialized` field to check if `StateData` is not defaults.
 *
 * @param propsFn Funciton to obtain up-to-date React props.
 */
export const newStateDataFieldExtension = (propsFn: () => StateProps) => {
  return stateDataField.init(() => stateDataFromProps(propsFn()))
}

/**
 * Returns effect that sets StateData field in EditorState.
 *
 * Used to write field at a new, fresh empty EditorState.
 * Necessary as if StateData field is undefined, `state.field()` will return
 * every time a new copy of defaults based on current React props.
 *
 * This behavior breaks change detection and can be fixed by writing
 * StateData field manually to each new state.
 *
 * @param props
 */
export const setStateDataEffect = (props: EditorProps) => {
  const stateData = stateDataFromProps(props)
  stateData.isInitialised = true
  return updateStateDataEffect.of(stateData)
}

/**
 * Obtains StateData from passed editor state.
 *
 * Just returns defaults built from current props if state doesn't yet have it's own StateData.
 * Check `isInitialized` field to check if `StateData` is not defaults.
 *
 * @see `newStateDataFieldExtension`
 */
export const getStateData = (state: EditorState) => {
  return state.field(stateDataField, false)!
}

/**
 * Checks whether editor state is not empty by checking if it has initialized StateData field.
 */
export const hasStateData = (state: EditorState) => !!getStateData(state).isInitialised

type StateEffects = Array<StateEffect<any>>

interface ChangeSet {
  effects: StateEffects
  isChanged: boolean
  changes: Partial<StateData>
}

/**
 * Data necessary to operate with state extensions.
 */
interface ExtensionContext {
  lspClient?: LanguageServerClient
}

const gnoLanguageId = 'gno'

/**
 * Compares a StateData object obtained from EditorState with component props.
 *
 * Returns a list of effects to be applied on editor based updated component props.
 * Used to update editor state according to values in props.
 *
 * Does most of heavy lifting job on change detection.
 *
 * @param props React component props.
 * @param stateData StateData to compare.
 */
export const checkStateDataChanges = (ctx: ExtensionContext, props: EditorProps, stateData: StateData): ChangeSet => {
  const { value, theme = 'light', inputMode = 'classic', readonly = false } = props
  const effects: StateEffects = []
  const changes: Partial<StateData> = {
    isInitialised: true,
  }

  // This flag is false if state was just created and doesn't have stateData assigned yet.
  // This happens when EditorState was just created for a new file or `getStateData` called on empty state.
  //
  // Just initialise StateData by setting it to EditorState and initialize all compartments.
  // This is necessary because some compartments like themes needs to be initialized for every state.
  const { isInitialised, isLoading } = stateData

  const fileSyntax = syntaxFromFileName(value?.path)

  if (!isInitialised || stateData.fileName !== value?.path) {
    // Clear highlights on tab switch
    effects.push(clearHighlightsEffect())
    changes.fileName = value?.path

    // Enable per-document LSP extension if client is configured.
    // LSP server doesn't support modfiles and thus has to be disabled for modfile document.
    const isSourceFile = fileSyntax === Syntax.Gno
    if (value && props.lsp && ctx.lspClient && isSourceFile && !isLoading) {
      effects.push(
        updateLspExtensionEffect({
          ...defaultPluginOptions,
          languageId: gnoLanguageId,
          documentUri: `${props.lsp.rootUri}/${value.path}`,
          client: ctx.lspClient,
        }),
      )
    } else {
      effects.push(resetLspExtensionEffect())
    }
  }

  if (!isInitialised || stateData.theme !== theme) {
    effects.push(updateThemeEffect(theme))
    changes.theme = theme
  }

  if (!isInitialised || stateData.syntax !== fileSyntax) {
    effects.push(updateSyntaxEffect(fileSyntax))
    changes.syntax = fileSyntax
  }

  if (!isInitialised || stateData.readOnly !== readonly) {
    effects.push(readOnlyEffect(readonly))
    changes.readOnly = readonly
  } else if (isLoading) {
    // View should be locked during loading
    effects.push(readOnlyEffect(true))
  }

  if (!isInitialised || stateData.inputMode !== inputMode) {
    const effect = updateInputModeEffect(inputMode)
    changes.inputMode = inputMode
    effects.push(effect)
  }

  const isChanged = Object.keys(changes).length > 0
  if (isChanged) {
    effects.unshift(updateStateDataEffect.of(changes))
  }

  return { isChanged, effects, changes }
}

/**
 * Returns a new StateData based on a current one in EditorState but for a different document.
 */
export const replaceStateFileName = (state?: EditorState, fileName?: string): StateData => {
  const stateData = state ? getStateData(state) : defaultStateData

  // Don't update syntax to keep change detection work.
  return {
    ...stateData,
    isInitialised: false,
    fileName,
  }
}

/**
 * Returns effects to mark document state as ready and start postponed extensions like LSP client.
 *
 * Intended to be executed after async document contents were loaded.
 */
export const markDocumentDownloadedEffect = (ctx: ExtensionContext, props: EditorProps): StateEffects => {
  const { readonly, value } = props

  const effects: StateEffects = [
    updateStateDataEffect.of({
      isLoading: false,
      readOnly: readonly,
    }),
    readOnlyEffect(!!readonly),
  ]

  // Resume suspended LSP plugin.
  const fileSyntax = syntaxFromFileName(value?.path)
  effects.push(updateSyntaxEffect(fileSyntax))

  const isSourceFile = fileSyntax === Syntax.Gno
  if (value && props.lsp && ctx.lspClient && isSourceFile) {
    effects.push(
      updateLspExtensionEffect({
        ...defaultPluginOptions,
        languageId: gnoLanguageId,
        documentUri: `${props.lsp.rootUri}/${value.path}`,
        client: ctx.lspClient,
      }),
    )
  }

  return effects
}

/**
 * Returns effect that enables LSP plugin.
 */
export const startLspPluginEffect = (
  state: EditorState | null | undefined,
  client: LanguageServerClient | null | undefined,
) => {
  if (!client || !state) {
    return null
  }

  const { fileName, isInitialised } = getStateData(state)
  if (!fileName || !isInitialised) {
    return null
  }

  return updateLspExtensionEffect({
    ...defaultPluginOptions,
    languageId: gnoLanguageId,
    documentUri: `${client.rootUri}/${fileName}`,
    client,
  })
}
