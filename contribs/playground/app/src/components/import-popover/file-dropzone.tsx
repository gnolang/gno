import React from 'react'
import { useDropzone, type FileRejection } from 'react-dropzone'
import { PiFile } from 'react-icons/pi'

import { css } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'

interface FileDropzoneProps {
  onFilesDropped: (acceptedFiles: File[], rejectedFiles: FileRejection[]) => void
}

export const FileDropzone: React.FC<FileDropzoneProps> = ({ onFilesDropped }) => {
  const { getRootProps, getInputProps } = useDropzone({
    accept: {
      'text/plain': ['.gno', '.toml'],
    },
    onDrop: onFilesDropped,
    noClick: true,
  })

  return (
    <div>
      <h3 className={css({ fontSize: 'md', fontWeight: 'medium', mb: '2' })}>Upload Local Files</h3>
      <div
        {...getRootProps()}
        className={stack({
          p: 12,
          border: '2px dashed',
          borderColor: 'border',
          borderRadius: 'md',
          align: 'center',
          textAlign: 'center',
          minH: 32,
          cursor: 'pointer',
        })}
      >
        <input {...getInputProps()} />
        <div className={stack({ align: 'center', gap: 2 })}>
          <PiFile size={32} color="gray" />
          <div>
            <p>
              <strong>Drag & drop</strong>
            </p>
            <p>Drag & drop .gno or .toml files here, or click to select files</p>
          </div>
        </div>
      </div>
    </div>
  )
}
