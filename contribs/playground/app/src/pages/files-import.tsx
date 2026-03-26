import React from 'react'
import { PiArrowsClockwise, PiCircle, PiGitBranch } from 'react-icons/pi'
import { useLocation, useParams, useSearchParams } from 'react-router-dom'

import { EvalMode } from '@gnostudio/core'
import { useWorker } from '@gnostudio/wasm'

import { useQuery } from '@tanstack/react-query'

import { Workspace } from '@/components/workspace'
import { useStore } from '@/contexts'
import { css } from '@/styled-system/css'
import { hstack, stack } from '@/styled-system/patterns'

interface ImportParams {
  strategy: 'github'
}

const useExprQuery = () => {
  const { search } = useLocation()

  return React.useMemo(() => {
    const qp = new URLSearchParams(search)
    return qp.get('run.expr')
  }, [search])
}

export const FilesImport: React.FC<ImportParams> = ({ strategy }) => {
  const expr = useExprQuery()
  const store = useStore()
  const worker = useWorker()
  const params = useParams<{ owner: string; repo: string }>()
  const [searchParams] = useSearchParams()
  const onceRef = React.useRef(false)

  const { status } = useQuery({
    queryKey: ['FILES_IMPORT'],
    queryFn: () => {
      const file = searchParams.get('file')
      const branch = searchParams.get('branch') || 'main'
      const { owner, repo } = params

      if (!owner || !repo) {
        throw new Error('Owner and repo are required')
      }

      if (file) {
        return store.workbench.loadSingleFileFromGithub(owner, repo, file, branch)
      } else {
        return store.workbench.loadFromGit(strategy, owner, repo)
      }
    },
  })

  React.useEffect(() => {
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

  if (status === 'error') {
    return <p>Failed to load files from {strategy}</p>
  }

  if (status === 'success') {
    return <Workspace />
  }

  return (
    <div className={stack({ gap: '8', w: 'full', h: 'full', alignItems: 'center', justifyContent: 'center' })}>
      <div>
        <h1
          className={css({
            textTransform: 'uppercase',
            fontWeight: 'medium',
          })}
        >
          Importing from {strategy}
        </h1>

        <p className={hstack({ gap: 2 })}>
          <PiGitBranch size={24} />
          <span className={css({ fontFamily: 'mono' })}>
            {params.owner}/{params.repo}
          </span>
        </p>
      </div>

      <ol className={stack({ gap: 2 })}>
        <li className={hstack({ gap: 2 })}>
          <span className={css({ animation: 'rotate 1s infinite' })}>
            <PiArrowsClockwise size={18} />
          </span>
          <span>Cloning repo from {strategy}</span>
        </li>
        <li className={hstack({ gap: 2 })}>
          <PiCircle size={18} />
          <span>Reading files from repo</span>
        </li>
        <li className={hstack({ gap: 2 })}>
          <PiCircle size={18} />
          <span>Importing files to the playground</span>
        </li>
      </ol>
    </div>
  )
}
