import { makeAutoObservable } from 'mobx'

import gnoWasmUrl from '@gnoide/tools/data/wasm/gno.wasm?url'
import { WasmService } from '../service'

const DEFAULT_VERSION = import.meta.env.VITE_GNOVM_WASM_VERSION ?? 'bundled'

class GnoVMWasmStore {
  private currentVersion = DEFAULT_VERSION

  private services: Record<string, WasmService> = {}

  state: 'idle' | 'fetching' = 'idle'

  constructor() {
    makeAutoObservable(this)
  }

  get version() {
    return this.currentVersion
  }

  get versions(): string[] {
    return [this.currentVersion]
  }

  getVersions() {
    return JSON.parse(JSON.stringify(this.versions))
  }

  setVersion(version: string) {
    this.currentVersion = version
  }

  async load(importObject: WebAssembly.Imports) {
    if (!this.services[this.version]) {
      await this.fetchWasmModule(this.version)
    }

    const service = this.services[this.version]
    return await service.load(importObject)
  }

  async fetchWasmModule(version?: string) {
    const newVersion = version ?? this.version

    this.state = 'fetching'

    const url = this.getWasmUrl(newVersion)
    const service = new WasmService(url)

    await service.fetchWasmModule()

    this.state = 'idle'

    this.services[newVersion] = service
  }

  async refreshVersions() {
    // No-op: single bundled version
  }

  private getWasmUrl(_version: string) {
    return import.meta.env.VITE_GNOVM_WASM_URL ?? gnoWasmUrl
  }
}

export const gnoVMWasmStore = new GnoVMWasmStore()
