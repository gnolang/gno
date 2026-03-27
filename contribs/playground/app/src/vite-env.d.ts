/// <reference types="vite/client" />
/// <reference types="vite-plugin-svgr/client" />

interface ImportMetaEnv {
  readonly DEV: boolean
  readonly PROD: boolean
  readonly VITE_BACKEND_GIT_PROXY_URL: string
  readonly VITE_DOMAIN_URL: string
  readonly VITE_DISABLE_BEFORE_UNLOAD?: string
  readonly VITE_LSP_DEBUG: string
  readonly VITE_BACKEND_GIT_PROXY_URL?: string
  readonly VITE_GNOVM_WASM_VERSION: string
  readonly VITE_GNOVM_WASM_URL?: string
  readonly VITE_GNOVM_ROOT_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
