/**
 * File defines a type for files.
 */
export interface File {
  path: string
  content: string
}

/**
 * GoFmtResult defines the type for Go formatting results.
 */
export interface GoFmtResult {
  out: string
  exitCode: number
}

export interface RunParams {
  /**
   * MessagePort to use for terminal input and output.
   *
   * Incoming messages are passed into stdin of a process.
   * Outcoming messages are produced by process stdout and stderr.
   */
  stdioPort: MessagePort

  /**
   * Current working directory to use when running the GnoVM.
   *
   * The default is to use the relative ".", which might not be right for all contexts.
   */
  workingDir?: string
}

export interface GnoProjectParams extends RunParams {
  files: Record<string, File>
}

export interface GnoEvalParams extends GnoProjectParams {
  expr?: string
}

/**
 * Interface for Comlink worker handler.
 */
export interface WorkerHandler {
  /**
   * Returns current used GnoVM version
   */
  getGnoVersion: () => string

  /**
   * Switch current GnoVM version
   */
  setGnoVersion: (version: string) => void

  /**
   * Returns list of available GnoVM versions.
   */
  getGnoVersions: () => string[]

  /**
   * Calls Gno binary to run all unit tests in a project.
   * @returns Process exit code.
   */
  gnotest: (params: GnoProjectParams) => Promise<number>

  /**
   * Calls Gno binary to open a REPL shell.
   * @returns Process exit code.
   */
  gnorepl: (params: GnoProjectParams) => Promise<number>

  /**
   * Calls Gno binary to evaluate a passed expression.
   * @returns Process exit code.
   */
  gnorun: (params: GnoEvalParams) => Promise<number>
  gofmt: (code: string) => Promise<GoFmtResult>
  gnoschema: (remote: string, realmPath: string) => Promise<string>
}
