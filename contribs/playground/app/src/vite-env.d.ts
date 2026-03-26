/// <reference types="vite/client" />
/// <reference types="vite-plugin-svgr/client" />

interface ImportMetaEnv {
  readonly DEV: boolean
  readonly PROD: boolean
  readonly VITE_BACKEND_GIT_PROXY_URL: string
  readonly VITE_BACKEND_GRPC_URL: string
  readonly VITE_DOMAIN_URL: string
  readonly VITE_DISABLE_BEFORE_UNLOAD?: string
  readonly VITE_LSP_DEBUG: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
