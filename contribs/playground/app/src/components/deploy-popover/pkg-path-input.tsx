import React from 'react'
import { PiQuestionFill } from 'react-icons/pi'

import { packagePaths } from '@gnostudio/core'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'
import { input } from '@/styled-system/recipes'

interface PkgPathInputProps {
  onSubmit?: () => void
}

const isValidPackageName = (name: string): boolean => {
  // Allow paths for sub-packages by splitting on '/' and validating each part
  const parts = name.split('/').filter((part) => part.length > 0) // Remove empty parts from trailing/leading slashes

  // Must have at least one valid part
  if (parts.length === 0) return false

  return parts.every((part) => {
    // Each part must be a valid package name (lowercase letters, numbers, hyphens, underscores)
    return /^[a-z0-9][a-z0-9_-]*[a-z0-9]$|^[a-z0-9]$/.test(part)
  })
}

const NamespaceSelect: React.FC<React.SelectHTMLAttributes<HTMLSelectElement>> = observer(() => {
  const inputStyles = input({ size: 'md' })
  const store = useStore()

  return (
    <div className={css({ pos: 'relative', w: 'full' })}>
      <select
        className={cx(inputStyles.root, css({ w: 'full', pl: '7' }))}
        value={store.deployer.pathNamespace}
        onChange={(e) => store.deployer.setPathNamespace(e.target.value)}
      >
        <option disabled>Namespace</option>
        <option value={store.wallet.account?.address}>{store.wallet.account?.address}</option>
        {store.deployer.userNamespace ? (
          <option value={store.deployer.userNamespace}>{store.deployer.userNamespace}</option>
        ) : null}
      </select>

      <a
        title="Learn more about namespaces"
        className={css({
          pos: 'absolute',
          top: '50%',
          transform: 'translateY(-50%)',
          left: '2',
          cursor: 'pointer',
          color: 'foreground.muted',
          _hover: { color: 'foreground' },
        })}
        href="https://docs.gno.land/concepts/namespaces/"
        target="_blank"
        rel="noreferrer"
      >
        <PiQuestionFill size={16} />
      </a>
    </div>
  )
})

export const PkgPathInput: React.FC<PkgPathInputProps> = observer(({ onSubmit }) => {
  const inputStyles = input({ size: 'md' })
  const store = useStore()

  const onKeyDown = ({ key }: React.KeyboardEvent) => {
    if (key === 'Enter' && store.deployer.hasValidPath) {
      onSubmit?.()
    }
  }

  const packageNameValid = !store.deployer.pathPart || isValidPackageName(store.deployer.pathPart)

  return (
    <div className={stack({ gap: '3' })}>
      <div className={stack({ gap: '2' })}>
        <label className={css({ fontWeight: 'medium', fontSize: 'sm', color: 'foreground' })}>1. Package Type</label>
        <select
          className={cx(inputStyles.root)}
          value={store.deployer.pathType}
          onChange={(e) => store.deployer.setPathType(e.target.value)}
        >
          <option value={packagePaths.packageNamespace}>Package ({packagePaths.packageNamespace})</option>
          <option value={packagePaths.realmNamespace}>Realm ({packagePaths.realmNamespace})</option>
        </select>
      </div>

      <div className={stack({ gap: '2' })}>
        <label className={css({ fontWeight: 'medium', fontSize: 'sm', color: 'foreground' })}>2. Namespace</label>
        {store.deployer.canModifyNamespace ? (
          <input
            className={cx(inputStyles.root)}
            value={store.deployer.pathNamespace}
            onChange={(e) => store.deployer.setPathNamespace(e.target.value)}
            onKeyDown={onKeyDown}
            placeholder="Your namespace or wallet address"
          />
        ) : (
          <NamespaceSelect />
        )}
        <span className={css({ fontSize: 'xs', color: 'foreground.muted' })}>
          Use your wallet address or custom namespace
        </span>
      </div>

      <div className={stack({ gap: '2' })}>
        <label className={css({ fontWeight: 'medium', fontSize: 'sm', color: 'foreground' })}>3. Package Name</label>
        <input
          className={cx(
            inputStyles.root,
            css({
              borderColor: store.deployer.pathPart && !packageNameValid ? 'red.500' : undefined,
            }),
          )}
          value={store.deployer.pathPart}
          onChange={(e) => store.deployer.setPathPart(e.target.value)}
          onKeyDown={onKeyDown}
          placeholder="my-awesome-package or sub/package"
        />
        <div className={css({ fontSize: 'xs' })}>
          {store.deployer.pathPart && !packageNameValid ? (
            <span className={css({ color: 'red.500' })}>
              Name must contain only lowercase letters, numbers, hyphens, underscores, and forward slashes for
              sub-packages
            </span>
          ) : (
            <span className={css({ color: 'foreground.muted' })}>
              Lowercase letters, numbers, hyphens, underscores, and forward slashes for sub-packages
            </span>
          )}
        </div>
      </div>

      {store.deployer.pathPart && (
        <div className={stack({ gap: '2' })}>
          <label className={css({ fontWeight: 'medium', fontSize: 'sm', color: 'foreground' })}>Final Path</label>
          <div
            className={css({
              p: '3',
              bg: 'header',
              border: '1px solid',
              borderColor: 'border',
              borderRadius: 'md',
              fontSize: 'sm',
              fontFamily: 'mono',
              color: 'foreground',
              wordBreak: 'break-all',
            })}
          >
            {store.deployer.pathPrefix}
            {store.deployer.pathType}/{store.deployer.pathNamespace || 'namespace'}/{store.deployer.pathPart || 'name'}
          </div>
        </div>
      )}
    </div>
  )
})
