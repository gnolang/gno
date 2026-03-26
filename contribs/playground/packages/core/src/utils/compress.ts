/**
 * Copied from Vue.js REPL
 * https://github.com/vuejs/repl/blob/main/src/utils.ts
 * MIT License
 */

import { strFromU8, strToU8, unzlibSync, zlibSync } from 'fflate'

function utoa(data: string): string {
  const buffer = strToU8(data)
  const zipped = zlibSync(buffer, { level: 9 })
  const binary = strFromU8(zipped, true)
  return btoa(binary)
}

function atou(base64: string): string {
  const binary = atob(base64)

  // zlib header (x78), level 9 (xDA)
  if (isCompressed(base64)) {
    const buffer = strToU8(binary, true)
    const unzipped = unzlibSync(buffer)
    return strFromU8(unzipped)
  }

  // old unicode hacks for backward compatibility
  // https://base64.guru/developers/javascript/examples/unicode-strings
  return decodeURIComponent(escape(binary))
}

function isCompressed(data: string): boolean {
  const binary = atob(data)
  return binary.startsWith('\x78\xDA')
}

export const compress = {
  utoa,
  atou,
  isCompressed,
}
