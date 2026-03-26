import { BackendInmemory, KeyValueFileSystem } from '@gnostudio/pkg'

import { join } from 'path'
import type { FileSystem } from 'browserfs/dist/node/core/file_system'
import * as Comlink from 'comlink'

import { WasmService } from '../service'
import {
  createWasmFS,
  isGnoFile,
  newInputFromData,
  newStreamOutputHandler,
  withGnoWorkFile,
  type WasmFS,
} from '../utils'
import { execWithStdout } from './exec'
import { gnoVMWasmStore } from './gnovm'
import { BUCKET_BASE_URL, defaultGnoRoot, defaultSourcesDir, fetchGnoRootZip } from './resources'
import type { GnoEvalParams, GoFmtResult, WorkerHandler } from './types'
import { cmdlineBuilder, createRootFs, findTestableDirs, groupFilesByPackage, stdioFromPort, wrapFetch } from './utils'
import { Go } from './wasm-exec'

const gofmtWasmUrl = import.meta.env.VITE_GOFMT_WASM_URL ?? `${BUCKET_BASE_URL}/gofmt.wasm`
const gnoschemaWasmUrl = import.meta.env.VITE_GNOSCHEMA_URL ?? `${BUCKET_BASE_URL}/gnoschema.wasm`

const gofmtWasmService = new WasmService(gofmtWasmUrl)
const gnoschemaWasmService = new WasmService(gnoschemaWasmUrl)

export async function preload() {
  await gnoVMWasmStore.refreshVersions()
  await gnoVMWasmStore.fetchWasmModule()
  await gofmtWasmService.fetchWasmModule()
  await fetchGnoRootZip()
}

/**
 * Runs a Gno command and returns exit code.
 * @param args Command-line arguments.
 * @param goFs Virtual file system instance.
 * @param workingDir Current working directory (default: .).
 */
const runGno = async (args: string[], goFs: WasmFS, workingDir: string | undefined): Promise<number> => {
  const go = new Go()
  go.argv = args
  // TODO: provide gnohome/.gno
  go.vars.fs = goFs
  go.env = {
    HOME: defaultSourcesDir,
    GNOROOT: defaultGnoRoot,
    CWD: join(defaultSourcesDir, workingDir ?? '.'),
  }

  return await gnoVMWasmStore
    .load(go.importObject)
    .then((instance) => go.run(instance))
    .then(() => go.getExitCode())
}

const replWorkspaceDir = '/gno/gnovm/stdlibs/gnostudio'

/**
 * Gno & Go web worker implementation.
 */
export const worker: WorkerHandler = {
  getGnoVersion: () => gnoVMWasmStore.version,
  getGnoVersions: () => gnoVMWasmStore.getVersions(),
  setGnoVersion: (v: string) => gnoVMWasmStore.setVersion(v),
  gofmt: async (code: string): Promise<GoFmtResult> => {
    const stdio = {
      stdin: newInputFromData(code),
      stdout: newStreamOutputHandler(),
      stderr: newStreamOutputHandler(),
    }

    const fs = await createWasmFS(null, stdio)

    const { stdout: out, exitCode } = await execWithStdout(gofmtWasmService, { fs, stdio })
    return { out, exitCode }
  },
  gnorepl: async ({ stdioPort, files, workingDir }) => {
    // Group Gno files by package
    const packages = groupFilesByPackage(files)

    // In REPL mode, user packages are searched as stdlibs under '/gno/gno/gnovm/stdlibs/gnostudio/*'.
    // They're imported as such because they don't start with the "gno.land/" prefix when imported in
    // the REPL, they are directly imported by their package name.
    const mounts: Record<string, FileSystem> = {}
    const imports: string[] = []
    Object.entries(packages).forEach(([pkgName, data]) => {
      // Mount package files
      const pkgFs = KeyValueFileSystem.fromObject(data.files, { stripRootSlash: true })
      const mountDir = join(replWorkspaceDir, pkgName)
      const importDir = join('gnostudio', data.pkgPath, pkgName)
      mounts[mountDir] = pkgFs

      // Use "gnostudio/" as prefix for all user defined packages so they are loaded from "gnostudio/" dir
      imports.push(`import "${importDir}"`)
    })

    const wasmFs = await createRootFs({
      workspaceFs: await BackendInmemory(),
      stdio: stdioFromPort(stdioPort, true),
      extraMountPoints: mounts,
    })

    const args = cmdlineBuilder.buildReplArgs({
      rootDir: defaultGnoRoot,
      init: imports.join(';'), // Auto import user defined packages on REPL init
      skipWelcome: true,
    })

    return await runGno(args, wasmFs, workingDir)
  },
  gnotest: async ({ files, stdioPort, workingDir }) => {
    files = withGnoWorkFile(files)
    const testableDirs = findTestableDirs(files)
    const args = cmdlineBuilder.buildTestArgs({
      rootDir: defaultGnoRoot,
      packages: testableDirs,
    })

    const workspaceFs = KeyValueFileSystem.fromObject(files, { stripRootSlash: true })
    const wasmFs = await createRootFs({
      workspaceFs,
      stdio: stdioFromPort(stdioPort, false),
    })

    return await runGno(args, wasmFs, workingDir)
  },
  gnorun: async ({ files, expr, stdioPort, workingDir }: GnoEvalParams) => {
    const rootDir = defaultSourcesDir
    const paths = Object.keys(files)
      .filter(isGnoFile)
      .map((path) => `${rootDir}/${path}`)

    const workspaceFs = KeyValueFileSystem.fromObject(files, { stripRootSlash: true })
    const wasmFs = await createRootFs({
      workspaceFs,
      stdio: stdioFromPort(stdioPort, false),
    })
    const cmdline = cmdlineBuilder.buildRunArgs({
      rootDir: defaultGnoRoot,
      expression: expr,
      files: paths,
    })
    return await runGno(cmdline, wasmFs, workingDir)
  },
  gnoschema: async (remote: string, realmPath: string) => {
    const stdio = {
      stdout: newStreamOutputHandler(),
      stderr: newStreamOutputHandler(),
    }

    const fs = await createWasmFS(null, stdio)
    const restoreFetch = wrapFetch()

    try {
      const { stdout } = await execWithStdout(gnoschemaWasmService, {
        fs,
        stdio,
        cmdline: ['gnoschema', '-enable-render-proxy', '-remote-types', '-remote', remote, realmPath],
      })

      return stdout
    } finally {
      restoreFetch()
    }
  },
}

Comlink.expose(worker)
