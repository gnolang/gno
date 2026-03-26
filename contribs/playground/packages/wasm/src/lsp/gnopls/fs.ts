import { BackendInmemory, BackendMountableFileSystem, BackendZipFS, FolderAdapter } from '@gnostudio/pkg'

import * as BrowserFS from 'browserfs'
import { type FileSystem } from 'browserfs/dist/node/core/file_system'

import { fetchGnoRootZip } from '../../go/resources'
import type { WasmFS } from '../../utils/wasm-fs'
import { bfsToWasmFS, createWorkspaceFs } from '../common/fs'

interface WorkspaceConfig {
  gnoRoot: string
  gnoHome: string
  workspaceDir: string
}

/**
 * Fetches Gno SDK archive and returns GOROOT and GOPATH mountpoints.
 * @param cfg
 */
const createMountPoints = async ({ gnoHome, gnoRoot }: WorkspaceConfig): Promise<Record<string, FileSystem>> => {
  const sdkFiles = await fetchGnoRootZip()
  const gnoZipFs = await BackendZipFS({ zipData: Buffer.from(sdkFiles) })
  const goRootFs = new FolderAdapter('/', gnoZipFs)

  // TODO: cache dependencies in $GNOHOME/pkg
  return {
    [gnoRoot]: goRootFs,
    [gnoHome]: await BackendInmemory(),
  }
}

/**
 * Initializes and returns filesystem.
 * @param files Workspace files.
 * @param cfg Workspace configuration.
 */
export const createRootFs = async (filesPort: MessagePort, cfg: WorkspaceConfig): Promise<WasmFS> => {
  const workspaceFs = await createWorkspaceFs(filesPort)
  const gnoMountPoints = await createMountPoints(cfg)
  const mountPoints = {
    ...gnoMountPoints,
    [cfg.workspaceDir]: workspaceFs,
    '/tmp': await BackendInmemory(),
  }

  const bfs = BrowserFS.initialize(await BackendMountableFileSystem(mountPoints))
  const vfs = await bfsToWasmFS(bfs)

  return vfs
}
