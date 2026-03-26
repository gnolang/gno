import React from 'react'
import { useDropzone, type FileRejection } from 'react-dropzone'

import { useStore } from '@/contexts'
import { css } from '@/styled-system/css'

interface UploadDropzoneProps {
  children: React.ReactNode
}

export const UploadDropzone: React.FC<UploadDropzoneProps> = ({ children }) => {
  const store = useStore()

  const onDrop = React.useCallback(
    (acceptedFiles: File[], rejectedFiles: FileRejection[]) => {
      if (rejectedFiles.length > 0) {
        const errorMessage = [
          'Only .gno and .toml files are supported. Failed to load:',
          rejectedFiles.map((file) => file.file.name).join(', '),
        ].join(' ')
        console.error(errorMessage)
      }

      if (acceptedFiles.length === 0) return

      store.workbench.dropFiles(acceptedFiles)
    },
    [store],
  )

  const { getRootProps, isDragActive } = useDropzone({
    accept: {
      'text/plain': ['.gno', '.toml'],
    },
    noClick: true,
    onDrop,
  })

  return (
    <div
      data-drop-target={isDragActive}
      className={css({
        w: 'full',
        h: 'full',
        pos: 'relative',
      })}
      {...getRootProps()}
    >
      {children}

      {isDragActive && (
        <div
          className={css({
            w: 'full',
            h: 'full',
            pos: 'absolute',
            top: 0,
            left: 0,
            bg: 'gray.600',
            opacity: 0.5,
          })}
        />
      )}
    </div>
  )
}
