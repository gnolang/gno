import gnozip from '@gnoide/tools/data/gno/root.zip?url'

export const BUCKET_BASE_URL = 'https://storage.googleapis.com/tendermintproduct.appspot.com'

export const defaultGnoRoot = '/gno'
export const defaultSourcesDir = '/src'

let gnoRootZip: ArrayBuffer | null = null

/**
 * Fetches Gno SDK and returns it as ArrayBuffer of Zip archive.
 *
 * If the SDK is already fetched, returns it from cache.
 */
export const fetchGnoRootZip = async (): Promise<ArrayBuffer> => {
  if (gnoRootZip) {
    return gnoRootZip
  }

  const baseUrl = new URL(import.meta.url).origin
  const zipUrl = new URL(gnozip, baseUrl)

  const rsp = await fetch(zipUrl)
  if (!rsp.ok) {
    throw new Error(`Failed to fetch Gno SDK: ${rsp.status} ${rsp.statusText}`)
  }

  gnoRootZip = await rsp.arrayBuffer()
  return gnoRootZip
}
