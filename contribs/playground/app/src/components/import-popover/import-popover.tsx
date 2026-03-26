import React, { useState } from 'react'
import { type FileRejection } from 'react-dropzone'
import { PiCopy, PiX } from 'react-icons/pi'
import { useNavigate } from 'react-router-dom'

import { buildPlaygroundGithubUrl, parseGithubUrl } from '@gnostudio/core/utils'
import { Popover, Portal, useWriteClipboard } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { css, cx } from '@/styled-system/css'
import { hstack, stack } from '@/styled-system/patterns'
import { button, input, link, popover } from '@/styled-system/recipes'

import { FileDropzone } from './file-dropzone'

interface ImportContentProps {
  onClose: () => void
}

type ImportMode = 'repository' | 'singleFile' | 'pasteUrl'

const ImportContent: React.FC<ImportContentProps> = observer(({ onClose }) => {
  const store = useStore()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)
  const [owner, setOwner] = useState('')
  const [repo, setRepo] = useState('')
  const [importMode, setImportMode] = useState<ImportMode>('pasteUrl')
  const [filePath, setFilePath] = useState('')
  const [branch, setBranch] = useState('main')
  const [pastedUrl, setPastedUrl] = useState('')
  const [copyLinkToClipboard, isLinkCopied] = useWriteClipboard()
  const githubTarget = parseGithubUrl(pastedUrl)
  const publicBase = import.meta.env.VITE_PUBLIC_PLAY_URL
  const importHref = githubTarget ? buildPlaygroundGithubUrl(publicBase, githubTarget) : ''
  const isInvalidUrl = pastedUrl.length > 0 && !githubTarget

  const popoverStyles = popover({ size: 'md' })
  const inputStyles = input({ size: 'lg' })

  const handleGithubImport = async (e: React.FormEvent) => {
    e.preventDefault()
    onClose()

    const searchParams = new URLSearchParams()

    if (importMode === 'singleFile') {
      searchParams.append('file', filePath)
      searchParams.append('branch', branch || 'main')
    }

    navigate({
      pathname: `/github/${owner}/${repo}`,
      search: searchParams.toString(),
    })
  }

  const handleFilesDropped = React.useCallback(
    async (acceptedFiles: File[], rejectedFiles: FileRejection[]) => {
      if (rejectedFiles.length > 0) {
        setError(
          `Only .gno and .toml files are supported. Failed to load: ${rejectedFiles
            .map((file) => file.file.name)
            .join(', ')}`,
        )
        return
      }

      if (acceptedFiles.length === 0) {
        setError('No valid files selected')
        return
      }

      try {
        onClose()

        await store.workbench.dropFiles(acceptedFiles)
      } catch (err) {
        console.error('Failed to import files:', String(err))
      }
    },
    [store, onClose],
  )

  const renderToggleButton = (mode: ImportMode, label: string) => (
    <button
      type="button"
      onClick={() => setImportMode(mode)}
      className={button({
        variant: importMode === mode ? 'solid' : 'outline',
        size: 'sm',
      })}
      aria-pressed={importMode === mode}
    >
      {label}
    </button>
  )

  return (
    <>
      <Popover.Title className={popoverStyles.title}>Import from GitHub</Popover.Title>
      <div className={stack({ gap: '4', mt: '4' })}>
        <p>Import gno code from a GitHub repository to experiment with it in Playground.</p>

        {error && <div className={css({ color: 'error', fontSize: 'sm' })}>{error}</div>}

        <div className={hstack({ gap: '2', mb: '3' })}>
          <label className={css({ fontWeight: 'medium', fontSize: 'sm', mr: '2' })}>Import Type:</label>
          {renderToggleButton('pasteUrl', 'Paste URL')}
          {renderToggleButton('repository', 'Repository')}
          {renderToggleButton('singleFile', 'Single File')}
        </div>

        {importMode !== 'pasteUrl' && (
          <form onSubmit={handleGithubImport}>
            <div className={stack({ gap: '4' })}>
              <div>
                <label className={css({ display: 'block', mb: '2', fontWeight: 'medium' })}>Repository Owner</label>
                <input
                  type="text"
                  value={owner}
                  onChange={(e) => setOwner(e.target.value)}
                  placeholder="e.g. gnolang"
                  className={cx(
                    inputStyles.root,
                    css({
                      width: 'full',
                      bg: 'bg.muted',
                      borderWidth: '1px',
                      borderColor: 'border.default',
                      _placeholder: { color: 'fg.subtle' },
                    }),
                  )}
                  required
                />
              </div>

              <div>
                <label className={css({ display: 'block', mb: '2', fontWeight: 'medium' })}>Repository Name</label>
                <input
                  type="text"
                  value={repo}
                  onChange={(e) => setRepo(e.target.value)}
                  placeholder="e.g. gno"
                  className={cx(
                    inputStyles.root,
                    css({
                      width: 'full',
                      bg: 'bg.muted',
                      borderWidth: '1px',
                      borderColor: 'border.default',
                      _placeholder: { color: 'fg.subtle' },
                    }),
                  )}
                  required
                />
              </div>

              {importMode === 'singleFile' && (
                <>
                  <div>
                    <label className={css({ display: 'block', mb: '2', fontWeight: 'medium' })}>
                      Path in Repository
                    </label>
                    <input
                      type="text"
                      value={filePath}
                      onChange={(e) => setFilePath(e.target.value)}
                      placeholder="e.g., examples/hello.gno"
                      className={cx(
                        inputStyles.root,
                        css({
                          width: 'full',
                          bg: 'bg.muted',
                          borderWidth: '1px',
                          borderColor: 'border.default',
                          _placeholder: { color: 'fg.subtle' },
                        }),
                      )}
                      required
                    />
                    <p className={css({ color: 'fg.subtle', fontSize: 'xs', mt: '1' })}>
                      Enter the full path to the file from the repository root (e.g., src/my_file.gno)
                    </p>
                  </div>
                  <div>
                    <label className={css({ display: 'block', mb: '2', fontWeight: 'medium' })}>
                      Branch (optional)
                    </label>
                    <input
                      type="text"
                      value={branch}
                      onChange={(e) => setBranch(e.target.value ?? 'main')}
                      placeholder="main"
                      className={cx(
                        inputStyles.root,
                        css({
                          width: 'full',
                          bg: 'bg.muted',
                          borderWidth: '1px',
                          borderColor: 'border.default',
                          _placeholder: { color: 'fg.subtle' },
                        }),
                      )}
                    />
                  </div>
                </>
              )}

              <div className={hstack({ justifyContent: 'flex-end', gap: '2' })}>
                <button
                  type="submit"
                  disabled={!owner || !repo || (importMode === 'singleFile' && !filePath)}
                  className={cx(button({ variant: 'solid' }), css({ width: 'auto' }))}
                >
                  Import
                </button>
              </div>
            </div>
          </form>
        )}

        {importMode === 'pasteUrl' && (
          <div className={stack({ gap: '3' })}>
            <label className={css({ display: 'block', mb: '1', fontWeight: 'medium' })}>GitHub URL</label>
            <input
              type="url"
              value={pastedUrl}
              onChange={(e) => setPastedUrl(e.target.value)}
              placeholder="https://github.com/{owner}/{repo} or blob URL"
              className={cx(inputStyles.root, css({ w: 'full' }))}
            />
            {isInvalidUrl && <span className={css({ color: 'error', fontSize: 'xs' })}>Invalid GitHub URL</span>}
            {githubTarget && (
              <div className={hstack({ justifyContent: 'flex-end', gap: '2' })}>
                <button
                  type="button"
                  onClick={() => copyLinkToClipboard(importHref)}
                  className={button({ variant: 'outline', size: 'sm' })}
                >
                  <PiCopy size={14} /> {isLinkCopied ? 'Copied!' : 'Copy link'}
                </button>
                <button
                  type="button"
                  className={button({ variant: 'solid', size: 'sm' })}
                  onClick={() => {
                    const sp = new URLSearchParams()
                    if (githubTarget.filePath) {
                      sp.append('file', githubTarget.filePath)
                      sp.append('branch', githubTarget.branch || 'main')
                    }
                    navigate({
                      pathname: `/github/${githubTarget.owner}/${githubTarget.repo}`,
                      search: sp.toString(),
                    })
                  }}
                >
                  Open in Playground
                </button>
              </div>
            )}
          </div>
        )}

        <div className={hstack({ alignItems: 'center', gap: '4' })}>
          <div className={css({ flex: 1, h: '1px', bg: 'border.default' })} />
          <span className={css({ color: 'fg.subtle', fontSize: 'sm' })}>OR</span>
          <div className={css({ flex: 1, h: '1px', bg: 'border.default' })} />
        </div>

        <FileDropzone onFilesDropped={handleFilesDropped} />
      </div>
    </>
  )
})

export const ImportPopover: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false)
  const popoverStyles = popover({ size: 'lg' })

  const handleOpenChange = (details: { open: boolean }) => {
    setIsOpen(details.open)
  }

  return (
    <Popover.Root
      open={isOpen}
      onOpenChange={handleOpenChange}
      modal
      autoFocus
      lazyMount
      positioning={{ offset: { mainAxis: 14 } }}
    >
      <Popover.Trigger asChild>
        <button className={link()} onClick={() => setIsOpen(true)}>
          Import
        </button>
      </Popover.Trigger>
      <Portal>
        <Popover.Positioner>
          <Popover.Content className={popoverStyles.content}>
            <Popover.Arrow className={popoverStyles.arrow}>
              <Popover.ArrowTip className={popoverStyles.arrowTip} />
            </Popover.Arrow>
            <Popover.CloseTrigger className={popoverStyles.close}>
              <PiX size={16} />
            </Popover.CloseTrigger>
            <ImportContent onClose={() => setIsOpen(false)} />
          </Popover.Content>
        </Popover.Positioner>
      </Portal>
    </Popover.Root>
  )
}
