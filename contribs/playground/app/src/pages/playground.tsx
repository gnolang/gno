import { useEffect, useMemo, useRef } from 'react'
import { Helmet } from 'react-helmet'
import { useLocation, useParams } from 'react-router-dom'

import { EvalMode } from '@gnostudio/core'
import { useWorker } from '@gnostudio/wasm'

import { observer } from 'mobx-react-lite'

import { Workspace } from '@/components/workspace'
import { useStore } from '@/contexts'
import { useLoadWorkspace } from '@/hooks/use-cloud-mutation'

const useExprQuery = () => {
  const { search } = useLocation()

  return useMemo(() => {
    const qp = new URLSearchParams(search)
    return qp.get('run.expr')
  }, [search])
}

export const Playground: React.FC = observer(() => {
  const expr = useExprQuery()
  const store = useStore()
  const worker = useWorker()
  const params = useParams<{ id?: string }>()
  const { mutate: loadWorkspaceFiles, status } = useLoadWorkspace()
  const onceRef = useRef(false)

  useEffect(() => {
    loadWorkspaceFiles()
  }, [loadWorkspaceFiles])

  useEffect(() => {
    // Component is rendered twice in dev mode when React strict mode is enabled.
    if (status !== 'success' || onceRef?.current) {
      return
    }

    // Run expression when snippet is loaded
    onceRef.current = true
    if (expr?.length) {
      store.evalState.setEvalExpression(expr)
      store.startWorker(worker, EvalMode.Run)
    }
  }, [expr, status, onceRef, store, worker])

  if (status === 'pending') {
    return <p>Loading</p>
  }

  if (status === 'error') {
    return (
      <p>
        Failed to load shared files for ID &quot;<code>{params.id}</code>&quot;
      </p>
    )
  }

  return (
    <>
      <Helmet>
        <title>Gno Playground - Online editor for exploring gno.land</title>
        <meta
          name="description"
          content="Explore Gno Playground for smart contract development on gno.land: Create, test, deploy, and share contracts easily in a user-friendly interface."
        />
      </Helmet>
      <Workspace />
    </>
  )
})
