import React, { useState } from 'react'
import { PiX } from 'react-icons/pi'

import { Tabs } from '@gnostudio/react'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { type FileType } from '@/store/workbench'
import { css, cx } from '@/styled-system/css'
import { center } from '@/styled-system/patterns'
import { tabs } from '@/styled-system/recipes'

import { FileInputName } from './file-input-name'

interface CloseButtonProps {
  file: FileType
  disabled?: boolean
}

const CloseButton: React.FC<CloseButtonProps> = observer(({ file, disabled }) => {
  const store = useStore()
  const showUnsavedIndicator = file.hasUnsavedChanges && store.projects.hasActive

  const handleDelete = (evt: React.MouseEvent<HTMLElement>) => {
    evt.stopPropagation()
    if (disabled) {
      return
    }

    store.workbench.deleteFile(file.path)
  }

  return (
    <span
      role="button"
      data-action="delete"
      tabIndex={-1}
      title="Delete file"
      className={center({
        ml: 2,
        w: 5,
        h: 5,
        _hover: { bg: 'primary' },
      })}
      onClick={handleDelete}
    >
      {showUnsavedIndicator ? (
        <span
          className={css({
            display: 'inline-block',
            w: 2,
            h: 2,
            bg: 'foreground',
            borderRadius: '50%',
            verticalAlign: 'middle',
            p: 1,
          })}
        />
      ) : (
        <PiX />
      )}
    </span>
  )
})

type Props = {
  file: FileType
  index: number
} & Omit<React.ComponentProps<typeof Tabs.Trigger>, 'value'>

export const FileTab: React.FC<Props> = observer(({ file, index, disabled, ...props }) => {
  const store = useStore()
  const [isDragging, setIsDragging] = useState(false)

  const handleOnDrop = (ev: React.DragEvent<HTMLButtonElement>) => {
    ev.preventDefault()
    ev.currentTarget.removeAttribute('data-drop-target')
    const path = ev.dataTransfer.getData('text/plain')
    store.workbench.reorderFile(path, index)
  }

  const handleOnDragStart = (ev: React.DragEvent<HTMLButtonElement>) => {
    setIsDragging(true)
    ev.dataTransfer.setData('text/plain', file.path)
    ev.dataTransfer.effectAllowed = 'move'
  }

  const handleOnDragEnd = () => {
    setIsDragging(false)
  }

  const handleOnDragOver = (ev: React.DragEvent<HTMLButtonElement>) => {
    ev.preventDefault()
    ev.dataTransfer.dropEffect = 'move'
    if (isDragging) return
    ev.currentTarget.setAttribute('data-drop-target', 'true')
  }

  const handleOnDragLeave = (ev: React.DragEvent<HTMLButtonElement>) => {
    ev.currentTarget.removeAttribute('data-drop-target')
  }

  return (
    <Tabs.Trigger
      {...props}
      value={file.path}
      draggable
      onDrop={handleOnDrop}
      onDragStart={handleOnDragStart}
      onDragOver={handleOnDragOver}
      onDragLeave={handleOnDragLeave}
      onDragEnd={handleOnDragEnd}
      onDoubleClick={() => store.workbench.startRenameFile(file.path)}
      disabled={disabled}
      className={cx(
        tabs().trigger,
        css({
          position: 'relative',
          '&[data-drop-target=true]': {
            _after: {
              content: '""',
              position: 'absolute',
              w: 'full',
              h: 'full',
              top: 0,
              left: 0,
              bg: 'background',
              opacity: 0.5,
            },
          },
        }),
      )}
    >
      <div className={css({ pos: 'relative', overflow: 'hidden' })}>
        <span className={css({ pointerEvents: 'none' })}>{file.path}</span>
        {store.workbench.renamingPath === file.path && (
          <div className={css({ pos: 'absolute', top: '0', bg: 'background', w: 'full' })}>
            <FileInputName
              fullWidth
              onSubmit={() => store.workbench.renameFile(file.path, store.workbench.pendingFileName as string)}
              onCancel={() => store.workbench.cancelNameFile()}
              onBlur={() => store.workbench.cancelNameFile()}
              isDisabled={disabled}
            />
          </div>
        )}
      </div>
      <CloseButton file={file} disabled={disabled} />
    </Tabs.Trigger>
  )
})
