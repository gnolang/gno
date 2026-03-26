import { BUCKET_BASE_URL } from '../../go/resources'

/**
 * @see: https://github.com/gnostudio/studio-gopls-patches/releases
 */
const goplsVersion = 'v0.16.2-gnostudio3'
const bucketBaseUrl = `${BUCKET_BASE_URL}/gopls/${goplsVersion}`

const urlParentPath = (url: string) => {
  const u = new URL(url)
  const { pathname } = u

  const slashPos = pathname.lastIndexOf('/')
  u.pathname = slashPos === -1 ? pathname : pathname.slice(0, slashPos)
  return u.toString()
}

/**
 * Gopls WebAssembly file URL.
 *
 * Can be overriden using `VITE_GOPLS_URL` for local debugging.
 */
export const goplsUrl = import.meta.env.VITE_GOPLS_URL ?? `${bucketBaseUrl}/gopls.wasm`

/**
 * Path to Gno package stubs necessary to get gopls working.
 */
export const packageStubsUrl = `${urlParentPath(goplsUrl)}/stubs.zip`

export const GOPATH = '/gnopath'
export const GOROOT = '/gnoroot'

export const ignoredRequests = new Set(['textDocument/codeAction'])

/**
 * List of configuration environment variables for gnopls LSP server.
 */
export const lspConfig = {
  // Required for monaco, throws errors if disabled
  LSP_ENABLE_SEMANTIC_TOKENS: 'true',

  // List of inlay hints to enable.
  // See: https://github.com/golang/tools/blob/fa12f34b4218307705bf0365ab7df7c119b3653a/gopls/doc/inlayHints.md
  LSP_INLAY_HINT_OPTIONS: '',
  //  'assignVariableTypes,compositeLiteralFields,compositeLiteralTypes,constantValues,functionTypeParameters,parameterNames,rangeVariableTypes',

  // TODO: change when Gno doc will be available.
  // LSP_GODOC_DOMAIN: 'pkg.go.dev',

  // Enables LSP traces submission if not empty.
  LSP_TRACE_URL: import.meta.env.VITE_LSP_TRACE_URL ?? '',
}

export const env = {
  ...lspConfig,
  GOPATH,
  GOROOT,
  GOMODCACHE: `${GOPATH}/pkg/mod`,
  GOCACHE: '/tmp/go-build',
}
