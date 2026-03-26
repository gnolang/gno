import { reader } from '@gnostudio/pkg'

import type { WasmService } from '../service'
import type { StreamOutputHandler, WasmFS } from '../utils/wasm-fs'
import { Go } from './wasm-exec'

interface ExecOpts {
  fs: WasmFS
  cmdline?: string[]
  stdio: {
    stdout: StreamOutputHandler
    stderr: StreamOutputHandler
  }
}

/**
 * Executes a WASM module and returns the stdout.
 *
 * Throws an error if stderr is not empty or non-zero exit code.
 *
 * @param mod The WASM module to execute
 * @param opts Execution options.
 */
export const execWithStdout = async (mod: WasmService, { fs, cmdline, stdio }: ExecOpts) => {
  const go = new Go()
  go.vars.fs = fs
  if (cmdline) {
    go.argv = cmdline
  }

  await mod.load(go.importObject).then((instance) => go.run(instance))

  const stdout = await reader(stdio.stdout.stream).readAll()
  const stderr = await reader(stdio.stderr.stream).readAll()

  if (stderr.length > 0) {
    throw new Error(stderr)
  }

  const exitCode = go.getExitCode()
  if (exitCode !== 0) {
    const progName = cmdline?.[0] ?? 'program'
    throw new Error(`${progName} returned non-zero exit code: ${exitCode}`)
  }

  return { stdout, exitCode }
}
