import { BackendWorkerFS } from '@gnostudio/pkg'

import { type FileSystem } from 'browserfs/dist/node/core/file_system'

import { createWasmFS, newLogOutputWriter, newNoopInputHandler, type WasmFS } from '../../utils/wasm-fs'

/**
 * Returns WasmFS instance from root BFS filesystem object.
 */
export const bfsToWasmFS = async (fs: FileSystem): Promise<WasmFS> => {
  const stdio = {
    stdin: newNoopInputHandler(),
    stdout: newLogOutputWriter(),
    stderr: newLogOutputWriter(),
  }

  const { unlink: fsUnlink, ...originalFs } = await createWasmFS(fs, stdio)

  // CHANGE: It seems that a temporary go mod hash file file is not created for some reason.
  // For example:
  //   /tmp/go.903023c377b3aba4cb9f80be8ed7fffbe782d032347b0dae35c60512d27bc68c.3669936616.sum
  //
  // Override unlink to avoid the issue for now.
  // TODO: is this is still necessary?
  const unlink = (path: string, callback: any) => {
    try {
      fsUnlink(path, callback)
    } catch (e: any) {
      if (e.code !== 'ENOENT') {
        throw e
      }
    }
  }

  return {
    ...originalFs,
    unlink,
  }
}

export const createWorkspaceFs = async (filesPort: MessagePort) => {
  // BFS uses `addEventListener` which won't work unless `start()` is called.
  // See: https://issues.chromium.org/issues/40429353
  filesPort.start()

  const fs = await BackendWorkerFS({
    worker: filesPort,
  })
  return fs
}
