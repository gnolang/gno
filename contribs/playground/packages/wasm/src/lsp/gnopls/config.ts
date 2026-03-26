/**
 * Gnopls WebAssembly file URL.
 *
 * Currently, WASM not deployed anywhere and path should be supplied manually for local dev.
 */
export const gnoplsUrl = import.meta.env.VITE_GNOPLS_URL

/**
 * Environment variables populated to LSP server.
 */
export const env = {
  GNOHOME: '/gnopath',
  GNOROOT: '/gnoroot',
}
