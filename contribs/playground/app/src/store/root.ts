import { EvalMode, EvalState, keybindingsStore, TerminalStore, WalletStore, type InputMode } from '@gnostudio/core'
import type { Worker } from '@gnostudio/wasm'

import { dirname } from 'path'
import { reaction } from 'mobx'
import { addDisposer, applySnapshot, flow, getSnapshot, types, type Instance } from 'mobx-state-tree'
import qs from 'qs'

import { Chains } from './chains'
import { Deployer } from './deployer'
import { Editor } from './editor'
import { Projects } from './projects'
import { Settings } from './settings'
import { MessageType } from './types/status-message'
import { Workbench } from './workbench'

const defaultFileName = 'package.gno'
const defaultCodeSnippet = `package hello

func Render(path string) string {
  return "Hello World!"
}
`

interface StoreQueryParams {
  file?: string
  highlight?: string
}

export const rootStore = types
  .model({
    deployer: Deployer,
    projects: Projects,
    settings: Settings,
    workbench: Workbench,
    wallet: WalletStore,
    editor: Editor,
    evalState: EvalState,
    chains: Chains,
  })
  .volatile(() => ({
    terminal: new TerminalStore(),
    queryParams: {} as StoreQueryParams,
    worker: null as Worker | null,
  }))
  .props({
    deployer: types.optional(Deployer, {}),
    projects: types.optional(Projects, {}),
    settings: types.optional(Settings, {}),
    wallet: types.optional(WalletStore, {}),
    workbench: types.optional(Workbench, {
      activePath: defaultFileName,
      files: {
        [defaultFileName]: {
          content: defaultCodeSnippet,
          path: defaultFileName,
        },
      },
    }),
    editor: types.optional(Editor, {}),
    evalState: types.optional(EvalState, {}),
    chains: types.optional(Chains, {}),
  })
  .actions((self) => ({
    loadQueryParams() {
      const storeQueryParams = qs.parse(window.location.search, { ignoreQueryPrefix: true }) as StoreQueryParams
      self.queryParams = storeQueryParams
    },

    updateQueryParams(params: StoreQueryParams) {
      const result = Object.assign({}, self.queryParams)

      for (const key in params) {
        const k = key as keyof StoreQueryParams
        result[k] = params[k]
      }

      self.queryParams = result

      // Update the URL with the new query params
      const search = qs.stringify(result, { addQueryPrefix: true })
      const hash = window.location.hash
      window.history.replaceState({}, '', `${search}${hash}`)
    },

    // Run this after the store is hydrated
    boot() {
      if (self.queryParams.file) {
        self.workbench.setActivePath(self.queryParams.file)
      }

      if (self.queryParams.highlight) {
        self.editor.highlightLine(parseInt(self.queryParams.highlight))
      }
    },

    resetWorkbench() {
      if (self.projects.hasUnsavedChanges) {
        const confirmed = window.confirm('You have unsaved changes. Are you sure you want to proceed?')
        if (!confirmed) return false
      }

      applySnapshot(self.workbench, {
        activePath: defaultFileName,
        files: {
          [defaultFileName]: {
            content: defaultCodeSnippet,
            path: defaultFileName,
          },
        },
      })

      return true
    },
  }))
  .actions((self) => ({
    startWorker: flow(function* (worker: Worker, evalMode: EvalMode) {
      const isRepl = evalMode === EvalMode.Repl
      const { port1: clientPort, port2: stdioPort } = new MessageChannel()
      yield self.terminal.attachToMessagePort(clientPort, !isRepl)

      const t = self.terminal

      t.open()
      t.clear()

      // #300: Notify user about source code update when in REPL mode
      const isReplRefreshed = isRepl && self.evalState.isRepl
      if (isReplRefreshed) {
        t.write('// Source code was changed, REPL reloaded\n')
      }

      // Intentionally do not show a loading spinner in the status bar during REPL (re)start
      // to avoid confusion after the action has finished. We'll only show a brief
      // informational message once the REPL is ready.

      self.evalState.setEvalMode(evalMode)

      // disable spinner for run mode
      if (evalMode !== EvalMode.Run) {
        t.startSpinner()
      }

      const workingDir = dirname(self.workbench.activePath)
      const files = getSnapshot(self.workbench.files)

      try {
        switch (evalMode) {
          case EvalMode.Repl:
            yield worker.gnorepl({ files, stdioPort, workingDir })
            break
          case EvalMode.Test:
            yield worker.gnotest({ files, stdioPort, workingDir })
            break
          case EvalMode.Run: {
            const { evalExpression: expr } = self.evalState
            if (!expr) {
              // Show expression prompt for first time
              self.evalState.setRunPromptOpen(true)
              return
            }
            yield worker.gnorun({ files, expr, stdioPort, workingDir })
            break
          }
        }
        self.worker = worker

        if (isRepl) {
          const text = isReplRefreshed ? 'REPL reloaded' : 'REPL started'
          self.workbench.setStatusMessage({ type: MessageType.Info, tag: 'REPL', text })
          yield new Promise((resolve) => setTimeout(resolve, 2000))
          self.workbench.setStatusMessage(null)
        }
      } catch (err) {
        // Notify user about WASM startup error
        self.terminal.printFatalError(err as Error)
        if (isRepl) {
          self.workbench.setStatusMessage({ type: MessageType.Error, tag: 'REPL', text: 'Failed to (re)load REPL' })
        }
      }
    }),

    stopWorker() {
      self.evalState.setEvalMode(EvalMode.None)
      self.worker?.terminate()
      self.worker = null
    },

    afterCreate() {
      self.loadQueryParams()

      // Update query params when the active path or highlighted line changes
      const qsDisposer = reaction(
        () => [self.workbench.activePath, self.editor.highlightedLine] as const,
        ([path, line]) => {
          self.updateQueryParams({ file: path, highlight: line?.toString() })
        },
        { delay: 100 },
      )

      // Update the editor keybindings when the settings change
      const keybindingsDisposer = reaction(
        () => self.settings.editorMode,
        () => {
          if (!self.editor.editorCtrl) return
          return self.editor.keybindingsStore.setInputMode(
            self.editor.editorCtrl,
            self.settings.editorMode as InputMode,
          )
        },
        { delay: 800, fireImmediately: true },
      )

      addDisposer(self, qsDisposer)
      addDisposer(self, keybindingsDisposer)
      addDisposer(
        self,
        reaction(() => keybindingsStore.currentInputMode, self.settings.setEditorMode),
      )

      // Sync account chain ID with selected chain.
      addDisposer(
        self,
        reaction(
          () => self.wallet.account,
          (account) => {
            if (!account) return
            self.chains.setActiveChainById(account.chain)
          },
        ),
      )
    },
  }))

export type RootStore = Instance<typeof rootStore>
