import { BackendInmemory, BackendMountableFileSystem, BackendOverlayFS, BackendZipFS } from '@gnostudio/pkg'

import { dirname } from 'path'
import * as BrowserFS from 'browserfs'
import type { FileSystem } from 'browserfs/dist/node/core/file_system'

import {
  createWasmFS,
  newMessagePortInputHandler,
  newMessagePortOutputHandler,
  newNoopInputHandler,
  parsePackageName,
  type StdioHandlers,
} from '../utils'
import { defaultSourcesDir, fetchGnoRootZip } from './resources'
import type { File } from './types'

const { Buffer } = BrowserFS.BFSRequire('buffer')

interface RootFsOpts {
  /**
   * Filesystem with source project files.
   */
  workspaceFs: FileSystem

  /**
   * Handlers for standard input and output calls.
   */
  stdio?: Partial<StdioHandlers>

  /**
   * Additional mount points
   */
  extraMountPoints?: Record<string, FileSystem>

  /**
   * Custom image of GNOROOT zip archive.
   *
   * If empty, default is provided.
   */
  gnoRootArchive?: ArrayBufferLike
}

/**
 * Creates and prepares root file system with Gno source files.
 */
export const createRootFs = async ({ workspaceFs, stdio, gnoRootArchive, extraMountPoints = {} }: RootFsOpts) => {
  gnoRootArchive ??= await fetchGnoRootZip()
  const fsSystem = await BackendInmemory()
  const fsGnoRoot = await BackendOverlayFS({
    readable: await BackendZipFS({ zipData: Buffer.from(gnoRootArchive) }),
    writable: await BackendInmemory(),
  })
  const fsGrouped = await BackendMountableFileSystem({
    '/gno': fsGnoRoot,
    '/tmp': fsSystem,
    [defaultSourcesDir]: workspaceFs,
    ...extraMountPoints,
  })

  const fs = BrowserFS.initialize(fsGrouped)
  const wasmFs = await createWasmFS(fs, stdio)
  return wasmFs
}

export const stdioFromPort = (port: MessagePort, withStdin: boolean): StdioHandlers => ({
  stdin: withStdin ? newMessagePortInputHandler(port) : newNoopInputHandler(),
  stdout: newMessagePortOutputHandler(port),
  stderr: newMessagePortOutputHandler(port, true),
})

type FileSet = Record<string, File>
type Packages = Record<
  string,
  {
    pkgName: string
    pkgPath: string
    files: FileSet
  }
>

export const findTestableDirs = (files?: FileSet, rootDir = defaultSourcesDir): string[] => {
  if (!files) {
    return []
  }

  const testableDirs = new Set<string>()

  Object.values(files).forEach((file) => {
    if (file.path.match(/_(filetest|test).gno$/i)) {
      const dir = file.path.substring(0, file.path.lastIndexOf('/'))
      testableDirs.add(rootDir + '/' + dir)
    }
  })

  return Array.from(testableDirs)
}

export const groupFilesByPackage = (files?: FileSet): Packages => {
  if (!files) {
    return {}
  }

  const packages: Packages = {}
  for (const file of Object.values(files)) {
    if (file.path.match(/_(filetest|test).gno$/i)) {
      continue
    }
    if (!file.path.endsWith('.gno')) {
      continue
    }
    const pkgName = parsePackageName(file.content)
    if (!pkgName) {
      // Skip files that don't contain a package definition
      continue
    }
    if (!packages[pkgName]) {
      const dir = dirname(file.path)
      packages[pkgName] = { pkgName, pkgPath: dir, files: { [file.path]: file } }
    } else {
      packages[pkgName].files[file.path] = file
    }
  }
  return packages
}

type CleanupFunction = () => void

export const wrapFetch = (): CleanupFunction => {
  const _fetch = self.globalThis.fetch

  // Wrapping the fetch function allows WebAssembly to succeed when calling it,
  // otherwise it raises the error:
  //   Failed to execute 'fetch' on 'WorkerGlobalScope': Illegal invocation
  //
  // TODO: Figure out why wrapping fetch works and fails otherwise (?)
  self.globalThis.fetch = async (resource: RequestInfo | URL, options?: RequestInit): Promise<Response> => {
    return await _fetch(resource, options)
  }

  return () => {
    // Restore the original fetch function
    self.globalThis.fetch = _fetch
  }
}

interface GnoGlobalArgs {
  rootDir: string
}

interface GnoRunArgs extends GnoGlobalArgs {
  files?: string[]
  expression?: string
}

interface GnoTestArgs extends GnoGlobalArgs {
  packages: string[]
  testPattern?: string
}

interface GnoReplArgs extends GnoGlobalArgs {
  /**
   * Initial code or commands to run (passed as --init)
   */
  init?: string
  /**
   * If true, do not print welcome line (passed as --skip-welcome)
   */
  skipWelcome?: boolean
}

/**
 * A helper to build command line arguments for Gno.
 */
export const cmdlineBuilder = {
  buildRunArgs: ({ rootDir, files, expression }: GnoRunArgs) => {
    const args = ['gno', 'run', '-v', '--root-dir', rootDir]
    if (expression?.length) {
      args.push('--expr', expression)
    }

    if (files?.length) {
      args.push(...files)
    }

    return args
  },
  buildTestArgs: ({ rootDir, packages, testPattern }: GnoTestArgs) => {
    const args = ['gno', 'test', ...packages, '-v', '--root-dir', rootDir]
    if (testPattern?.length) {
      args.push('--run', testPattern)
    }

    return args
  },
  buildReplArgs: ({ rootDir, init, skipWelcome }: GnoReplArgs) => {
    const args = ['gno', 'repl', '--root-dir', rootDir]
    if (init?.length) {
      args.push('--init', init)
    }
    if (skipWelcome) {
      args.push('--skip-welcome')
    }
    return args
  },
}
