/// <reference types="vite/client" />
/// <reference types="vite-plugin-comlink/client" />

interface ImportMetaEnv {
  readonly VITE_GOPLS_URL?: string
  readonly VITE_LSP_TRACE_URL?: string
  readonly VITE_GNOVM_WASM_VERSION: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
