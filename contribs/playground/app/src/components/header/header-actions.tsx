import { DEFAULT_RUN_EXPRESSION, EvalMode } from '@gnostudio/core'
import { useWorker } from '@gnostudio/wasm'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { divider, hstack } from '@/styled-system/patterns'
import { link } from '@/styled-system/recipes'

import { RunPopover } from '../run-popover'

const verticalDividerClass = divider({ orientation: 'vertical', h: '4' })
const horizontalStackClass = hstack({ gap: '4' })

export const HeaderActions: React.FC = observer(() => {
  const store = useStore()
  const worker = useWorker()

  const handleFormat = () => {
    store.editor?.formatDocument()
  }

  const handleTest = async () => {
    store.startWorker(worker, EvalMode.Test)
  }

  const handleRepl = async () => {
    store.startWorker(worker, EvalMode.Repl)
  }

  const handleRun = (expr?: string) => {
    expr = expr ?? DEFAULT_RUN_EXPRESSION

    // Popover might still remain open/focused if terminal is already open.
    store.evalState.setRunPromptOpen(false)
    store.evalState.setEvalExpression(expr)
    store.startWorker(worker, EvalMode.Run)
  }

  return (
    <>
      <hr className={verticalDividerClass} />

      <ul className={horizontalStackClass}>
        <li>
          <button
            data-testid="btn-fmt"
            className={cx(css({ paw: 'Click+Header+Format' }), link())}
            onClick={handleFormat}
          >
            Format
          </button>
        </li>
      </ul>

      <hr className={verticalDividerClass} />

      <ul className={horizontalStackClass}>
        <li className={hstack({ gap: '1' })}>
          <button className={cx(css({ paw: 'Click+Header+Run' }), link())} onClick={() => handleRun()}>
            Run
          </button>

          <RunPopover
            onRunClick={handleRun}
            isOpen={store.evalState.isRunPromptOpen}
            onVisibilityChange={store.evalState.setRunPromptOpen}
            initialValue={store.evalState.evalExpression}
          />
        </li>
        <li>
          <button data-testid="btn-test" className={cx(css({ paw: 'Click+Header+Test' }), link())} onClick={handleTest}>
            Test
          </button>
        </li>
        <li>
          <button data-testid="btn-repl" className={cx(css({ paw: 'Click+Header+REPL' }), link())} onClick={handleRepl}>
            REPL
          </button>
        </li>
      </ul>
    </>
  )
})
