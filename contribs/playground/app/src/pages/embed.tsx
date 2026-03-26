import React, { useEffect } from 'react'
import { PiArrowUpRight, PiPlayFill, PiX } from 'react-icons/pi'
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels'
import { useParams, useSearchParams } from 'react-router-dom'

import { EvalMode } from '@gnostudio/core'
import { CMCodeEditorReact, Terminal } from '@gnostudio/react'
import { useWorker } from '@gnostudio/wasm'

import { useMutation } from '@tanstack/react-query'
import { observer } from 'mobx-react-lite'

import { FileTabs } from '@/components/file-tabs'
import { useStore } from '@/contexts'
import { useLoadWorkspace } from '@/hooks'
import { useGnoVMVersionQuery } from '@/hooks/use-gnovm-query'
import { css, cx } from '@/styled-system/css'
import { hstack, stack } from '@/styled-system/patterns'
import { button, link } from '@/styled-system/recipes'

export const Embed: React.FC = observer(() => {
  const worker = useWorker()
  const store = useStore()

  const params = useParams<{ id?: string }>()
  const [searchParams] = useSearchParams()

  const runExpr = searchParams.get('run.expr')
  const theme = searchParams.get('theme')

  const { mutate: loadWorkspaceFiles, status: loadStatus } = useLoadWorkspace()
  const { data } = useGnoVMVersionQuery()

  const { mutate: run, status: runStatus } = useMutation({
    mutationKey: ['gnorun', params.id],
    mutationFn: async () => {
      store.evalState.setEvalExpression(runExpr as string)
      store.startWorker(worker, EvalMode.Run)
    },
  })

  const buildOpenInPlaygroundUrl = () => {
    if (params.id) {
      return new URL(`/p/${params.id}`, window.location.origin).toString()
    } else {
      const url = new URL('/p', window.location.origin)
      url.hash = window.location.hash
      return url.toString()
    }
  }

  useEffect(() => {
    loadWorkspaceFiles()
  }, [loadWorkspaceFiles])

  useEffect(
    function initialRun() {
      if (runExpr && runStatus === 'idle') {
        run()
      }
    },
    [loadStatus, runExpr, runStatus, run],
  )

  useEffect(
    function updateSettings() {
      if (theme) {
        document.documentElement.setAttribute('data-theme', theme)
      }
    },
    [theme],
  )

  return (
    <div
      className={stack({
        gap: '0',
        w: 'full',
        h: 'full',
        borderWidth: '1px',
        overflow: 'hidden',
      })}
    >
      <FileTabs
        showActions={false}
        rightElement={
          <div className={hstack({ gap: '2', mr: '4' })}>
            <button onClick={() => run()} className={cx(button({ variant: 'ghost' }), css({ gap: '1' }))}>
              <PiPlayFill size={14} /> Run
            </button>
          </div>
        }
      />

      <hr />

      <PanelGroup direction="vertical" onLayout={() => {}}>
        <Panel id="editor" order={1}>
          <div className={css({ h: 'full' })}>
            <CMCodeEditorReact
              workspaceId={store.workbench.workspaceId}
              theme={(theme as any) ?? 'light'}
              value={store.workbench.activeFile}
              inputMode={store.editor.keybindingsStore.currentInputMode}
              onMount={store.editor.setEditor}
              onChange={(value) => {
                store.workbench.updateFileContent(store.workbench.activePath, value)
              }}
            />
          </div>
        </Panel>

        {store.terminal.isOpen && (
          <>
            <PanelResizeHandle
              className={css({
                h: '1px',
                bg: 'border',
                _hover: { outline: 'solid 1px', outlineColor: 'border.highlight' },
                _active: { outline: 'solid 1px', outlineColor: 'border.highlight' },
                zIndex: 1,
              })}
            />

            <Panel
              id="terminal"
              order={2}
              className={css({ display: 'flex', flexDir: 'column' })}
              onResize={() => {
                store.terminal.fit()
              }}
            >
              <Terminal
                store={store.terminal}
                className={css({
                  position: 'relative',
                  h: 'full',
                  '& [data-part="console"]': {
                    h: 'full',
                  },
                  '& .xterm': {
                    h: 'full',
                    p: '4',
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

      <div
        className={hstack({
          mt: '0',
          borderTopWidth: '1px',
          fontSize: 'sm',
          justify: 'space-between',
          px: '2',
          py: '1',
        })}
      >
        <a
          href={buildOpenInPlaygroundUrl()}
          target="_blank"
          rel="noreferrer noopener"
          className={cx(link(), hstack({ gap: '1' }))}
        >
          Open in Playground <PiArrowUpRight size={16} />
        </a>

        <div className={hstack({ gap: '1' })}>
          <span>GnoVM:</span>
          <span>{data}</span>
        </div>
      </div>
    </div>
  )
})
