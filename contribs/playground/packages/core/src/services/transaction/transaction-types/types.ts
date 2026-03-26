// Should be in sync with https://github.com/gnolang/gno/blob/master/gnovm/gno.proto

export interface MemFile {
  name: string
  body: string
}

/**
 * Represents a package with set of files.
 *
 * @see https://github.com/gnolang/gno/blob/84e53f51b6988528196159d824351885cdeb57b8/gnovm/gno.proto#L606
 */
export interface MemPackage {
  name: string
  path: string
  files: MemFile[]
}
