import React, { useCallback, useEffect, useRef } from 'react'
import { PiX } from 'react-icons/pi'
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels'

import { ContextMenuAction, EvalMode, useFormatter, type EditorController } from '@gnostudio/core'
import { Terminal, useDebounce } from '@gnostudio/react'
import { CMCodeEditorReact } from '@gnostudio/react/src/components'
import { useWorker } from '@gnostudio/wasm'

import { observer } from 'mobx-react-lite'

import { FileTabs } from '@/components/file-tabs'
import { useStore } from '@/contexts'
import { css } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'

import { StatusBar } from '../status-bar'
import { UploadDropzone } from '../upload-dropzone'
import { LSPServerType, useLspClient } from './hooks/use-lsp-client'

/**
 * Debounce timer interval to restart REPL on file change.
 */
const REPL_RESTART_DEBOUNCE_INTERVAL = 3000

export const Workspace: React.FC = observer(() => {
  const store = useStore()
  const worker = useWorker()
  const lspConfig = useLspClient(LSPServerType.Gopls)

  const containerRef = useRef<HTMLDivElement>(null)
  const resizeTimeout = useRef<NodeJS.Timeout | undefined>()
  const [restartRepl] = useDebounce(REPL_RESTART_DEBOUNCE_INTERVAL, () => {
    store.startWorker(worker, EvalMode.Repl)
  })

  const handleOnSave = (value: string) => {
    store.workbench.updateFileContent(store.workbench.activePath, value)
    if (store.evalState.isRepl) {
      restartRepl()
    }
  }

  const handleFormat = useFormatter()
  const [debouncedSave, cancelAutosave] = useDebounce(100, (changes: string) => {
    handleOnSave(changes)
  })
  const onManualSave = (changes: string) => {
    cancelAutosave()
    handleOnSave(changes)
    store.projects.save()
  }

  const handleEditorAction = (action: ContextMenuAction) => {
    switch (action) {
      case ContextMenuAction.RunLastAction:
        store.startWorker(worker, EvalMode.Run)
        break
      case ContextMenuAction.OpenRunPrompt:
        store.evalState.setRunPromptOpen(true)
        break
    }
  }

  const ENABLE_FUNCTION_GUTTER = store.settings.enableFunctionGutter

  const autoResizeEditor = useCallback(() => {
    clearTimeout(resizeTimeout.current)

    resizeTimeout.current = setTimeout(() => {
      if (!containerRef.current) return
      const height = containerRef.current.clientHeight

      store.editor.changeLayout({ width: containerRef.current.clientWidth, height })
      store.terminal.fit()
    }, 100)
  }, [store])

  const handleResizePanel = () => {
    // Immediately resize the terminal to fit the new panel size
    store.terminal.fit()
    autoResizeEditor()
  }

  const onMount = (editor: EditorController) => {
    editor.focus()
    store.editor.setEditor(editor)
  }

  const handleOnCloseTerminal = () => {
    store.stopWorker()
  }

  useEffect(
    function handleWindowResize() {
      window.addEventListener('resize', autoResizeEditor)
      return () => window.removeEventListener('resize', autoResizeEditor)
    },
    [autoResizeEditor],
  )

  return (
    <UploadDropzone>
      <section
        data-testid="workspace"
        className={stack({
          w: 'full',
          h: 'full',
          gap: '0',
        })}
      >
        <FileTabs disabled={store.editor.isLoading} />

        <PanelGroup direction="vertical" onLayout={handleResizePanel}>
          <Panel id="editor" order={1}>
            <div ref={containerRef} className={css({ height: 'full' })}>
              <CMCodeEditorReact
                key={store.settings.enableFunctionGutter ? 'editor-fn-gutter-on' : 'editor-fn-gutter-off'}
                workspaceId={store.workbench.workspaceId}
                theme={store.settings.isDark ? 'dark' : 'light'}
                value={store.workbench.activeFile}
                inputMode={store.editor.keybindingsStore.currentInputMode}
                lsp={lspConfig}
                onMount={onMount}
                onContextMenuAction={handleEditorAction}
                formatter={handleFormat}
                onChange={debouncedSave}
                onSave={onManualSave}
                onRunFunction={
                  ENABLE_FUNCTION_GUTTER
                    ? ({ func, values }) => {
                        const argsExpr = values
                          .map((v, index) => {
                            const type = func.params[index].type
                            return type === 'string' ? `"${v}"` : v
                          })
                          .join(', ')
                        const expr = `${func.name}(${argsExpr})`
                        store.evalState.setEvalExpression(expr)
                        store.startWorker(worker, EvalMode.Run)
                      }
                    : undefined
                }
                onEvent={(event) => store.editor.handleEditorEvent(event)}
              />
            </div>
          </Panel>

          {store.terminal.isOpen && (
            <>
              <PanelResizeHandle
                className={css({
                  h: '2px',
                  bg: 'border',
                  _active: { outline: 'solid 1px', outlineColor: 'border.highlight' },
                  zIndex: 1,
                })}
              />

              <Panel minSize={10} id="terminal" order={2}>
                <Terminal
                  store={store.terminal}
                  onClose={handleOnCloseTerminal}
                  className={css({
                    position: 'relative',
                    h: 'full',
                    '& [data-part="console"]': {
                      h: 'full',
                    },
                    '& .xterm': {
                      h: 'full',
                      p: '4',
                      pb: '20',
                    },
                  })}
                >
                  <Terminal.CloseButton asChild>
                    <button
                      className={css({ bg: 'black', position: 'absolute', right: 8, top: 4, color: 'white', p: 1 })}
                    >
                      <PiX size={20} />
                    </button>
                  </Terminal.CloseButton>
                </Terminal>
              </Panel>
            </>
          )}
        </PanelGroup>

        <StatusBar />
      </section>
    </UploadDropzone>
  )
})
