import React, { useEffect, useMemo } from 'react'

import { getChainById, isStandardChain, type Chain } from '@gnostudio/core'

import { css } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'
import { link } from '@/styled-system/recipes'

type TxRenderFunc = ((hash: string) => string | undefined) | undefined

interface LinkParams {
  /**
   * Transaction hash. Used when URL is empty.
   */
  hash?: string

  /**
   * Deployed realm or package path.
   *
   * @example `gno.land/r/foo/bar`
   */
  packagePath: string

  /**
   * Whether passed package path is realm or package.
   */
  isRealm?: boolean

  /**
   * Active chain to build preview URL.
   *
   * Gno Studio link is displayed for dev, test3 and portal-loop networks.
   * Gnoscan is displayed for other chains.
   */
  chain?: Chain | null
}

interface Props extends LinkParams {
  onPreviewUrl?: (url: string | undefined) => void
}

/**
 * Constructs transaction URL build func for known chains.
 * Used as chain from store doesn't have links.
 *
 * @param chain Chain
 */
export const getTxRenderFunc = (chain?: Chain | null) => {
  if (!chain) {
    return
  }

  const chainDetails = getChainById(chain.id)
  return chainDetails?.link?.buildTxUrl.bind(chainDetails.link)
}

const buildLinkAndLabel = ({ hash, packagePath, chain, isRealm }: LinkParams, txRenderFunc: TxRenderFunc) => {
  const canViewOnChain = chain && isRealm && isStandardChain(chain.id) && chain.render
  if (canViewOnChain) {
    return { label: 'View on gno.land', url: chain.render!.buildUrl(packagePath) }
  }

  if (hash) {
    const url = txRenderFunc?.(hash)
    if (url) {
      return { label: 'Transaction link', url }
    }
  }

  return { label: 'Transaction hash' }
}

export const TxPreviewLink: React.FC<Props> = ({ onPreviewUrl, ...props }) => {
  const { chain, hash } = props
  const txRenderFunc = useMemo<TxRenderFunc>(() => getTxRenderFunc(chain), [chain])
  const { label, url } = buildLinkAndLabel(props, txRenderFunc)

  useEffect(() => {
    onPreviewUrl?.(url)
  }, [url, onPreviewUrl])

  return (
    <div className={stack()}>
      <p className={css({ fontWeight: 'semibold' })}>{label}</p>

      {url ? (
        <a href={url} target="_blank" className={link()} rel="noreferrer">
          {url}
        </a>
      ) : (
        <pre className={css({ fontFamily: 'monospace' })}>{hash}</pre>
      )}
    </div>
  )
}
