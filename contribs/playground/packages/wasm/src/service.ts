// Polifyll for WebAssembly.instantiateStreaming
if (!WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer()
    return await WebAssembly.instantiate(source, importObject)
  }
}

export class WasmService {
  private wasmSource: ArrayBuffer | null = null
  public wasmInstance: WebAssembly.Instance | null = null

  constructor(private readonly wasmUrl: string) {}

  public async fetchWasmModule() {
    if (this.wasmSource) {
      return
    }

    const rsp = await fetch(this.wasmUrl)
    if (!rsp.ok) {
      throw new Error(`Failed to fetch "${this.wasmUrl}": ${rsp.status} ${rsp.statusText}`)
    }

    this.wasmSource = await rsp.arrayBuffer()
  }

  public async load(importObject: WebAssembly.Imports) {
    await this.fetchWasmModule()

    const response = await WebAssembly.instantiate(this.wasmSource as ArrayBuffer, importObject)
    const instance = response.instance

    this.wasmInstance = instance

    return instance
  }

  public async getInstantce(importObject: WebAssembly.Imports) {
    if (!this.wasmInstance) {
      await this.load(importObject)
    }

    return this.wasmInstance
  }

  public call(method: string, args: any[]) {
    const instance = this.wasmInstance
    const fn = instance?.exports[method]

    // eslint-disable-next-line prefer-spread
    const result = (fn as any).apply(null, args)

    return result
  }

  public get instance() {
    return this.wasmInstance
  }
}
