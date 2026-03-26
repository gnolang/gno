import React, { useMemo, useState } from 'react'
import { PiCheck, PiCopySimpleFill } from 'react-icons/pi'

import { packagePaths } from '@gnostudio/core'
import { Popover, useWriteClipboard } from '@gnostudio/react'

import { useMutation } from '@tanstack/react-query'
import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { hstack, stack } from '@/styled-system/patterns'
import { button, link } from '@/styled-system/recipes'

import { PkgPathInput } from './pkg-path-input'
import { TxPreviewLink } from './tx-preview-link'

const titleMap = {
  idle: 'Deploy',
  pending: 'Deploy',
  success: 'Deployment successful',
  error: 'Deployment failed',
}

const buttonLabelMap = {
  idle: 'Deploy',
  pending: 'Deploying...',
  success: 'New deployment',
  error: 'New deployment',
}

export const DeployStep: React.FC = observer(() => {
  const store = useStore()
  const [previewUrl, setPreviewUrl] = useState<string | undefined>()

  const {
    mutate: deploy,
    status,
    reset,
    error,
  } = useMutation({
    mutationFn: () => store.deployer.deploy(store.wallet.account?.chain as string),
  })

  const [writeToClipboard, isCopied] = useWriteClipboard()

  const isCTAButtonDisabled = useMemo(() => {
    if (status === 'pending') return true
    if (status === 'idle') return !store.deployer.hasValidPath
    return false
  }, [status, store.deployer.hasValidPath])

  const handleNewDeploy = () => {
    if (status === 'idle') {
      return deploy()
    }
    reset()
  }

  const handleTryAgain = () => {
    deploy()
  }

  const handleCopyLink = () => {
    const { txHash } = store.deployer
    if (!txHash) {
      return
    }

    writeToClipboard(previewUrl ?? txHash).catch(console.error)
  }

  const handleCopyError = () => {
    const message = error?.message ?? 'Deploy failed in silence'
    writeToClipboard(message).catch(console.error)
  }

  const handleDisconnect = () => {
    store.wallet.disconnect()
  }

  return (
    <>
      <Popover.Title className={css({ fontSize: 'lg', fontWeight: 'bold' })}>{titleMap[status]}</Popover.Title>
      <div className={stack({ gap: '4', mt: '4' })}>
        <div className={stack({ gap: '2' })}>
          <p className={css({ fontSize: 'md', color: 'foreground' })}>You are connected with:</p>
          <div
            className={hstack({
              gap: '2',
              alignItems: 'center',
              p: '3',
              bg: 'header',
              borderRadius: 'md',
              border: '1px solid',
              borderColor: 'border',
            })}
          >
            <span className={css({ fontFamily: 'mono', fontSize: 'sm', flex: '1', wordBreak: 'break-all' })}>
              {store.wallet.account?.address}
            </span>
            <button
              onClick={() => writeToClipboard(store.wallet.account?.address || '').catch(console.error)}
              className={css({ p: '1', borderRadius: 'sm', _hover: { bg: 'background' }, flexShrink: 0 })}
              title="Copy wallet address"
            >
              {isCopied ? <PiCheck size={14} /> : <PiCopySimpleFill size={14} />}
            </button>
          </div>
          <button className={cx(css({ paw: 'Click+Popover+Disconnect' }), link())} onClick={handleDisconnect}>
            Disconnect wallet
          </button>
        </div>

        {status === 'idle' && (
          <div className={stack()}>
            <label htmlFor="pkg-path" className={css({ fontWeight: 'semibold' })}>
              Path
            </label>

            <PkgPathInput onSubmit={handleNewDeploy} />
          </div>
        )}

        {status === 'success' && (
          <TxPreviewLink
            hash={store.deployer.txHash}
            chain={store.chains.selectedChain}
            packagePath={store.deployer.pkgPath}
            isRealm={packagePaths.isRealmPath(store.deployer.pkgPath)}
            onPreviewUrl={setPreviewUrl}
          />
        )}

        {status === 'error' && (
          <div className={stack()}>
            <div
              className={css({
                position: 'relative',
                border: '1px solid',
                color: 'foreground',
                fontFamily: 'monospace',
              })}
            >
              <div className={css({ display: 'flex', justifyContent: 'space-between', p: '2' })}>
                <p className={css({ fontWeight: 'semibold', color: 'red', position: 'relative' })}>Error</p>
                <button onClick={handleCopyError}>
                  {isCopied ? <PiCheck size={16} /> : <PiCopySimpleFill size={16} />}
                </button>
              </div>
              <pre className={css({ maxHeight: '240px', overflow: 'auto', px: '2', py: '1' })}>{error.message}</pre>
            </div>
          </div>
        )}

        <div className={hstack({ justifyContent: 'flex-end' })}>
          <button
            className={cx(css({ paw: 'Click+Popover+Deploy' }), button())}
            onClick={handleNewDeploy}
            disabled={isCTAButtonDisabled}
          >
            {buttonLabelMap[status]}
          </button>

          {status === 'success' && (
            <>
              <button className={cx(css({ paw: 'Click+Popover+Deploy+TxUrlCopy' }), button())} onClick={handleCopyLink}>
                {isCopied ? 'Copied!' : `Copy ${previewUrl ? 'link' : 'hash'}`}
              </button>
            </>
          )}

          {status === 'error' && (
            <button className={cx(css({ paw: 'Click+Popover+Deploy+TryAgain' }), button())} onClick={handleTryAgain}>
              Try again
            </button>
          )}
        </div>

        {status === 'idle' && store.wallet.chainDetails?.banner && (
          <p
            className={css({ color: 'foreground', mt: '4', fontSize: 'sm', '& a': { textDecoration: 'underline' } })}
            dangerouslySetInnerHTML={{ __html: store.wallet.chainDetails.banner.message }}
          />
        )}
      </div>
    </>
  )
})
