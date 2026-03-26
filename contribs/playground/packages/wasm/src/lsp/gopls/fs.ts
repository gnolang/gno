import { BackendInmemory, BackendMountableFileSystem, BackendZipFS, FolderAdapter } from '@gnostudio/pkg'

import * as BrowserFS from 'browserfs'
import { FileFlag } from 'browserfs/dist/node/core/file_flag'
import { type FileSystem } from 'browserfs/dist/node/core/file_system'

import { fetchGnoRootZip } from '../../go/resources'
import type { WasmFS } from '../../utils/wasm-fs'
import { bfsToWasmFS, createWorkspaceFs } from '../common/fs'
import { packageStubsUrl } from './config'

interface WorkspaceConfig {
  GOROOT: string
  GOPATH: string
  workspaceDir: string
}

/**
 * Returns list of mounts for Gno stub packages which will be mounted on top of GNOROOT.
 *
 * Stub packages contain declarations for missing or injectable symbols
 * and necessary for make gopls module resolution work.
 */
const createPackageStubsMounts = async (pkgRootDir: string) => {
  const rsp = await fetch(packageStubsUrl)
  if (!rsp.ok) {
    throw new Error(`Failed to fetch Gno stub packages: ${rsp.status} ${rsp.statusText}`)
  }

  const zipfs = await BackendZipFS({
    zipData: Buffer.from(await rsp.arrayBuffer()),
  })

  const stubsList: string[] = zipfs
    .readFileSync('/stubs.txt', 'utf-8', new FileFlag('r'))
    .split('\n')
    .map((s: string) => s.trim())
    .filter((s: string) => s.length && !s.startsWith('#'))

  const overlays: Record<string, FolderAdapter> = Object.fromEntries(
    stubsList.map((pkgName) => [`${pkgRootDir}/${pkgName}`, new FolderAdapter(`/${pkgName}`, zipfs)]),
  )

  return overlays
}

/**
 * Fetches Gno SDK archive and returns GOROOT and GOPATH mountpoints.
 * @param cfg
 */
const createMountPoints = async ({ GOROOT, GOPATH }: WorkspaceConfig): Promise<Record<string, FileSystem>> => {
  // The Gno SDK archive contains separate entries which should be put in GOROOT and GOPATH.
  //
  // 'gnovm/stdlibs' go into GOROOT and 'examples' into GOPATH.
  const sdkFiles = await fetchGnoRootZip()
  const gnoZipFs = await BackendZipFS({ zipData: Buffer.from(sdkFiles) })

  const goRootFs = new FolderAdapter('/gnovm/stdlibs', gnoZipFs)
  const gnoPackagesFs = new FolderAdapter('/examples', gnoZipFs)

  const pkgRootDir = `${GOROOT}/src`
  const gnoStubsMounts = await createPackageStubsMounts(pkgRootDir)

  return {
    [pkgRootDir]: goRootFs,
    ...gnoStubsMounts,
    [`${GOPATH}/src`]: gnoPackagesFs,
    [`${GOPATH}/pkg`]: await BackendInmemory(),
    [`${GOPATH}/cache`]: await BackendInmemory(),
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
