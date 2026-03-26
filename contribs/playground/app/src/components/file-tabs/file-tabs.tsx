import React, { useRef, type ChangeEvent } from 'react'
import { AiOutlinePlus, AiOutlineUpload } from 'react-icons/ai'

import { Tabs } from '@gnostudio/react'
import { isGnoFile } from '@gnostudio/wasm'

import { observer } from 'mobx-react-lite'

import { useStore } from '@/contexts'
import { type InvalidFileNameException } from '@/lib'
import { css, cx } from '@/styled-system/css'
import { spacer } from '@/styled-system/patterns'
import { tabs } from '@/styled-system/recipes'

import { FileInputName } from './file-input-name'
import { FileTab } from './file-tab'

interface Props {
  rightElement?: React.ReactNode
  showActions?: boolean
  disabled?: boolean
}

export const FileTabs: React.FC<Props> = observer(({ rightElement, disabled, showActions = true }) => {
  const tabsStyle = tabs()
  const store = useStore()

  const renameInputRef = useRef<HTMLInputElement>(null)
  const uploadInputRef = useRef<HTMLInputElement>(null)

  const handleNew = () => {
    store.workbench.startAddFile()
  }

  const onFileSelected = async ({ target }: ChangeEvent<HTMLInputElement>) => {
    try {
      // macOS native file picker doesn't care about file extensions in "accept", even in Chrome.
      const gnoFiles = Array.prototype.filter.call(
        target.files,
        (file: File) => isGnoFile(file.name) && !store.workbench.files.has(file.name),
      )
      await store.workbench.dropFiles(gnoFiles)
    } finally {
      target.value = ''
    }
  }

  const cancelRename = () => {
    store.workbench.cancelNameFile()
  }

  const completeRename = () => {
    try {
      store.workbench.addFile()
    } catch (error) {
      alert((error as InvalidFileNameException).message)
      setTimeout(() => renameInputRef.current?.focus(), 0)
    }
  }

  const handleRenameOnBlur = () => {
    try {
      store.workbench.addFile()
    } catch {
      cancelRename()
    }
  }

  return (
    <Tabs.Root
      className={tabsStyle.root}
      value={store.workbench.activePath}
      onValueChange={(evt) => store.workbench.setActivePath(evt.value)}
    >
      <input
        type="file"
        ref={uploadInputRef}
        accept="*.gno,*.toml"
        style={{ display: 'none' }}
        onChange={onFileSelected}
        multiple
      />
      <Tabs.List className={tabsStyle.list}>
        {store.workbench.tabs.map((file, index) => (
          <FileTab key={file.path} file={file} index={index} disabled={disabled} />
        ))}

        {store.workbench.isCreatingFile && (
          <div className={tabsStyle.trigger}>
            <FileInputName
              ref={renameInputRef}
              onSubmit={completeRename}
              onCancel={cancelRename}
              onBlur={handleRenameOnBlur}
              isDisabled={disabled}
            />
          </div>
        )}

        {showActions && (
          <>
            <button
              title="New file"
              className={cx(css({ paw: 'Click+Add+New+File' }), tabsStyle.trigger)}
              onClick={handleNew}
              disabled={disabled}
            >
              <AiOutlinePlus />
            </button>

            <button
              title="Upload files"
              className={cx(css({ paw: 'Click+Upload+File' }), tabsStyle.trigger)}
              onClick={() => uploadInputRef.current?.click()}
              disabled={disabled}
            >
              <AiOutlineUpload size={18} />
            </button>
          </>
        )}
        <div className={spacer()} />
        {rightElement}
      </Tabs.List>
    </Tabs.Root>
  )
})
