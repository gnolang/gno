import { WasmService } from '../../service'
import { BaseServerController } from '../common/controller'
import type { LSPWorkerStartParams } from '../common/types'
import { env, gnoplsUrl } from './config'
import { createRootFs } from './fs'

/**
 * Server controller implementation for gnopls.
 */
export class ServerController extends BaseServerController {
  protected getEnvironmentVariables(): Record<string, string> {
    return {
      ...env,
      GNO_HOME: env.GNOHOME,
    }
  }

  static async create(params: LSPWorkerStartParams) {
    // Gnopls might misbehave if gno.mod file is missing.
    const fs = await createRootFs(params.ports.fs, {
      gnoRoot: env.GNOROOT,
      gnoHome: env.GNOHOME,
      workspaceDir: params.workspacePath,
    })
    const gnoplsService = new WasmService(gnoplsUrl)
    return new ServerController(fs, gnoplsService, params)
  }
}
